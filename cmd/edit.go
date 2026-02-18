package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/ainsuotain/sshtie/internal/profile"
)

var editCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Edit a profile in $EDITOR (fallback: nano)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Verify profile exists and load it.
		p, err := profile.Get(name)
		if err != nil {
			return err
		}

		// Write the profile to a temp file for editing.
		tmp, err := os.CreateTemp("", "sshtie-*.yaml")
		if err != nil {
			return fmt.Errorf("create temp file: %w", err)
		}
		tmpPath := tmp.Name()
		defer os.Remove(tmpPath)

		data, err := yaml.Marshal(p)
		if err != nil {
			return fmt.Errorf("marshal profile: %w", err)
		}
		if _, err := tmp.Write(data); err != nil {
			return fmt.Errorf("write temp file: %w", err)
		}
		tmp.Close()

		// Determine editor.
		editor := os.Getenv("EDITOR")
		if editor == "" {
			if runtime.GOOS == "windows" {
				editor = "notepad"
			} else {
				editor = "nano"
			}
		}

		// Open editor (blocking until user closes it).
		c := exec.Command(editor, tmpPath)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return fmt.Errorf("editor exited with error: %w", err)
		}

		// Parse the edited file.
		editedData, err := os.ReadFile(tmpPath)
		if err != nil {
			return fmt.Errorf("read edited file: %w", err)
		}
		var updated profile.Profile
		if err := yaml.Unmarshal(editedData, &updated); err != nil {
			return fmt.Errorf("parse edited profile: %w", err)
		}

		// Always keep the original name (prevent accidental rename).
		updated.Name = name

		// Merge back into profiles.yaml.
		profiles, err := profile.Load()
		if err != nil {
			return err
		}
		found := false
		for i, pr := range profiles {
			if pr.Name == name {
				profiles[i] = updated
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("profile %q not found", name)
		}
		if err := profile.Save(profiles); err != nil {
			return err
		}

		fmt.Printf("✅ Profile '%s' updated.\n", name)
		return nil
	},
}
