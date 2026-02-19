package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

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
		port := p.Port
		if port == 0 {
			port = 22
		}

		// 1. Fingerprint check — warn on first-time connections.
		if !isKnownHost(p.Host, port) {
			fmt.Printf("⚠  처음 접속하는 서버입니다 (%s)\n", p.Host)
			fmt.Println("   SSH 키가 자동으로 저장됩니다.")
			if !promptYN("계속할까요? (y/n): ") {
				return nil
			}
		}

		// 2. Quick remote dep check (key-auth only).
		//    If the server requires a password, BatchMode fails → returns (true, true)
		//    → skip and let connector handle the fallback gracefully.
		hasMosh, hasTmux := quickCheckRemoteDeps(p, port)
		if !hasMosh || !hasTmux {
			missing := []string{}
			if !hasMosh {
				missing = append(missing, "mosh-server")
			}
			if !hasTmux {
				missing = append(missing, "tmux")
			}
			fmt.Printf("⚠  서버에 %s 가 설치되어 있지 않습니다.\n", strings.Join(missing, ", "))
			if promptYN("지금 설치할까요? (y/n): ") {
				if err := runInstall(p); err != nil {
					return err
				}
				fmt.Println()
			}
		}

		fmt.Printf("→ Connecting to %s (%s@%s)…\n", p.Name, p.User, p.Host)
		return connector.Connect(p)
	},
}

// isKnownHost reports whether the host is already in ~/.ssh/known_hosts.
// Returns true (assume known) if ssh-keygen is unavailable.
func isKnownHost(host string, port int) bool {
	target := host
	if port != 0 && port != 22 {
		target = fmt.Sprintf("[%s]:%d", host, port)
	}
	err := exec.Command("ssh-keygen", "-F", target).Run()
	if err == nil {
		return true // exit 0 = host found in known_hosts
	}
	if _, ok := err.(*exec.ExitError); ok {
		return false // exit non-0 = host not in known_hosts
	}
	return true // ssh-keygen unavailable — assume known to avoid noise
}

// quickCheckRemoteDeps checks mosh-server and tmux presence on the remote
// using BatchMode SSH (key-auth only). On any failure returns (true, true)
// so the caller can proceed normally.
func quickCheckRemoteDeps(p profile.Profile, port int) (hasMosh, hasTmux bool) {
	args := []string{
		"-p", strconv.Itoa(port),
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "ConnectTimeout=5",
		"-o", "BatchMode=yes",
	}
	key := p.DefaultKey()
	if _, err := os.Stat(key); err == nil {
		args = append(args, "-i", key)
	}
	args = append(args, fmt.Sprintf("%s@%s", p.User, p.Host))
	args = append(args, "which mosh-server >/dev/null 2>&1 && echo M; which tmux >/dev/null 2>&1 && echo T")

	out, err := exec.Command("ssh", args...).Output()
	if err != nil {
		return true, true // can't check — let connector handle it
	}
	s := string(out)
	return strings.Contains(s, "M"), strings.Contains(s, "T")
}

// promptYN prints msg and reads a y/n answer from stdin. Returns true for "y".
func promptYN(msg string) bool {
	fmt.Print(msg)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return strings.ToLower(strings.TrimSpace(scanner.Text())) == "y"
	}
	return false
}
