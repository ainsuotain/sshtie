package connector

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
	sess "github.com/ainsuotain/sshtie/internal/session"
	"github.com/ainsuotain/sshtie/internal/tailscale"
)

func netAddr(host string, port int) string {
	return net.JoinHostPort(host, strconv.Itoa(port))
}

// Connect executes the best available connection strategy for the given profile.
// Strategy: mosh+tmux → ssh+tmux → ssh only
// On Windows, mosh is skipped entirely (not supported natively).
func Connect(p profile.Profile) error {
	port := p.Port
	if port == 0 {
		port = 22
	}
	session := p.TmuxSession
	if session == "" {
		session = "main"
	}

	// mosh is not available on Windows natively — skip straight to SSH.
	if runtime.GOOS == "windows" {
		fmt.Fprintln(os.Stderr, "→ Windows detected: mosh not supported, using SSH directly")
		return connectSSH(p, port, session)
	}

	// Tailscale routing check.
	switch p.Network {
	case "tailscale":
		// Profile explicitly requires Tailscale — fail fast if unavailable.
		if !tailscale.ClientRunning() {
			return fmt.Errorf("Tailscale is not running (profile requires network=tailscale)")
		}
		if !tailscale.HostInNetwork(p.Host) {
			return fmt.Errorf("host %q is not in the Tailscale network", p.Host)
		}
		fmt.Println("→ Routing via Tailscale")
	case "direct":
		// Skip Tailscale and mosh; go straight to SSH.
	default: // "auto"
		if tailscale.ClientRunning() && tailscale.HostInNetwork(p.Host) {
			fmt.Println("→ Tailscale detected: routing via Tailscale network")
		}
	}

	// Try mosh first (unless network mode is "direct").
	if p.Network != "direct" {
		if err := tryMosh(p, port, session); err == nil {
			return nil
		} else {
			if strings.Contains(err.Error(), "UDP port") {
				fmt.Fprintln(os.Stderr, "⚠  mosh: the server's firewall is blocking UDP ports (60000–61000).")
				fmt.Fprintln(os.Stderr, "   mosh needs these ports open to maintain a stable, reconnect-friendly session.")
				fmt.Fprintln(os.Stderr, "   To fix it, run this command on your server:")
				fmt.Fprintln(os.Stderr, "     sudo ufw allow 60000:61000/udp")
				fmt.Fprintln(os.Stderr, "→ Falling back to SSH for now.")
			} else {
				fmt.Fprintf(os.Stderr, "⚠  mosh failed (%v) — falling back to SSH + tmux.\n", err)
			}
		}
	}

	return connectSSH(p, port, session)
}

// connectSSH tries ssh+tmux then falls back to plain ssh.
func connectSSH(p profile.Profile, port int, session string) error {
	if err := trySSHTmux(p, port, session); err == nil {
		return nil
	} else {
		fmt.Fprintf(os.Stderr, "⚠  ssh+tmux failed (%v) — falling back to a plain SSH session.\n", err)
	}
	return trySSH(p, port)
}

// tryMosh launches mosh → tmux attach/new.
func tryMosh(p profile.Profile, port int, tmuxSession string) error {
	moshBin, err := findMosh()
	if err != nil {
		return fmt.Errorf("mosh not found in PATH")
	}

	// Quick UDP reachability check on default mosh port 60001.
	if !udpReachable(p.Host, 60001, 2*time.Second) {
		return fmt.Errorf("UDP port 60001 appears blocked")
	}

	args := []string{
		"--ssh=" + buildSSHFlag(p, port),
	}
	if p.MoshServer != "" {
		args = append(args, "--server="+p.MoshServer)
	}
	args = append(args, fmt.Sprintf("%s@%s", p.User, p.Host))
	args = append(args, "--", "tmux", "new-session", "-A", "-s", tmuxSession)

	cmd := exec.Command(moshBin, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	_ = sess.Write(sess.Session{Profile: p.Name, PID: cmd.Process.Pid, Method: "mosh", StartedAt: time.Now()})
	defer sess.Delete(p.Name)
	return cmd.Wait()
}

// trySSHTmux connects via SSH and attaches/creates a tmux session.
func trySSHTmux(p profile.Profile, port int, tmuxSession string) error {
	if !tcpReachable(p.Host, port, 5*time.Second) {
		return fmt.Errorf("TCP port %d unreachable", port)
	}

	remoteCmd := fmt.Sprintf("tmux new-session -A -s %s", tmuxSession)
	args := buildSSHBaseArgs(p, port)
	args = append(args, "-t", fmt.Sprintf("%s@%s", p.User, p.Host))
	args = append(args, remoteCmd)

	cmd := exec.Command("ssh", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	_ = sess.Write(sess.Session{Profile: p.Name, PID: cmd.Process.Pid, Method: "ssh+tmux", StartedAt: time.Now()})
	defer sess.Delete(p.Name)
	return cmd.Wait()
}

// trySSH does a plain SSH connection.
func trySSH(p profile.Profile, port int) error {
	args := buildSSHBaseArgs(p, port)
	args = append(args, fmt.Sprintf("%s@%s", p.User, p.Host))

	cmd := exec.Command("ssh", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	_ = sess.Write(sess.Session{Profile: p.Name, PID: cmd.Process.Pid, Method: "ssh", StartedAt: time.Now()})
	defer sess.Delete(p.Name)
	return cmd.Wait()
}

func buildSSHBaseArgs(p profile.Profile, port int) []string {
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

	args := []string{
		"-p", strconv.Itoa(port),
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", fmt.Sprintf("ServerAliveInterval=%d", aliveInterval),
		"-o", fmt.Sprintf("ServerAliveCountMax=%d", aliveCount),
		"-o", "TCPKeepAlive=yes",
		"-o", fmt.Sprintf("ConnectionAttempts=%d", attempts),
		// Disable multiplexing — causes subtle issues when tmux is involved.
		"-o", "ControlMaster=no",
		"-o", "ControlPath=none",
	}
	if p.ForwardAgent {
		args = append(args, "-o", "ForwardAgent=yes")
	}
	key := p.DefaultKey()
	if _, err := os.Stat(key); err == nil {
		args = append(args, "-i", key)
	}
	return args
}

// buildSSHFlag builds the --ssh= value for mosh.
func buildSSHFlag(p profile.Profile, port int) string {
	parts := []string{"ssh", "-p", strconv.Itoa(port)}
	key := p.DefaultKey()
	if _, err := os.Stat(key); err == nil {
		parts = append(parts, "-i", key)
	}
	return strings.Join(parts, " ")
}

func findMosh() (string, error) {
	if path, err := exec.LookPath("mosh"); err == nil {
		return path, nil
	}
	// Common non-PATH locations per platform.
	candidates := []string{
		"/opt/homebrew/bin/mosh", // macOS ARM Homebrew
		"/usr/local/bin/mosh",    // macOS Intel Homebrew / Linux manual install
		"/usr/bin/mosh",          // Linux (apt / dnf / pacman)
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}
	return "", fmt.Errorf("mosh not found")
}

func tcpReachable(host string, port int, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", netAddr(host, port), timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// udpReachable is a best-effort heuristic. Firewalls that silently drop UDP
// will still return true — mosh itself will confirm reachability.
func udpReachable(host string, port int, timeout time.Duration) bool {
	conn, err := net.DialTimeout("udp", netAddr(host, port), timeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(timeout))
	_, _ = conn.Write([]byte{})
	buf := make([]byte, 1)
	_, err = conn.Read(buf)
	if err != nil && strings.Contains(err.Error(), "connection refused") {
		return false
	}
	return true
}
