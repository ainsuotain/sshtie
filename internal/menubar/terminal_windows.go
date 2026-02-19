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

// OpenEdit opens a terminal and runs "sshtie edit <name>".
func OpenEdit(profileName string) {
	_ = openTerminal(fmt.Sprintf("%s edit %s", resolveBin(), profileName))
}

// openTerminal opens a new terminal window running the given command.
// Priority: Windows Terminal → PowerShell → cmd.exe
func openTerminal(shellCmd string) error {
	// Wrap command with a banner so users see immediate feedback.
	wrapped := fmt.Sprintf(
		`echo. & echo  sshtie - connecting... & echo. & %s & echo. & echo  Session ended. Press any key to close. & pause >nul`,
		shellCmd,
	)

	// Windows Terminal (wt.exe) — best ANSI support, Windows 10 1903+
	if wt, err := exec.LookPath("wt.exe"); err == nil {
		return exec.Command(wt, "new-tab", "--", "cmd.exe", "/K", wrapped).Start()
	}
	// PowerShell (pwsh or legacy powershell)
	for _, ps := range []string{"pwsh.exe", "powershell.exe"} {
		if _, err := exec.LookPath(ps); err == nil {
			psCmd := fmt.Sprintf(`Write-Host ""; Write-Host "  sshtie - connecting..." -ForegroundColor Cyan; Write-Host ""; & %s`, shellCmd)
			return exec.Command(ps, "-NoExit", "-Command", psCmd).Start()
		}
	}
	// cmd.exe fallback
	return exec.Command("cmd.exe", "/C", "start", "cmd.exe", "/K", wrapped).Start()
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
