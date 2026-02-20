//go:build darwin

package menubar

import (
	"os/exec"
	"strings"
)

// iconBytes returns the correct icon for the current appearance.
// In dark mode: ivory chevron (warm white).
// In light mode: black chevron via template (macOS auto-inverts on system transitions).
func iconBytes() []byte {
	if darkMode() {
		return generateLightPNG()
	}
	return generatePNG()
}

// darkMode reports whether macOS is currently in Dark Mode.
func darkMode() bool {
	out, err := exec.Command("defaults", "read", "-g", "AppleInterfaceStyle").Output()
	return err == nil && strings.TrimSpace(string(out)) == "Dark"
}
