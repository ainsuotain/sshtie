package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ainsuotain/sshtie/internal/profile"
)

const (
	sshtieBegin = "# BEGIN sshtie managed — do not edit this block manually"
	sshtieEnd   = "# END sshtie managed"
)

var sshConfigCmd = &cobra.Command{
	Use:   "ssh-config",
	Short: "Sync all profiles to ~/.ssh/config (for Cursor, VS Code, etc.)",
	Long: `Write all sshtie profiles as Host entries into ~/.ssh/config.

Existing sshtie-managed entries are replaced on each run.
Your own SSH config entries outside the managed block are untouched.

After running, Cursor / VS Code Remote-SSH will show the profiles
in their server picker automatically.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		profiles, err := profile.Load()
		if err != nil {
			return err
		}
		if len(profiles) == 0 {
			fmt.Println("No profiles found. Run: sshtie add")
			return nil
		}

		sshConfigPath, err := sshConfigPath()
		if err != nil {
			return err
		}

		// Read existing config (ok if file doesn't exist yet).
		existing := ""
		if data, err := os.ReadFile(sshConfigPath); err == nil {
			existing = string(data)
		}

		// Strip previous sshtie block.
		existing = stripBlock(existing)

		// Build the new managed block.
		var block strings.Builder
		block.WriteString("\n" + sshtieBegin + "\n")
		for _, p := range profiles {
			block.WriteString(buildHostEntry(p))
		}
		block.WriteString(sshtieEnd + "\n")

		// Append managed block to the existing config.
		output := strings.TrimRight(existing, "\n") + block.String()

		// Ensure ~/.ssh directory exists.
		if err := os.MkdirAll(filepath.Dir(sshConfigPath), 0700); err != nil {
			return fmt.Errorf("create ~/.ssh: %w", err)
		}
		if err := os.WriteFile(sshConfigPath, []byte(output), 0600); err != nil {
			return fmt.Errorf("write ssh config: %w", err)
		}

		fmt.Printf("✅ %d profile(s) written to %s\n\n", len(profiles), sshConfigPath)
		for _, p := range profiles {
			port := p.Port
			if port == 0 {
				port = 22
			}
			fmt.Printf("   Host %-20s → %s@%s:%d\n", p.Name, p.User, p.Host, port)
		}
		fmt.Println("\n→ Restart Cursor / VS Code to refresh the SSH target list.")
		return nil
	},
}

// buildHostEntry generates an SSH config Host block for a profile.
func buildHostEntry(p profile.Profile) string {
	var sb strings.Builder

	port := p.Port
	if port == 0 {
		port = 22
	}

	fmt.Fprintf(&sb, "\nHost %s\n", p.Name)
	fmt.Fprintf(&sb, "  HostName %s\n", p.Host)
	fmt.Fprintf(&sb, "  User %s\n", p.User)
	if port != 22 {
		fmt.Fprintf(&sb, "  Port %d\n", port)
	}

	// Key file — only add if explicitly set (don't write the default).
	if p.Key != "" {
		fmt.Fprintf(&sb, "  IdentityFile %s\n", p.Key)
	}

	// Advanced SSH options — only emit non-default values.
	if p.ForwardAgent {
		sb.WriteString("  ForwardAgent yes\n")
	}

	aliveInterval := p.ServerAliveInterval
	if aliveInterval <= 0 {
		aliveInterval = 10
	}
	aliveCount := p.ServerAliveCountMax
	if aliveCount <= 0 {
		aliveCount = 60
	}
	attempts := p.ConnectionAttempts
	if attempts <= 0 {
		attempts = 3
	}
	fmt.Fprintf(&sb, "  ServerAliveInterval %d\n", aliveInterval)
	fmt.Fprintf(&sb, "  ServerAliveCountMax %d\n", aliveCount)
	fmt.Fprintf(&sb, "  ConnectionAttempts %d\n", attempts)

	return sb.String()
}

// stripBlock removes the sshtie-managed block from a config string.
func stripBlock(config string) string {
	lines := strings.Split(config, "\n")
	var out []string
	inside := false
	for _, line := range lines {
		if strings.TrimSpace(line) == sshtieBegin {
			inside = true
			continue
		}
		if strings.TrimSpace(line) == sshtieEnd {
			inside = false
			continue
		}
		if !inside {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}

// sshConfigPath returns the path to ~/.ssh/config.
func sshConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ssh", "config"), nil
}
