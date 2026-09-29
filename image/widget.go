// Package image draws images with half blocks, two pixels per cell.
package image

import (
	"image"
	stdcolor "image/color"

	"github.com/ayn2op/tview"
	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

// minAlpha is the least opacity drawn, so the faint edges of transparent images do not show as dark cells.
const minAlpha = 50

// Widget draws an image scaled to fit its area, keeping the aspect ratio.
type Widget struct {
	src   image.Image
	width int
}

var _ tview.Element = Widget{}

// New draws src at one pixel per column.
func New(src image.Image) Widget {
	return Widget{src: src}
}

// Width sets the width in cells, which the height follows.
func (w Widget) Width(cells int) Widget {
	w.width = cells
	return w
}

// Size returns the width and the height that keeps the aspect ratio.
func (w Widget) Size() (width, height tview.Length) {
	b := w.src.Bounds()
	if b.Empty() {
		return tview.Fixed(0), tview.Fixed(0)
	}
	cells := w.width
	if cells == 0 {
		cells = b.Dx()
	}
	return tview.Fixed(cells), tview.Fixed((cells*b.Dy()/b.Dx() + 1) / 2)
}

// Draw draws the image from the top-left corner of area. Transparent pixels leave the cell beneath.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	b := w.src.Bounds()
	if b.Empty() {
		return
	}
	width := min(area.Width, area.Height*2*b.Dx()/b.Dy())
	if width <= 0 {
		return
	}
	height := width * b.Dy() / b.Dx()
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
				screen.Put(area.X+x, area.Y+y/2, "▀", tcell.StyleDefault.Foreground(top).Background(bottom))
			case bottom != color.Default:
				screen.Put(area.X+x, area.Y+y/2, "▄", tcell.StyleDefault.Foreground(bottom))
			}
		}
	}
}

// Handle passes msg through unchanged.
func (Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	return msg
}
