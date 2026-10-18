package image

import (
	"image"
	stdcolor "image/color"
	"testing"

	"github.com/ayn2op/tview/layout"

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
	t.Run("cell size", func(t *testing.T) {
		src := image.NewRGBA(image.Rect(0, 0, 100, 100))
		width, height := New(src).Width(10).CellSize(10, 25).Size()
		if width != layout.Fixed(10) || height != layout.Fixed(4) {
			t.Fatalf("size = %v, %v", width, height)
		}
	})
}

func TestWidgetDraw(t *testing.T) {
	screen := screentest.New(t, 1, 2)
	New(column(red, blue, faint, translucent)).Draw(screen, tview.Rectangle{Width: 1, Height: 2})
	for y, want := range []struct {
		str   string
		style tview.Style
	}{
		{"▀", tcell.StyleDefault.Foreground(color.FromImageColor(red)).Background(color.FromImageColor(blue))},
		{"▄", tcell.StyleDefault.Foreground(color.NewRGBColor(0, 255, 0))},
	} {
		if str, style, _ := screen.Get(0, y); str != want.str || style != want.style {
			t.Fatalf("row %d = %q %v, want %q %v", y, str, style, want.str, want.style)
		}
	}
}

func TestWidgetDrawOverBackground(t *testing.T) {
	screen := screentest.New(t, 1, 1)
	white := color.NewRGBColor(255, 255, 255)
	screen.Put(0, 0, " ", tcell.StyleDefault.Background(white))
	New(column(red)).Draw(screen, tview.Rectangle{Width: 1, Height: 1})
	want := tcell.StyleDefault.Foreground(color.FromImageColor(red)).Background(white)
	if str, style, _ := screen.Get(0, 0); str != "▀" || style != want {
		t.Fatalf("cell = %q %v, want the transparent bottom half to keep the white beneath", str, style)
	}
}

func TestWidgetLines(t *testing.T) {
	t.Run("half blocks", func(t *testing.T) {
		// A pixel above a transparent one, then a transparent cell.
		lines := New(column(red, faint, faint, faint)).Lines()
		if len(lines) != 2 || len(lines[0]) != 1 || len(lines[1]) != 1 {
			t.Fatalf("lines = %v, want two of one cell", lines)
		}
		if got := lines[0][0]; got.Text != "▀" || got.Style != tcell.StyleDefault.Foreground(color.FromImageColor(red)) {
			t.Fatalf("first cell = %+v", got)
		}
		if got := lines[1][0]; got.Text != " " {
			t.Fatalf("transparent cell = %q, want a space", got.Text)
		}
	})
	t.Run("kitty", func(t *testing.T) {
		lines := New(column(red)).Width(2).Kitty(257).Lines()
		want := string([]rune{placeholder, diacritics[0], diacritics[1], diacritics[1]})
		if len(lines) != 1 || len(lines[0]) != 2 || lines[0][1].Text != want || lines[0][1].Style.GetForeground() != color.PaletteColor(16) {
			t.Fatalf("lines = %+v, want the placeholders of image 257", lines)
		}
	})
}
