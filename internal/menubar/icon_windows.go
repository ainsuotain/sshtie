//go:build windows

package menubar

import (
	"bytes"
	"encoding/binary"
)

// iconBytes returns the PNG wrapped in an ICO container.
// Windows systray requires ICO format; Vista+ supports PNG-inside-ICO.
func iconBytes() []byte { return pngToICO(generatePNG()) }

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
	binary.Write(&buf, le, uint16(1))               // color planes
	binary.Write(&buf, le, uint16(32))              // bits per pixel
	binary.Write(&buf, le, uint32(len(pngData)))    // size of image data
	binary.Write(&buf, le, uint32(6+16))            // offset to image data

	// Actual PNG payload
	buf.Write(pngData)
	return buf.Bytes()
}
