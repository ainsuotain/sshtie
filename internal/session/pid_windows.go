//go:build windows

package session

import (
	"golang.org/x/sys/windows"
)

// IsAlive reports whether the process with the given PID is still running
// by attempting to open a handle to it.
func IsAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	defer windows.CloseHandle(h)

	var code uint32
	if err := windows.GetExitCodeProcess(h, &code); err != nil {
		return false
	}
	// STILL_ACTIVE == 259
	return code == 259
}
