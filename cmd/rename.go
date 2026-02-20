package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ainsuotain/sshtie/internal/profile"
)

var renameCmd = &cobra.Command{
	Use:     "rename <name>",
	Aliases: []string{"mv"},
	Short:   "Rename a profile",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		oldName := args[0]

		// Verify it exists.
		if _, err := profile.Get(oldName); err != nil {
			return err
		}

		fmt.Printf("Rename '%s' → new name: ", oldName)
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		newName := strings.TrimSpace(input)

		if newName == "" || newName == oldName {
			fmt.Println("→ Cancelled.")
			return nil
		}

		if err := profile.Rename(oldName, newName); err != nil {
			return err
		}
		fmt.Printf("✅ Renamed '%s' → '%s'\n", oldName, newName)
		syncSSHConfig()
		return nil
	},
}
