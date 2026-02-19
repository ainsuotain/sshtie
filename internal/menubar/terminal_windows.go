//go:build windows

package menubar

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// OpenConnect opens a terminal and runs "sshtie connect <name>".
func OpenConnect(profileName string) {
	_ = openTerminal(fmt.Sprintf("%s connect %s", resolveBin(), profileName))
}

// OpenAdd opens a terminal and runs "sshtie add".
func OpenAdd() {
	_ = openTerminal(fmt.Sprintf("%s add", resolveBin()))
}

// openTerminal opens a new terminal window running the given command.
// Priority: Windows Terminal → PowerShell → cmd.exe
func openTerminal(shellCmd string) error {
	// Windows Terminal (wt.exe) — available on Windows 10 1903+
	if _, err := exec.LookPath("wt.exe"); err == nil {
		return exec.Command("wt.exe", "new-tab", "--", "cmd.exe", "/K", shellCmd).Start()
	}
	// PowerShell
	if _, err := exec.LookPath("pwsh.exe"); err == nil {
		return exec.Command("pwsh.exe", "-NoExit", "-Command", shellCmd).Start()
	}
	// cmd.exe fallback
	return exec.Command("cmd.exe", "/C", "start", "cmd.exe", "/K", shellCmd).Start()
}

// resolveBin returns the absolute path of the sshtie CLI binary.
// Priority: next to this executable → Scoop → PATH.
func resolveBin() string {
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "sshtie.exe")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	// Scoop package manager
	if home, err := os.UserHomeDir(); err == nil {
		p := filepath.Join(home, "scoop", "shims", "sshtie.exe")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "sshtie"
}
