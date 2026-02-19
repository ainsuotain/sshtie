package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ainsuotain/sshtie/internal/doctor"
	"github.com/ainsuotain/sshtie/internal/profile"
)

var installTailscale bool

var installCmd = &cobra.Command{
	Use:   "install <name>",
	Short: "Install mosh-server, tmux (and optionally Tailscale) on a remote server",
	Long: `Connects to the remote server and installs mosh + tmux
using the appropriate package manager (apt / dnf / yum / brew / pacman).

Use --tailscale to also install Tailscale on the remote server.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		p, err := profile.Get(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Profile '%s' not found.\n", name)
			fmt.Fprintln(os.Stderr, "→ Run 'sshtie list' to see available profiles.")
			fmt.Fprintln(os.Stderr, "→ Run 'sshtie add' to register a new server.")
			return nil
		}
		return runInstall(p)
	},
}

func init() {
	installCmd.Flags().BoolVar(&installTailscale, "tailscale", false, "Also install Tailscale on the remote server")
}

// ── Types ─────────────────────────────────────────────────────────────────────

type remoteOS struct {
	display string // e.g. "Ubuntu 22.04 LTS", "macOS 14.2"
	pkgMgr  string // apt | dnf | yum | brew | pacman | ""
}

type pkgStep struct {
	label  string // printed label, e.g. "tmux..."
	binary string // checked with `which` on remote
	pkg    string // package name passed to the package manager
}

// ── Main install flow ─────────────────────────────────────────────────────────

func runInstall(p profile.Profile) error {
	port := p.Port
	if port == 0 {
		port = 22
	}

	fmt.Printf("\n🔧 Installing dependencies on %s (%s)...\n\n", p.Name, p.Host)

	// Step 1: OS detection
	fmt.Printf("  %-26s", "Detecting OS...")
	ros, err := detectRemoteOS(p, port)
	if err != nil {
		fmt.Println("⚠  SSH connection failed")
		fmt.Fprintf(os.Stderr, "  error: %v\n", err)
		return nil
	}

	switch ros.pkgMgr {
	case "":
		if ros.display == "macos-no-brew" {
			fmt.Println("⚠  macOS (Homebrew not installed)")
			fmt.Println()
			fmt.Println("  → Homebrew is not installed on the remote server.")
			fmt.Println("  → Install it first: https://brew.sh")
			return nil
		}
		fmt.Printf("⚠  %s — unknown package manager\n", ros.display)
		printManualInstallHint()
		return nil
	default:
		fmt.Printf("✅ %s (%s)\n", ros.display, ros.pkgMgr)
	}

	// Steps 2–3: install tmux and mosh-server
	steps := []pkgStep{
		{"tmux...", "tmux", "tmux"},
		{"mosh-server...", "mosh-server", "mosh"},
	}

	allOK := true
	for _, step := range steps {
		if !runPkgStep(p, port, ros, step) {
			allOK = false
		}
	}

	// Optional: install Tailscale
	if installTailscale {
		if !installRemoteTailscale(p, port, ros) {
			allOK = false
		}
	}

	// Summary
	fmt.Println()
	if allOK {
		fmt.Println("→ Server is ready!")
	} else {
		fmt.Println("→ Some packages failed to install. See hints above.")
	}
	fmt.Println("→ Running doctor check...")
	doctor.Run(p)
	fmt.Printf("→ Try: sshtie connect %s\n\n", p.Name)
	return nil
}

// runPkgStep checks if binary exists; installs it if not. Returns true on success.
func runPkgStep(p profile.Profile, port int, ros remoteOS, step pkgStep) bool {
	fmt.Printf("  %-26s", step.label)

	// Already installed?
	out, _ := remoteCapture(p, port, "which "+step.binary+" 2>/dev/null")
	if strings.TrimSpace(out) != "" {
		fmt.Println("✅ Already installed")
		return true
	}

	// Install
	cmdStr := buildInstallCmd(ros, step.pkg)
	fmt.Println("Installing...")
	if err := remoteInteractive(p, port, cmdStr); err != nil {
		fmt.Printf("  %-26s⚠  Failed\n", "")
		fmt.Fprintln(os.Stderr, "  → sudo 권한이 필요합니다. 서버 관리자에게 문의하세요.")
		fmt.Fprintf(os.Stderr, "  → Manual: %s\n", cmdStr)
		return false
	}
	fmt.Printf("  %-26s✅ Installed\n", "")
	return true
}

// ── OS Detection ──────────────────────────────────────────────────────────────

func detectRemoteOS(p profile.Profile, port int) (remoteOS, error) {
	// Single round-trip: /etc/os-release covers Linux; uname -s detects Darwin.
	out, err := remoteCapture(p, port, "cat /etc/os-release 2>/dev/null; uname -s 2>/dev/null")
	if err != nil {
		return remoteOS{}, err
	}

	kvs := make(map[string]string)
	var uname string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		switch line {
		case "Darwin", "Linux", "FreeBSD":
			uname = line
			continue
		}
		if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.Trim(strings.TrimSpace(parts[1]), `"`)
			kvs[k] = v
		}
	}

	id := strings.ToLower(kvs["ID"])
	idLike := strings.ToLower(kvs["ID_LIKE"])
	pretty := kvs["PRETTY_NAME"]
	if pretty == "" {
		pretty = id
	}

	// macOS: /etc/os-release doesn't exist, so id is empty and uname is Darwin.
	if uname == "Darwin" && id == "" {
		brewPath, _ := remoteCapture(p, port, "which brew 2>/dev/null")
		if strings.TrimSpace(brewPath) == "" {
			return remoteOS{display: "macos-no-brew"}, nil
		}
		macVer, _ := remoteCapture(p, port, "sw_vers -productVersion 2>/dev/null")
		return remoteOS{
			display: fmt.Sprintf("macOS %s", strings.TrimSpace(macVer)),
			pkgMgr:  "brew",
		}, nil
	}

	switch {
	case id == "ubuntu" || id == "debian" ||
		strings.Contains(idLike, "ubuntu") || strings.Contains(idLike, "debian"):
		return remoteOS{display: pretty, pkgMgr: "apt"}, nil

	case id == "fedora" || strings.Contains(idLike, "fedora"):
		return remoteOS{display: pretty, pkgMgr: "dnf"}, nil

	case id == "centos" || id == "rhel" || id == "almalinux" || id == "rocky" ||
		strings.Contains(idLike, "rhel") || strings.Contains(idLike, "centos"):
		// Prefer dnf (el8+) over yum (el7).
		dnfPath, _ := remoteCapture(p, port, "which dnf 2>/dev/null")
		if strings.TrimSpace(dnfPath) != "" {
			return remoteOS{display: pretty, pkgMgr: "dnf"}, nil
		}
		return remoteOS{display: pretty, pkgMgr: "yum"}, nil

	case id == "arch" || id == "manjaro" || strings.Contains(idLike, "arch"):
		return remoteOS{display: pretty, pkgMgr: "pacman"}, nil

	default:
		return remoteOS{display: pretty, pkgMgr: ""}, nil
	}
}

