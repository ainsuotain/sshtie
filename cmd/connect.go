package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ainsuotain/sshtie/internal/connector"
	"github.com/ainsuotain/sshtie/internal/profile"
)

var connectCmd = &cobra.Command{
	Use:   "connect <name>",
	Short: "Connect to a profile (mosh → ssh fallback → tmux)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		p, err := profile.Get(name)
		if err != nil {
			return err
		}
		fmt.Printf("→ Connecting to %s (%s@%s)…\n", p.Name, p.User, p.Host)
		return connector.Connect(p)
	},
}
