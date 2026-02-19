//go:build windows

package menubar

import (
	"os"
	"os/exec"
)

const regKey = `HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Run`
const regVal = "sshtie"

// IsAutoStartEnabled reports whether the registry entry exists.
func IsAutoStartEnabled() bool {
	err := exec.Command("reg", "query", regKey, "/v", regVal).Run()
	return err == nil
}

// EnableAutoStart adds the registry entry to run at Windows startup.
func EnableAutoStart() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	return exec.Command(
		"reg", "add", regKey,
		"/v", regVal,
		"/t", "REG_SZ",
		"/d", exe,
		"/f",
	).Run()
}

// DisableAutoStart removes the registry entry.
func DisableAutoStart() error {
	return exec.Command(
		"reg", "delete", regKey,
		"/v", regVal,
		"/f",
	).Run()
}
