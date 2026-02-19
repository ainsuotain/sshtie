package menubar

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

// iconBytes returns a 22×22 PNG for the macOS menu-bar template icon.
// It draws a ">" chevron (terminal-prompt glyph) in solid black on a
// transparent background.  macOS inverts template icons automatically for
// Dark Mode, so no extra work is needed.
func iconBytes() []byte {
	const size = 22
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	black := color.RGBA{R: 0, G: 0, B: 0, A: 255}

	// ">" chevron: two arms meeting at the right-centre point (16, 11).
	// Top arm  : (4, 4)  → (16, 11)   dx=12 dy=7
	// Bottom arm: (16,11) → (4, 18)   dx=-12 dy=7
	drawLine(img, 4, 4, 16, 11, black)
	drawLine(img, 16, 11, 4, 18, black)

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

// drawLine draws a 2-pixel-wide line using integer interpolation.
func drawLine(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	dx := x1 - x0
	dy := y1 - y0
	steps := abs(dx)
	if abs(dy) > steps {
		steps = abs(dy)
	}
	if steps == 0 {
		img.Set(x0, y0, c)
		return
	}
	for i := 0; i <= steps; i++ {
		x := x0 + i*dx/steps
		y := y0 + i*dy/steps
		img.Set(x, y, c)
		img.Set(x+1, y, c) // 2 px thick
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
