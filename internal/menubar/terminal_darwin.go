//go:build darwin

package menubar

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// OpenConnect opens a terminal window and runs "sshtie connect <name>".
func OpenConnect(profileName string) {
	_ = openTerminal(fmt.Sprintf("%s connect %s", resolveBin(), profileName))
}

// OpenAdd opens a terminal window and runs "sshtie add".
func OpenAdd() {
	_ = openTerminal(fmt.Sprintf("%s add", resolveBin()))
}

// OpenEdit opens a terminal window and runs "sshtie edit <name>".
func OpenEdit(profileName string) {
	_ = openTerminal(fmt.Sprintf("%s edit %s", resolveBin(), profileName))
}

// OpenRename opens a terminal window and runs "sshtie rename <name>".
func OpenRename(profileName string) {
	_ = openTerminal(fmt.Sprintf("%s rename %s", resolveBin(), profileName))
}

// OpenRemove opens a terminal window and runs "sshtie remove <name>".
func OpenRemove(profileName string) {
	_ = openTerminal(fmt.Sprintf("%s remove %s", resolveBin(), profileName))
}

// openTerminal tries iTerm2 first, then falls back to Terminal.app.
func openTerminal(shellCmd string) error {
	if isRunning("iTerm2") {
		if err := withITerm2(shellCmd); err == nil {
			return nil
		}
	}
	return withTerminalApp(shellCmd)
}

func isRunning(appName string) bool {
	out, err := exec.Command("pgrep", "-x", appName).Output()
	return err == nil && strings.TrimSpace(string(out)) != ""
}

func withITerm2(cmd string) error {
	script := fmt.Sprintf(`
tell application "iTerm2"
	activate
	set w to (create window with default profile)
	tell current session of w
		write text "%s"
	end tell
end tell`, escAS(cmd))
	return exec.Command("osascript", "-e", script).Run()
}

func withTerminalApp(cmd string) error {
	script := fmt.Sprintf(`
tell application "Terminal"
	activate
	set w to do script "%s"
	tell window of w
		set bounds to {100, 100, 1000, 700}
	end tell
end tell`, escAS(cmd))
	return exec.Command("osascript", "-e", script).Run()
}

// escAS escapes a string for safe embedding in an AppleScript string literal.
func escAS(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

// resolveBin returns the absolute path of the sshtie CLI binary.
// Priority: next to this executable (bundle layout) → Homebrew → PATH.
func resolveBin() string {
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "sshtie")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	for _, p := range []string{
		"/opt/homebrew/bin/sshtie", // Apple Silicon Homebrew
		"/usr/local/bin/sshtie",   // Intel Homebrew
	} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "sshtie"
}
