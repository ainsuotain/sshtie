package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ainsuotain/sshtie/internal/connector"
	"github.com/ainsuotain/sshtie/internal/profile"
	"github.com/ainsuotain/sshtie/internal/tui"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor <name>",
	Short: "Diagnose connectivity for a profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := profile.Get(args[0])
		if err != nil {
			return err
		}

		result, err := tui.RunDoctor(p)
		if err != nil {
			return err
		}

		// Terminal is fully restored here.
		switch result.Action {
		case tui.DoctorInstall:
			if err := runInstall(p); err != nil {
				return err
			}
			fmt.Printf("\n→ Connecting to %s (%s@%s)…\n", p.Name, p.User, p.Host)
			return connector.Connect(p)

		case tui.DoctorConnect:
			fmt.Printf("→ Connecting to %s (%s@%s)…\n", p.Name, p.User, p.Host)
			return connector.Connect(p)

		default: // DoctorQuit / DoctorNone
			return nil
		}
	},
}
