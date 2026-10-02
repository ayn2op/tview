package image

import (
	"github.com/ayn2op/tview/layout"
	"image"
	stdcolor "image/color"
	"testing"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

// column returns a one pixel wide image of pixels from top to bottom.
func column(pixels ...stdcolor.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 1, len(pixels)))
	for y, c := range pixels {
		img.Set(0, y, c)
	}
	return img
}

var (
	red         = stdcolor.RGBA{R: 255, A: 255}
	blue        = stdcolor.RGBA{B: 255, A: 255}
	faint       = stdcolor.NRGBA{R: 255, A: minAlpha - 1}
	translucent = stdcolor.NRGBA{G: 255, A: 128}
)

func TestWidgetSize(t *testing.T) {
	t.Run("one pixel per column", func(t *testing.T) {
		width, height := New(column(red, red, red)).Size()
		if width != layout.Fixed(1) || height != layout.Fixed(2) {
			t.Fatalf("size = %v, %v", width, height)
		}
	})
	t.Run("height follows width", func(t *testing.T) {
		width, height := New(column(red, red)).Width(3).Size()
		if width != layout.Fixed(3) || height != layout.Fixed(3) {
			t.Fatalf("size = %v, %v", width, height)
		}
	})
}

func TestWidgetDraw(t *testing.T) {
	screen := screentest.New(t, 1, 2)
	New(column(red, blue, faint, translucent)).Draw(screen, tview.Rectangle{Width: 1, Height: 2})
	for y, want := range []struct {
		str   string
		style tcell.Style
	}{
		{"▀", tcell.StyleDefault.Foreground(color.FromImageColor(red)).Background(color.FromImageColor(blue))},
		{"▄", tcell.StyleDefault.Foreground(color.NewRGBColor(0, 255, 0))},
	} {
		if str, style, _ := screen.Get(0, y); str != want.str || style != want.style {
			t.Fatalf("row %d = %q %v, want %q %v", y, str, style, want.str, want.style)
		}
	}
}
