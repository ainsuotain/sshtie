//go:build darwin

// Package menubar implements the macOS status-bar application.
package menubar

import (
	"fmt"
	"time"

	"fyne.io/systray"

	"github.com/ainsuotain/sshtie/internal/checker"
	"github.com/ainsuotain/sshtie/internal/profile"
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

	// Background loop: re-check every 60 s, rebuild on change.
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		for {
			select {
			case <-ticker.C:
				profiles, _ = profile.Load() // pick up newly added servers
				go chk.CheckAll(profiles, trigger)
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
			label := menuLabel(p.Name, reachable, known)
			tooltip := fmt.Sprintf("%s@%s · port %d", p.User, p.Host, portOf(p))
			item := systray.AddMenuItem(label, tooltip)

			pCopy := p
			go func() {
				for range item.ClickedCh {
					OpenConnect(pCopy.Name)
				}
			}()
		}
	}

	systray.AddSeparator()

	// ── actions ──
	addItem := systray.AddMenuItem("Add Server…", "Open terminal to add a new profile")
	refreshItem := systray.AddMenuItem("Refresh Status", "Re-check all servers now")
	systray.AddSeparator()
	quitItem := systray.AddMenuItem("Quit sshtie", "")

	go func() {
		for {
			select {
			case <-addItem.ClickedCh:
				OpenAdd()
			case <-refreshItem.ClickedCh:
				profiles, _ := profile.Load()
				go chk.CheckAll(profiles, trigger)
			case <-quitItem.ClickedCh:
				systray.Quit()
			}
		}
	}()
}

func menuLabel(name string, reachable, known bool) string {
	switch {
	case !known:
		return "🟡  " + name // still checking
	case reachable:
		return "🟢  " + name
	default:
		return "🔴  " + name
	}
}

func portOf(p profile.Profile) int {
	if p.Port == 0 {
		return 22
	}
	return p.Port
}
