package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ainsuotain/sshtie/internal/profile"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		profiles, err := profile.Load()
		if err != nil {
			return err
		}
		if len(profiles) == 0 {
			fmt.Println("No profiles yet. Run: sshtie add")
			return nil
		}

		fmt.Println()
		fmt.Printf("  %-16s %-24s %-16s %-6s  %s\n",
			"NAME", "HOST", "USER", "PORT", "TAGS")
		fmt.Println("  " + strings.Repeat("─", 72))

		for _, p := range profiles {
			port := p.Port
			if port == 0 {
				port = 22
			}
			tags := strings.Join(p.Tags, ", ")
			if tags == "" {
				tags = "—"
			}
			fmt.Printf("  %-16s %-24s %-16s %-6d  %s\n",
				p.Name, p.Host, p.User, port, tags)
		}
		fmt.Println()
		return nil
	},
}
