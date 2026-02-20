//go:build !windows

package connector

// InstallWindowHideHandler is a no-op on non-Windows platforms.
func InstallWindowHideHandler() {}
