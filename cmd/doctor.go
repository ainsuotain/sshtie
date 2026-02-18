package cmd

import (
	"github.com/spf13/cobra"

	"github.com/ainsuotain/sshtie/internal/doctor"
	"github.com/ainsuotain/sshtie/internal/profile"
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
		doctor.Run(p)
		return nil
	},
}
