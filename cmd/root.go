package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ainsuotain/sshtie/internal/connector"
	"github.com/ainsuotain/sshtie/internal/doctor"
	"github.com/ainsuotain/sshtie/internal/profile"
	"github.com/ainsuotain/sshtie/internal/tui"
)

var rootCmd = &cobra.Command{
	Use:   "sshtie [profile-name]",
	Short: "SSH + mosh + tmux in one command",
	Long: `sshtie manages SSH/mosh/tmux profiles and picks the best
connection strategy automatically.

  sshtie connect <name>   Connect to a profile
  sshtie add              Add a new profile
  sshtie list             List all profiles
  sshtie doctor <name>    Diagnose connection
  sshtie remove <name>    Remove a profile
  sshtie install <name>   Install mosh + tmux on a remote server`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// No profile name given — launch interactive TUI.
			profiles, err := profile.Load()
			if err != nil {
				return err
			}
			result, err := tui.Run(profiles)
			if err != nil {
				return err
			}
			// Terminal is now fully restored; safe to exec ssh/mosh.
			switch result.Action {
			case tui.ActionConnect:
				fmt.Printf("→ Connecting to %s (%s@%s)…\n", result.Profile.Name, result.Profile.User, result.Profile.Host)
				return connector.Connect(result.Profile)
			case tui.ActionDoctor:
				doctor.Run(result.Profile)
				return nil
			default:
				// ActionQuit or ActionNone — nothing to do.
				return nil
			}
		}
		// `sshtie <name>` shortcut → connect directly.
		name := args[0]
		p, err := profile.Get(name)
		if err != nil {
			return err
		}
		fmt.Printf("→ Connecting to %s (%s@%s)…\n", p.Name, p.User, p.Host)
		return connector.Connect(p)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(connectCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(installCmd)
}
