package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ainsuotain/sshtie/internal/profile"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new profile (interactive)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAdd()
	},
}

func runAdd() error {
	r := bufio.NewReader(os.Stdin)

	fmt.Println()

	// Required fields
	name := prompt(r, "Profile name", "", true)
	host := prompt(r, "Host", "", true)
	user := prompt(r, "User", "", true)

	// Optional fields with defaults
	portStr := prompt(r, "Port", "22", false)
	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 {
		port = 22
	}

	key := prompt(r, "SSH Key", "~/.ssh/id_ed25519", false)
	if key == "~/.ssh/id_ed25519" {
		key = "" // store empty to use default
	}

	tmuxSession := prompt(r, "tmux session", "main", false)
	if tmuxSession == "main" {
		tmuxSession = "main"
	}

	network := prompt(r, "Network mode (auto/tailscale/direct)", "auto", false)
	if network != "tailscale" && network != "direct" {
		network = "auto"
	}

	p := profile.Profile{
		Name:        name,
		Host:        host,
		User:        user,
		Port:        port,
		Key:         key,
		TmuxSession: tmuxSession,
		Network:     network,
	}

	if err := profile.Add(p); err != nil {
		return err
	}

	fmt.Printf("\n✅ Profile '%s' saved!\n", name)
	fmt.Printf("→ Try: sshtie connect %s\n\n", name)
	return nil
}

// prompt prints a labelled prompt with an optional default and reads a line.
// If required is true, it loops until the user enters something.
func prompt(r *bufio.Reader, label, defaultVal string, required bool) string {
	for {
		if defaultVal != "" {
			fmt.Printf("%-22s [%s]: ", label, defaultVal)
		} else {
			fmt.Printf("%-22s : ", label)
		}

		line, _ := r.ReadString('\n')
		line = strings.TrimSpace(line)

		if line == "" && defaultVal != "" {
			return defaultVal
		}
		if line == "" && required {
			fmt.Println("  (required — please enter a value)")
			continue
		}
		return line
	}
}