// ── Package manager helpers ───────────────────────────────────────────────────

func buildInstallCmd(ros remoteOS, pkg string) string {
	switch ros.pkgMgr {
	case "apt":
		return fmt.Sprintf("sudo apt-get install -y %s", pkg)
	case "dnf":
		return fmt.Sprintf("sudo dnf install -y %s", pkg)
	case "yum":
		return fmt.Sprintf("sudo yum install -y %s", pkg)
	case "brew":
		return fmt.Sprintf("brew install %s", pkg)
	case "pacman":
		return fmt.Sprintf("sudo pacman -S --noconfirm %s", pkg)
	default:
		return ""
	}
}

func printManualInstallHint() {
	fmt.Println()
	fmt.Println("  → Install manually:")
	fmt.Println("    Ubuntu/Debian : sudo apt-get install -y tmux mosh")
	fmt.Println("    CentOS/RHEL   : sudo yum install -y tmux mosh")
	fmt.Println("    Fedora        : sudo dnf install -y tmux mosh")
	fmt.Println("    Arch          : sudo pacman -S --noconfirm tmux mosh")
	fmt.Println("    macOS         : brew install tmux mosh")
}

// ── Tailscale installer ───────────────────────────────────────────────────────

// installRemoteTailscale installs Tailscale on the remote server and prints
// instructions for authenticating. Returns true on success.
func installRemoteTailscale(p profile.Profile, port int, ros remoteOS) bool {
	fmt.Printf("  %-26s", "tailscale...")

	// Already installed?
	out, _ := remoteCapture(p, port, "which tailscale 2>/dev/null")
	if strings.TrimSpace(out) != "" {
		fmt.Println("✅ Already installed")
		printTailscaleAuthHint(p.Name)
		return true
	}

	fmt.Println("Installing...")

	var cmdStr string
	if ros.pkgMgr == "brew" {
		cmdStr = "brew install tailscale"
	} else {
		// Official Tailscale install script — works on all major Linux distros.
		cmdStr = "curl -fsSL https://tailscale.com/install.sh | sh"
	}

	if err := remoteInteractive(p, port, cmdStr); err != nil {
		fmt.Printf("  %-26s⚠  Failed\n", "")
		fmt.Fprintln(os.Stderr, "  → Manual: https://tailscale.com/download/linux")
		return false
	}
	fmt.Printf("  %-26s✅ Installed\n", "")
	printTailscaleAuthHint(p.Name)
	return true
}

func printTailscaleAuthHint(profileName string) {
	fmt.Println()
	fmt.Println("  → Authenticate Tailscale on the server:")
	fmt.Println("    sudo tailscale up")
	fmt.Println("  → Then update the profile with the server's Tailscale IP:")
	fmt.Printf("    sshtie edit %s\n", profileName)
	fmt.Println()
}

// ── SSH helpers ───────────────────────────────────────────────────────────────

// remoteCapture runs a command non-interactively and returns stdout.
// Stdin and Stderr are forwarded so password prompts appear on the terminal.
func remoteCapture(p profile.Profile, port int, command string) (string, error) {
	args := installSSHArgs(p, port, false)
	args = append(args, command)
	cmd := exec.Command("ssh", args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

// remoteInteractive runs a command with a PTY allocated, allowing sudo prompts.
func remoteInteractive(p profile.Profile, port int, command string) error {
	args := installSSHArgs(p, port, true)
	args = append(args, command)
	c := exec.Command("ssh", args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func installSSHArgs(p profile.Profile, port int, interactive bool) []string {
	args := []string{
		"-p", strconv.Itoa(port),
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "ConnectTimeout=10",
		// No BatchMode — allows password authentication on servers without SSH keys.
	}
	if interactive {
		args = append(args, "-t") // PTY for sudo / password prompts
	}
	key := p.DefaultKey()
	if _, err := os.Stat(key); err == nil {
		args = append(args, "-i", key)
	}
	args = append(args, fmt.Sprintf("%s@%s", p.User, p.Host))
	return args
}
