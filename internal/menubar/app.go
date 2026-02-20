//go:build darwin || windows

// Package menubar implements the macOS / Windows system-tray application.
package menubar

import (
	"fmt"
	"os"
	"time"

	"fyne.io/systray"

	"github.com/ainsuotain/sshtie/internal/checker"
	"github.com/ainsuotain/sshtie/internal/profile"
	"github.com/ainsuotain/sshtie/internal/session"
)

// intervalPresets cycles: 10s → 30s → 60s → 10s …
var intervalPresets = []int{10, 30, 60}

// stopMenuCh is closed to stop all goroutines from the previous buildMenu call.
var stopMenuCh chan struct{}

// Run starts the menu-bar / system-tray app. It blocks until Quit is clicked.
func Run() {
	systray.Run(onReady, nil)
}

// ── lifecycle ──────────────────────────────────────────────────────────────────

func onReady() {
	icon := iconBytes()
	systray.SetTemplateIcon(icon, icon)
	systray.SetTooltip("sshtie — SSH profile manager")

	chk := checker.New()

	// rebuildCh triggers a menu rebuild from any goroutine.
	rebuildCh := make(chan struct{}, 1)
	trigger := func() {
		select {
		case rebuildCh <- struct{}{}:
		default:
		}
	}

	// Load profiles and run the first check immediately.
	profiles, _ := profile.Load()
	buildMenu(profiles, chk, trigger)
	go chk.CheckAll(profiles, trigger)
	go chk.RefreshSessions(trigger)

	// Background loop: TCP check every 60 s, sessions every 5 s.
	go func() {
		tcpTicker  := time.NewTicker(60 * time.Second)
		sessTicker := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-tcpTicker.C:
				profiles, _ = profile.Load()
				go chk.CheckAll(profiles, trigger)
			case <-sessTicker.C:
				go chk.RefreshSessions(trigger)
			case <-rebuildCh:
				profiles, _ = profile.Load()
				buildMenu(profiles, chk, trigger)
			}
		}
	}()
}

// ── menu builder ───────────────────────────────────────────────────────────────

func buildMenu(profiles []profile.Profile, chk *checker.Checker, trigger func()) {
	// Stop all goroutines from the previous buildMenu call.
	if stopMenuCh != nil {
		close(stopMenuCh)
	}
	stopMenuCh = make(chan struct{})
	stop := stopMenuCh // captured by goroutines below

	systray.ResetMenu()

	// ── header ──
	hdr := systray.AddMenuItem("sshtie", "SSH profile manager")
	hdr.Disable()
	systray.AddSeparator()

	// ── profile items ──
	if len(profiles) == 0 {
		empty := systray.AddMenuItem("No profiles yet — Add Server…", "")
		go func() {
			for {
				select {
				case <-stop:
					return
				case _, ok := <-empty.ClickedCh:
					if !ok {
						return
					}
					OpenAdd()
				}
			}
		}()
	} else {
		for _, p := range profiles {
			addProfileMenu(p, chk, trigger, stop)
		}
	}

	systray.AddSeparator()

	// ── global actions ──
	addItem     := systray.AddMenuItem("Add Server…", "Open terminal to add a new profile")
	refreshItem := systray.AddMenuItem("Refresh Status", "Re-check all servers now")

	systray.AddSeparator()

	loginLabel := "  Open at Login"
	if IsAutoStartEnabled() {
		loginLabel = "✓ Open at Login"
	}
	loginItem := systray.AddMenuItem(loginLabel, "Start sshtie automatically when you log in")

	systray.AddSeparator()
	quitItem := systray.AddMenuItem("Quit sshtie", "")

	go func() {
		for {
			select {
			case <-stop:
				return
			case _, ok := <-addItem.ClickedCh:
				if !ok {
					return
				}
				OpenAdd()
			case _, ok := <-refreshItem.ClickedCh:
				if !ok {
					return
				}
				ps, _ := profile.Load()
				go chk.CheckAll(ps, trigger)
				go chk.RefreshSessions(trigger)
			case _, ok := <-loginItem.ClickedCh:
				if !ok {
					return
				}
				if IsAutoStartEnabled() {
					_ = DisableAutoStart()
					loginItem.SetTitle("  Open at Login")
				} else {
					_ = EnableAutoStart()
					loginItem.SetTitle("✓ Open at Login")
				}
			case _, ok := <-quitItem.ClickedCh:
				if !ok {
					return
				}
				systray.Quit()
			}
		}
	}()
}

