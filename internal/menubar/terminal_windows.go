//go:build windows

package menubar

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/windows"
)

// OpenConnect opens a terminal and runs "sshtie connect <name>".
// Prefers WSL (gets mosh support) over native Windows terminal.
//
// For native Windows: sshtie.exe is spawned directly in its own console window
// (no CMD wrapper). SSHTIE_KEEP_WINDOW=1 tells sshtie to hide the window rather
// than exit when the user clicks X, keeping the SSH session alive in background.
func OpenConnect(profileName string) {
	if openWSL("connect " + profileName) {
		return
	}
	_ = openWindowsConnect(profileName)
}

// openWindowsConnect spawns sshtie.exe directly in a new console window.
// Compared with the CMD-wrapper approach this:
//   - eliminates the blank black window flash at startup
//   - enables the window-hide-on-close behaviour via SSHTIE_KEEP_WINDOW
func openWindowsConnect(profileName string) error {
	bin := resolveBin()
	cmd := exec.Command(bin, "connect", profileName)
	cmd.Env = append(os.Environ(), "SSHTIE_KEEP_WINDOW=1")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: windows.CREATE_NEW_CONSOLE,
	}
	return cmd.Start()
}

// OpenAdd opens a terminal and runs "sshtie add".
func OpenAdd() {
	if openWSL("add") {
		return
	}
	_ = openWindowsTerminal(fmt.Sprintf("%s add", resolveBin()))
}

// OpenEdit opens a terminal and runs "sshtie edit <name>".
func OpenEdit(profileName string) {
	if openWSL("edit " + profileName) {
		return
	}
	_ = openWindowsTerminal(fmt.Sprintf("%s edit %s", resolveBin(), profileName))
}

// OpenRename opens a terminal and runs "sshtie rename <name>".
func OpenRename(profileName string) {
	if openWSL("rename " + profileName) {
		return
	}
	_ = openWindowsTerminal(fmt.Sprintf("%s rename %s", resolveBin(), profileName))
}

// OpenRemove opens a terminal and runs "sshtie remove <name>".
func OpenRemove(profileName string) {
	if openWSL("remove " + profileName) {
		return
	}
	_ = openWindowsTerminal(fmt.Sprintf("%s remove %s", resolveBin(), profileName))
}

// ── WSL support ───────────────────────────────────────────────────────────────

// openWSL tries to open a WSL terminal running "sshtie <args>".
// Returns true (and starts the terminal) if WSL + sshtie-in-WSL are available.
// WSL gives full mosh support, unlike native Windows terminals.
func openWSL(sshtieArgs string) bool {
	if _, err := exec.LookPath("wsl.exe"); err != nil {
		return false // WSL not installed
	}
	// Check that sshtie is available inside WSL.
	// HideWindow suppresses the brief black console flash that appears when a
	// GUI-only process (the tray) spawns a console application (wsl.exe).
	check := exec.Command("wsl.exe", "which", "sshtie")
	check.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := check.Run(); err != nil {
		return false // sshtie not found in WSL PATH
	}

	// Wrap with a "session ended" prompt so the window stays open.
	bashCmd := fmt.Sprintf(
		`sshtie %s; echo ""; echo "  Session ended. Press Enter to close."; read`,
		sshtieArgs,
	)

	// Prefer Windows Terminal for best ANSI support.
	if wt, err := exec.LookPath("wt.exe"); err == nil {
		_ = exec.Command(wt, "new-tab", "--", "wsl.exe", "bash", "-c", bashCmd).Start()
		return true
	}
	// Fall back to launching WSL directly (opens in a new console window).
	_ = exec.Command("wsl.exe", "bash", "-c", bashCmd).Start()
	return true
}

// ── Native Windows terminal ───────────────────────────────────────────────────

// openWindowsTerminal opens a native Windows terminal (no mosh).
// Priority: Windows Terminal → PowerShell → cmd.exe
func openWindowsTerminal(shellCmd string) error {
	wrapped := fmt.Sprintf(
		`echo. & echo  sshtie - connecting... & echo. & %s & echo. & echo  Session ended. Press any key to close. & pause >nul`,
		shellCmd,
	)

	if wt, err := exec.LookPath("wt.exe"); err == nil {
		return exec.Command(wt, "new-tab", "--", "cmd.exe", "/K", wrapped).Start()
	}
	for _, ps := range []string{"pwsh.exe", "powershell.exe"} {
		if _, err := exec.LookPath(ps); err == nil {
			psCmd := fmt.Sprintf(`Write-Host ""; Write-Host "  sshtie - connecting..." -ForegroundColor Cyan; Write-Host ""; & %s`, shellCmd)
			return exec.Command(ps, "-NoExit", "-Command", psCmd).Start()
		}
	}
	return exec.Command("cmd.exe", "/C", "start", "cmd.exe", "/K", wrapped).Start()
}

// ── Binary resolution ─────────────────────────────────────────────────────────

// resolveBin returns the absolute path of the native Windows sshtie binary.
// Priority: next to this executable → Scoop → PATH.
func resolveBin() string {
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "sshtie.exe")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		p := filepath.Join(home, "scoop", "shims", "sshtie.exe")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "sshtie"
}
