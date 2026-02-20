//go:build windows

package menubar

import (
	"bytes"
	"encoding/binary"
	"os/exec"
	"strings"
)

// iconBytes returns the ICO-wrapped icon suited to the current appearance.
// In dark mode: ivory chevron. In light mode: black chevron.
func iconBytes() []byte {
	if darkMode() {
		return pngToICO(generateLightPNG())
	}
	return pngToICO(generatePNG())
}

// darkMode reports whether Windows is currently in Dark Mode.
// It reads the Apps theme preference via PowerShell to avoid cgo.
func darkMode() bool {
	out, err := exec.Command("powershell", "-NoProfile", "-Command",
		`(Get-ItemProperty -Path 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Themes\Personalize' -Name AppsUseLightTheme -ErrorAction SilentlyContinue).AppsUseLightTheme`,
	).Output()
	if err != nil {
		return false // assume light mode on error
	}
	return strings.TrimSpace(string(out)) == "0"
}

// pngToICO wraps raw PNG bytes in a minimal single-image ICO container.
func pngToICO(pngData []byte) []byte {
	var buf bytes.Buffer
	le := binary.LittleEndian

	// ICONDIR header (6 bytes)
	binary.Write(&buf, le, uint16(0)) // reserved
	binary.Write(&buf, le, uint16(1)) // type: 1 = icon
	binary.Write(&buf, le, uint16(1)) // image count

	// ICONDIRENTRY (16 bytes)
	buf.WriteByte(22) // width  (22 px)
	buf.WriteByte(22) // height (22 px)
	buf.WriteByte(0)  // color count (0 = more than 256)
	buf.WriteByte(0)  // reserved
	binary.Write(&buf, le, uint16(1))            // color planes
	binary.Write(&buf, le, uint16(32))           // bits per pixel
	binary.Write(&buf, le, uint32(len(pngData))) // size of image data
	binary.Write(&buf, le, uint32(6+16))         // offset to image data

	// Actual PNG payload
	buf.Write(pngData)
	return buf.Bytes()
}
