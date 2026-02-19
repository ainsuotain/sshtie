package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ainsuotain/sshtie/internal/profile"
	"github.com/ainsuotain/sshtie/internal/tui"
)

var editCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Edit advanced SSH options for a profile (interactive UI)",
	Long: `Edit advanced SSH connection options for an existing profile.

Use the slider UI to adjust:
  · ConnectionAttempts  — how many times to retry before giving up
  · ServerAliveInterval — keepalive ping frequency
  · ServerAliveCountMax — how many missed pings before disconnect
  · ForwardAgent        — SSH agent forwarding (for bastion hosts)

Controls:
  ↑/↓         select option
  ←/→         adjust value by 1
  shift+←/→   jump by larger step
  enter        save
  esc / q      cancel`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
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

		// Apply changes.
		p.ForwardAgent        = result.ForwardAgent
		p.ConnectionAttempts  = result.ConnectionAttempts
		p.ServerAliveInterval = result.ServerAliveInterval
		p.ServerAliveCountMax = result.ServerAliveCountMax

		// Persist updated profile.
		profiles, err := profile.Load()
		if err != nil {
			return err
		}
		for i, existing := range profiles {
			if existing.Name == name {
				profiles[i] = p
				break
			}
		}
		if err := profile.Save(profiles); err != nil {
			return err
		}

		fmt.Printf("✅ Profile '%s' updated!\n", name)
		fmt.Printf("   %s\n", tui.SummaryLine(p))
		return nil
	},
}
