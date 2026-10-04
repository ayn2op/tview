// Package image draws images with half-block characters or kitty's graphics protocol.
package image

import (
	"github.com/ayn2op/tview/layout"
	"image"
	stdcolor "image/color"

	"github.com/ayn2op/tview"
	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

// Pixels less opaque than minAlpha are drawn as transparent, so antialiased edges do not show as dark cells.
const (
	upperHalfBlock = "▀"
	lowerHalfBlock = "▄"
)

const minAlpha = 50

// The cell size assumed when the terminal's is unknown: twice as tall as wide.
const (
	defaultCellWidth  = 1
	defaultCellHeight = 2
)

// Widget draws an image scaled to fit its area, preserving its aspect ratio.
type Widget struct {
	src                   image.Image
	width                 int
	cellWidth, cellHeight int
	kitty                 int
}

var _ tview.Widget = Widget{}

// New returns a widget that draws src at one pixel per column.
func New(src image.Image) Widget {
	return Widget{src: src}
}

// Width sets the width in cells; the height follows from the aspect ratio.
func (w Widget) Width(cells int) Widget {
	w.width = cells
	return w
}

// CellSize sets the size of a terminal cell in pixels, so the image keeps its aspect ratio.
// A zero width or height leaves the cell assumed twice as tall as wide.
func (w Widget) CellSize(width, height int) Widget {
	w.cellWidth, w.cellHeight = width, height
	return w
}

// Size returns the width and the height that preserves the aspect ratio.
func (w Widget) Size() (width, height layout.Length) {
	cols, rows := w.cells()
	return layout.Fixed(cols), layout.Fixed(rows)
}

// Layout returns the size of the image within limits.
func (w Widget) Layout(limits layout.Limits) layout.Size {
	width, height := w.Size()
	return layout.Atomic(limits, width, height)
}

func (w Widget) cells() (cols, rows int) {
	b := w.src.Bounds()
	if b.Empty() {
		return 0, 0
	}
	cellWidth, cellHeight := w.cell()
	cols = w.width
	if cols == 0 {
		cols = b.Dx()
	}
	return cols, (cols*cellWidth*b.Dy()/b.Dx() + cellHeight - 1) / cellHeight
}

func (w Widget) cell() (width, height int) {
	if w.cellWidth <= 0 || w.cellHeight <= 0 {
		return defaultCellWidth, defaultCellHeight
	}
	return w.cellWidth, w.cellHeight
}

// Draw draws the image at the top-left corner of area.
// Transparent pixels leave the underlying cells untouched.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	if w.kitty != 0 {
		w.drawPlaceholders(screen, area)
		return
	}
	b := w.src.Bounds()
	if b.Empty() {
		return
	}
	// Each cell holds two pixels, one above the other, so height counts half cells.
	cellWidth, cellHeight := w.cell()
	width := min(area.Width, area.Height*cellHeight*b.Dx()/(cellWidth*b.Dy()))
	height := width * cellWidth * b.Dy() * 2 / (b.Dx() * cellHeight)
	pixel := func(x, y int) color.Color {
		if y >= height {
			return color.Default
		}
		c := stdcolor.NRGBAModel.Convert(w.src.At(b.Min.X+x*b.Dx()/width, b.Min.Y+y*b.Dy()/height)).(stdcolor.NRGBA)
		if c.A < minAlpha {
			return color.Default
		}
		return color.NewRGBColor(int32(c.R), int32(c.G), int32(c.B))
	}
	for y := 0; y < height; y += 2 {
		for x := range width {
			top, bottom := pixel(x, y), pixel(x, y+1)
			switch {
			case top != color.Default:
				screen.Put(area.X+x, area.Y+y/2, upperHalfBlock, tcell.StyleDefault.Foreground(top).Background(bottom))
			case bottom != color.Default:
				screen.Put(area.X+x, area.Y+y/2, lowerHalfBlock, tcell.StyleDefault.Foreground(bottom))
			}
		}
	}
}

// Handle passes msg through unchanged.
func (Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	return msg
}
