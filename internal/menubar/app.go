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

// Run starts the macOS menu-bar app. It blocks until the user clicks Quit.
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

	// Background loop: re-check every 60 s, sessions every 5 s.
	go func() {
		tcpTicker := time.NewTicker(60 * time.Second)
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

// ── menu builder ──────────────────────────────────────────────────────────────

func buildMenu(profiles []profile.Profile, chk *checker.Checker, trigger func()) {
	systray.ResetMenu()

	// ── header (non-clickable) ──
	hdr := systray.AddMenuItem("sshtie", "SSH profile manager")
	hdr.Disable()
	systray.AddSeparator()

	// ── profile items ──
	if len(profiles) == 0 {
		empty := systray.AddMenuItem("No profiles yet — Add Server…", "")
		go func() {
			for range empty.ClickedCh {
				OpenAdd()
			}
		}()
	} else {
		for _, p := range profiles {
			reachable, known := chk.Get(p.Name)
			activeSess, isActive := chk.ActiveSession(p.Name)
			label := menuLabel(p.Name, reachable, known, isActive)
			tooltip := fmt.Sprintf("%s@%s · port %d", p.User, p.Host, portOf(p))
			if isActive {
				tooltip += fmt.Sprintf(" · connected via %s", activeSess.Method)
			}
			item := systray.AddMenuItem(label, tooltip)

			pCopy := p
			if isActive {
				// Show sub-item to disconnect.
				disconnectItem := item.AddSubMenuItem("Disconnect", "Terminate this connection")
				go func(s session.Session) {
					for range disconnectItem.ClickedCh {
						killSession(s)
						trigger()
					}
				}(activeSess)

				go func() {
					for range item.ClickedCh {
						OpenConnect(pCopy.Name)
					}
				}()
			} else {
				go func() {
					for range item.ClickedCh {
						OpenConnect(pCopy.Name)
					}
				}()
			}
		}
	}

	systray.AddSeparator()

	// ── actions ──
	addItem := systray.AddMenuItem("Add Server…", "Open terminal to add a new profile")
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
			case <-addItem.ClickedCh:
				OpenAdd()
			case <-refreshItem.ClickedCh:
				ps, _ := profile.Load()
				go chk.CheckAll(ps, trigger)
				go chk.RefreshSessions(trigger)
			case <-loginItem.ClickedCh:
				if IsAutoStartEnabled() {
					_ = DisableAutoStart()
					loginItem.SetTitle("  Open at Login")
				} else {
					_ = EnableAutoStart()
					loginItem.SetTitle("✓ Open at Login")
				}
			case <-quitItem.ClickedCh:
				systray.Quit()
			}
		}
	}()
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
