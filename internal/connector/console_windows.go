//go:build windows

package connector

import "syscall"

const (
	ctrlCloseEvent = 2 // sent when user clicks X on the console window
	swHide         = 0
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	user32   = syscall.NewLazyDLL("user32.dll")

	procSetConsoleCtrlHandler = kernel32.NewProc("SetConsoleCtrlHandler")
	procGetConsoleWindow      = kernel32.NewProc("GetConsoleWindow")
	procShowWindow            = user32.NewProc("ShowWindow")
)

// InstallWindowHideHandler replaces the default CTRL_CLOSE_EVENT handler with
// one that hides the console window instead of terminating the process.
// This lets the SSH session keep running in the background after the user
// clicks the X button — [connected] stays visible in the tray.
//
// The process exits normally when SSH itself exits (user typed exit /
// tmux session ended). Disconnect from the tray kills the SSH process.
func InstallWindowHideHandler() {
	cb := syscall.NewCallback(func(ctrlType uint32) uintptr {
		if ctrlType != ctrlCloseEvent {
			return 0 // let default handler run for Ctrl+C, Ctrl+Break, etc.
		}
		hwnd, _, _ := procGetConsoleWindow.Call()
		if hwnd != 0 {
			procShowWindow.Call(hwnd, swHide)
		}
		return 1 // TRUE: handled — do NOT terminate the process
	})
	procSetConsoleCtrlHandler.Call(cb, 1) // 1 = add
}
