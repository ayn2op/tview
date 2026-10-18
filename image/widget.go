// Package image draws images with half-block characters or kitty's graphics protocol.
package image

import (
	"image"
	stdcolor "image/color"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/layout"
	"github.com/ayn2op/tview/text"
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
// Transparent pixels leave the underlying cells untouched, and a transparent half of a cell keeps its background.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	cols, rows := w.fit(area.Width, area.Height)
	for y := range rows {
		for x := range cols {
			if str, style, ok := w.at(x, y, cols); ok {
				if style.GetBackground() == color.Default {
					_, under, _ := screen.Get(area.X+x, area.Y+y)
					style = style.Background(under.GetBackground())
				}
				screen.Put(area.X+x, area.Y+y, str, style)
			}
		}
	}
}

// Lines returns the image as text, a line of spans per row of its cells, to show it within other text.
// Transparent cells are spaces.
func (w Widget) Lines() [][]text.Span {
	cols, rows := w.fit(w.cells())
	lines := make([][]text.Span, rows)
	for y := range rows {
		for x := range cols {
			str, style, ok := w.at(x, y, cols)
			if !ok {
				str, style = " ", tcell.StyleDefault
			}
			lines[y] = append(lines[y], text.Span{Text: str, Style: style})
		}
	}
	return lines
}

// fit returns the columns and rows the image takes within width by height cells.
func (w Widget) fit(width, height int) (cols, rows int) {
	if w.kitty != 0 {
		cols, rows = w.cells()
		return min(cols, width, len(diacritics)), min(rows, height, len(diacritics))
	}

	b := w.src.Bounds()
	if b.Empty() {
		return 0, 0
	}

	cellWidth, cellHeight := w.cell()
	cols = min(width, height*cellHeight*b.Dx()/(cellWidth*b.Dy()))
	return cols, (w.pixelHeight(cols) + 1) / 2
}

// pixelHeight returns the height of the image cols wide in pixels of half a cell, as each cell holds two pixels, one above the other.
func (w Widget) pixelHeight(cols int) int {
	b := w.src.Bounds()
	cellWidth, cellHeight := w.cell()
	return cols * cellWidth * b.Dy() * 2 / (b.Dx() * cellHeight)
}

// at returns the cell at x, y of the image drawn cols wide, or false if it is transparent.
func (w Widget) at(x, y, cols int) (string, tview.Style, bool) {
	if w.kitty != 0 {
		fg, msb := w.id()
		return string([]rune{placeholder, diacritics[y], diacritics[x], diacritics[msb]}), tcell.StyleDefault.Foreground(color.PaletteColor(fg)), true
	}

	b := w.src.Bounds()
	height := w.pixelHeight(cols)
	pixel := func(y int) color.Color {
		if y >= height {
			return color.Default
		}
		c := stdcolor.NRGBAModel.Convert(w.src.At(b.Min.X+x*b.Dx()/cols, b.Min.Y+y*b.Dy()/height)).(stdcolor.NRGBA)
		if c.A < minAlpha {
			return color.Default
		}
		return color.NewRGBColor(int32(c.R), int32(c.G), int32(c.B))
	}

	top, bottom := pixel(2*y), pixel(2*y+1)
	switch {
	case top != color.Default:
		return upperHalfBlock, tcell.StyleDefault.Foreground(top).Background(bottom), true
	case bottom != color.Default:
		return lowerHalfBlock, tcell.StyleDefault.Foreground(bottom), true
	}
	return "", tcell.StyleDefault, false
}

// Handle passes msg through unchanged.
func (Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	return msg
}
