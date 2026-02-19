//go:build darwin || windows

package main

import "github.com/ainsuotain/sshtie/internal/menubar"

func main() {
	menubar.Run()
}
