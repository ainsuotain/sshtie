package doctor

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/ainsuotain/sshtie/internal/profile"
	"github.com/ainsuotain/sshtie/internal/tailscale"
)

func netAddr(host string, port int) string {
	return net.JoinHostPort(host, strconv.Itoa(port))
}

// Result holds the outcome of a single diagnostic check.
type Result struct {
	Label  string
	OK     bool
	Detail string
}

// Run performs all diagnostic checks for the given profile and prints results.
func Run(p profile.Profile) {
	port := p.Port
	if port == 0 {
		port = 22
	}

	fmt.Printf("\n🔍 Diagnosing: %s (%s)\n\n", p.Name, p.Host)

	results := []Result{
		checkSSH(p, port),
		checkMoshServer(p, port),
		checkUDP(p.Host, 60001),
		checkTmux(p, port),
		checkTailscaleClient(),
		checkTailscaleServer(p.Host),
	}

	// On Windows, mosh is not supported natively — mark checks as skipped.
	if runtime.GOOS == "windows" {
		results[1] = Result{"mosh-server", false, "Skipped (not supported on Windows)"}
		results[2] = Result{fmt.Sprintf("UDP port %d", 60001), false, "Skipped (not supported on Windows)"}
	}

	strategy := "ssh only"
	moshOK := false
	tmuxOK := false

	for _, r := range results {
		icon := "✅"
		if !r.OK {
			icon = "⚠ "
		}
		fmt.Printf("  %-20s %s %s\n", r.Label, icon, r.Detail)
		if r.Label == "mosh-server" && r.OK {
			moshOK = true
		}
		if r.Label == "tmux" && r.OK {
			tmuxOK = true
		}
	}

	if moshOK && tmuxOK {
		strategy = "mosh + tmux"
	} else if tmuxOK {
		strategy = "ssh + tmux"
	}

	fmt.Printf("\n→ Recommended strategy: %s\n", strategy)
	if results[0].OK {
		fmt.Println("→ Ready to connect!")
	} else {
		fmt.Println("→ SSH unreachable — check host/port/key.")
	}
	fmt.Println()
}

func checkSSH(p profile.Profile, port int) Result {
	conn, err := net.DialTimeout("tcp", netAddr(p.Host, port), 5*time.Second)
	if err != nil {
		return Result{"SSH connection", false, err.Error()}
	}
	conn.Close()
	return Result{"SSH connection", true, "OK"}
}

func checkMoshServer(p profile.Profile, port int) Result {
	// Ask the remote server where mosh-server is via SSH.
	args := buildSSHArgs(p, port)
	args = append(args, "which mosh-server 2>/dev/null || echo ''")

	out, err := exec.Command("ssh", args...).Output()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		// Try well-known paths
		candidates := []string{
			"/opt/homebrew/bin/mosh-server",
			"/usr/local/bin/mosh-server",
			"/usr/bin/mosh-server",
		}
		for _, c := range candidates {
			checkArgs := buildSSHArgs(p, port)
			checkArgs = append(checkArgs, fmt.Sprintf("test -x %s && echo %s", c, c))
			o, e := exec.Command("ssh", checkArgs...).Output()
			if e == nil && strings.TrimSpace(string(o)) != "" {
				return Result{"mosh-server", true, fmt.Sprintf("Found (%s)", strings.TrimSpace(string(o)))}
			}
		}
		return Result{"mosh-server", false, "Not found on remote"}
	}
	path := strings.TrimSpace(string(out))
	return Result{"mosh-server", true, fmt.Sprintf("Found (%s)", path)}
}

func checkUDP(host string, port int) Result {
	label := fmt.Sprintf("UDP port %d", port)
	conn, err := net.DialTimeout("udp", netAddr(host, port), 3*time.Second)
	if err != nil {
		return Result{label, false, err.Error()}
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	conn.Write([]byte{})
	buf := make([]byte, 1)
	_, readErr := conn.Read(buf)
	if readErr != nil && strings.Contains(readErr.Error(), "connection refused") {
		return Result{label, false, "ICMP port unreachable"}
	}
	return Result{label, true, "Open (or filtered — mosh will confirm)"}
}

func checkTmux(p profile.Profile, port int) Result {
	args := buildSSHArgs(p, port)
	args = append(args, "tmux -V 2>/dev/null || echo ''")

	out, err := exec.Command("ssh", args...).Output()
	version := strings.TrimSpace(string(out))
	if err != nil || version == "" {
		return Result{"tmux", false, "Not found on remote"}
	}
	return Result{"tmux", true, version + " installed"}
}

func checkTailscaleClient() Result {
	if tailscale.ClientRunning() {
		return Result{"Tailscale (client)", true, "Running"}
	}
	if _, err := exec.LookPath("tailscale"); err != nil {
		return Result{"Tailscale (client)", false, "Not installed (optional)"}
	}
	return Result{"Tailscale (client)", false, "Installed but not running"}
}

func checkTailscaleServer(host string) Result {
	if !tailscale.ClientRunning() {
		return Result{"Tailscale (server)", false, "Skipped (client not running)"}
	}
	if tailscale.HostInNetwork(host) {
		return Result{"Tailscale (server)", true, "Found in Tailscale network"}
	}
	return Result{"Tailscale (server)", false, "Not in Tailscale network"}
}

func buildSSHArgs(p profile.Profile, port int) []string {
	args := []string{
		"-p", fmt.Sprintf("%d", port),
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "ConnectTimeout=5",
		"-o", "BatchMode=yes",
	}
	key := p.DefaultKey()
	if _, err := os.Stat(key); err == nil {
		args = append(args, "-i", key)
	}
	args = append(args, fmt.Sprintf("%s@%s", p.User, p.Host))
	return args
}
