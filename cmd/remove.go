package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ainsuotain/sshtie/internal/profile"
)

var removeCmd = &cobra.Command{
	Use:     "remove <name>",
	Aliases: []string{"rm", "delete"},
	Short:   "Remove a profile",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := profile.Remove(name); err != nil {
			return err
		}
		fmt.Printf("✅ Profile '%s' removed.\n", name)
		return nil
	},
}