// addProfileMenu creates one top-level profile item with its sub-menu.
func addProfileMenu(p profile.Profile, chk *checker.Checker, trigger func(), stop chan struct{}) {
	reachable, known    := chk.Get(p.Name)
	activeSess, isActive := chk.ActiveSession(p.Name)

	label   := menuLabel(p.Name, reachable, known, isActive)
	tooltip := fmt.Sprintf("%s@%s · port %d", p.User, p.Host, portOf(p))
	if isActive {
		tooltip += fmt.Sprintf(" · connected via %s", activeSess.Method)
	}

	item := systray.AddMenuItem(label, tooltip)

	// ── sub-items ──
	connectItem  := item.AddSubMenuItem("Connect", "Open a terminal and connect")
	sep1         := item.AddSubMenuItem("──────────", "")
	sep1.Disable()
	intervalItem := item.AddSubMenuItem(intervalLabel(p), "Cycle: 10s → 30s → 60s")
	faItem        := item.AddSubMenuItem(faLabel(p), "Toggle SSH agent forwarding")
	editItem      := item.AddSubMenuItem("Edit SSH Options…", "Open slider UI to adjust all options")

	var disconnectCh <-chan struct{}
	if isActive {
		sep2 := item.AddSubMenuItem("──────────", "")
		sep2.Disable()
		disconnectCh = item.AddSubMenuItem("Disconnect", "Terminate this connection").ClickedCh
	} else {
		disconnectCh = make(chan struct{}) // never fires
	}

	pCopy      := p
	activeCopy := activeSess

	go func() {
		for {
			select {
			case <-stop:
				return
			case _, ok := <-connectItem.ClickedCh:
				if !ok {
					return
				}
				OpenConnect(pCopy.Name)
			case _, ok := <-intervalItem.ClickedCh:
				if !ok {
					return
				}
				updateProfile(pCopy.Name, func(pr *profile.Profile) {
					pr.ServerAliveInterval = nextInterval(pr.ServerAliveInterval)
				})
				trigger()
			case _, ok := <-faItem.ClickedCh:
				if !ok {
					return
				}
				updateProfile(pCopy.Name, func(pr *profile.Profile) {
					pr.ForwardAgent = !pr.ForwardAgent
				})
				trigger()
			case _, ok := <-editItem.ClickedCh:
				if !ok {
					return
				}
				OpenEdit(pCopy.Name)
			case _, ok := <-disconnectCh:
				if !ok {
					return
				}
				killSession(activeCopy)
				trigger()
			}
		}
	}()
}

// ── helpers ───────────────────────────────────────────────────────────────────

func intervalLabel(p profile.Profile) string {
	v := p.ServerAliveInterval
	if v <= 0 {
		v = 10
	}
	return fmt.Sprintf("Interval: %ds", v)
}

func faLabel(p profile.Profile) string {
	if p.ForwardAgent {
		return "Forward agent: on"
	}
	return "Forward agent: off"
}

// nextInterval returns the smallest preset strictly greater than current.
// Wraps back to the first preset when current is at or beyond the max.
func nextInterval(current int) int {
	for _, v := range intervalPresets {
		if v > current {
			return v
		}
	}
	return intervalPresets[0]
}

// updateProfile loads profiles.yaml, applies fn to the named profile, and saves.
func updateProfile(name string, fn func(*profile.Profile)) {
	profiles, err := profile.Load()
	if err != nil {
		return
	}
	for i := range profiles {
		if profiles[i].Name == name {
			fn(&profiles[i])
			break
		}
	}
	_ = profile.Save(profiles)
}

// killSession terminates the process recorded in the session file.
func killSession(s session.Session) {
	if s.PID <= 0 {
		return
	}
	p, err := os.FindProcess(s.PID)
	if err != nil {
		return
	}
	_ = p.Kill()
	_ = session.Delete(s.Profile)
}

func menuLabel(name string, reachable, known, active bool) string {
	prefix := statusDot(reachable, known)
	if active {
		return prefix + "● " + name
	}
	return prefix + name
}

func statusDot(reachable, known bool) string {
	switch {
	case !known:
		return "🟡  "
	case reachable:
		return "🟢  "
	default:
		return "🔴  "
	}
}

func portOf(p profile.Profile) int {
	if p.Port == 0 {
		return 22
	}
	return p.Port
}
