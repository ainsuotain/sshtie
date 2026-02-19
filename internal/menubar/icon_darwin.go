//go:build darwin

package menubar

// iconBytes returns a PNG for macOS template icon.
func iconBytes() []byte { return generatePNG() }
