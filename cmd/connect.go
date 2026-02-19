package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/ainsuotain/sshtie/internal/connector"
	"github.com/ainsuotain/sshtie/internal/profile"
	"github.com/ainsuotain/sshtie/internal/tui"
)

var connectCmd = &cobra.Command{
	Use:   "connect <name>",
	Short: "Connect to a profile (mosh → ssh fallback → tmux)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := profile.Get(args[0])
		if err != nil {
			return err
		}
		return runConnect(p)
	},
}

// runConnect shows the connection-progress TUI then executes the chosen action.
// Shared by connectCmd, root shortcut, and the profile-picker TUI.
func runConnect(p profile.Profile) error {
	port := p.Port
	if port == 0 {
		port = 22
	}

	result, err := tui.RunConnect(p, !isKnownHost(p.Host, port))
	if err != nil {
		return err
	}

	// Terminal is fully restored here — safe to exec ssh/mosh.
	switch result.Action {
	case tui.ConnectInstall:
		if err := runInstall(p); err != nil {
			return err
		}
		fmt.Printf("\n→ Connecting to %s (%s@%s)…\n", p.Name, p.User, p.Host)
		return connector.Connect(p)

	case tui.ConnectProceed:
		fmt.Printf("→ Connecting to %s (%s@%s)…\n", p.Name, p.User, p.Host)
		return connector.Connect(p)

	default: // ConnectQuit / ConnectNone
		return nil
	}
}

// isKnownHost reports whether the host is already in ~/.ssh/known_hosts.
// Returns true (assume known) if ssh-keygen is unavailable.
func isKnownHost(host string, port int) bool {
	target := host
	if port != 0 && port != 22 {
		target = fmt.Sprintf("[%s]:%d", host, port)
	}
	err := exec.Command("ssh-keygen", "-F", target).Run()
	if err == nil {
		return true
	}
	if _, ok := err.(*exec.ExitError); ok {
		return false
	}
	return true // ssh-keygen unavailable — assume known
}
