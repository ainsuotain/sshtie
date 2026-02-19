//go:build !windows

package session

import (
	"os"
	"syscall"
)

// IsAlive reports whether the process with the given PID is still running.
// It uses signal(0) which returns an error only if the PID does not exist
// or the caller has no permission to signal it.
func IsAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// signal(0) probes existence without actually sending a signal.
	err = p.Signal(syscall.Signal(0))
	return err == nil || err == syscall.EPERM
}
