package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ainsuotain/sshtie/internal/profile"
)

var copyCmd = &cobra.Command{
	Use:     "copy <source> <dest>",
	Aliases: []string{"cp"},
	Short:   "Copy a profile with a new name",
	Long: `Duplicate an existing profile under a new name.

All settings (host, user, port, key, tmux session, SSH options) are copied.
Edit the new profile with:  sshtie edit <dest>

Example:
  sshtie copy work-server work-server-backup
  sshtie cp   homelab     homelab-test`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		src, dst := args[0], args[1]

		p, err := profile.Get(src)
		if err != nil {
			return err
		}

		p.Name = dst
		if err := profile.Add(p); err != nil {
			return err
		}

		fmt.Printf("✅ Profile '%s' copied → '%s'\n", src, dst)
		fmt.Printf("   %s\n", p.Host)
		syncSSHConfig()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(copyCmd)
}
