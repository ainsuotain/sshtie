package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ainsuotain/sshtie/internal/connector"
	"github.com/ainsuotain/sshtie/internal/profile"
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
	// Running `sshtie <name>` shortcuts to connect.
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// No TUI yet — just print help.
			return cmd.Help()
		}
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
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(installCmd)
}
