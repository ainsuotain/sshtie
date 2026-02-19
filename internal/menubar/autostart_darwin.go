//go:build darwin

package menubar

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
	"bytes"
)

const launchAgentLabel = "com.ainsuotain.sshtie-menubar"

var plistTemplate = template.Must(template.New("plist").Parse(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
    "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>{{.Label}}</string>
    <key>ProgramArguments</key>
    <array>
        <string>{{.ExePath}}</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <false/>
</dict>
</plist>
`))

func plistPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", launchAgentLabel+".plist")
}

// IsAutoStartEnabled reports whether the LaunchAgent plist exists.
func IsAutoStartEnabled() bool {
	_, err := os.Stat(plistPath())
	return err == nil
}

// EnableAutoStart writes the LaunchAgent plist and loads it.
func EnableAutoStart() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := plistTemplate.Execute(&buf, struct {
		Label   string
		ExePath string
	}{launchAgentLabel, exe}); err != nil {
		return err
	}

	path := plistPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return err
	}

	// Load the agent so it takes effect immediately (without reboot).
	return exec.Command("launchctl", "load", path).Run()
}

// DisableAutoStart unloads and removes the LaunchAgent plist.
func DisableAutoStart() error {
	path := plistPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	if err := exec.Command("launchctl", "unload", path).Run(); err != nil {
		return fmt.Errorf("launchctl unload: %w", err)
	}
	return os.Remove(path)
}
