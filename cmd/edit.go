package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ainsuotain/sshtie/internal/profile"
	"github.com/ainsuotain/sshtie/internal/tui"
)

// runEdit opens the interactive editor for the named profile.
// Shared by editCmd and the main TUI 'e' shortcut.
func runEdit(name string) error {
	p, err := profile.Get(name)
	if err != nil {
		return err
	}

	fmt.Printf("Current: %s\n\n", tui.SummaryLine(p))

	result, err := tui.RunEdit(p)
	if err != nil {
		return err
	}
	if !result.Saved {
		fmt.Println("→ Cancelled. No changes saved.")
		return nil
	}

	if result.Deleted {
		if err := profile.Remove(name); err != nil {
			return err
		}
		fmt.Printf("✅ Profile '%s' removed.\n", name)
		syncSSHConfig()
		return nil
	}

	profiles, err := profile.Load()
	if err != nil {
		return err
	}

	finalName := name
	if result.NewName != "" {
		for _, existing := range profiles {
			if existing.Name == result.NewName {
				return fmt.Errorf("profile %q already exists", result.NewName)
			}
		}
		finalName = result.NewName
	}

	for i, existing := range profiles {
		if existing.Name != name {
			continue
		}
		profiles[i].Name                = finalName
		profiles[i].ForwardAgent        = result.ForwardAgent
		profiles[i].ConnectionAttempts  = result.ConnectionAttempts
		profiles[i].ServerAliveInterval = result.ServerAliveInterval
		profiles[i].ServerAliveCountMax = result.ServerAliveCountMax
		break
	}

	if err := profile.Save(profiles); err != nil {
		return err
	}

	if finalName != name {
		fmt.Printf("✅ Renamed '%s' → '%s'\n", name, finalName)
	}
	updated, _ := profile.Get(finalName)
	fmt.Printf("✅ Profile '%s' updated!\n", finalName)
	fmt.Printf("   %s\n", tui.SummaryLine(updated))
	syncSSHConfig()
	return nil
}

var editCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Edit SSH options, rename, or delete a profile (interactive UI)",
	Long: `Edit an existing SSH connection profile via an interactive slider UI.

Controls:
  ↑/↓         select option
  ←/→         adjust value by 1
  shift+←/→   jump by larger step
  enter        save
  esc / q      cancel

SSH options:
  · ConnectionAttempts  — how many times to retry before giving up
  · ServerAliveInterval — keepalive ping frequency
  · ServerAliveCountMax — how many missed pings before disconnect
  · ForwardAgent        — SSH agent forwarding (for bastion hosts)

Profile management (bottom section):
  · Rename  — type a new name and press enter
  · Delete  — press enter twice to confirm deletion`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runEdit(args[0])
	},
}
