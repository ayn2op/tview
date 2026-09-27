package scrollbar

import (
	"testing"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
)

func column(t *testing.T, w Widget, height int) string {
	t.Helper()
	screen := screentest.New(t, 1, height)
	w.Draw(screen, tview.Rectangle{Width: 1, Height: height})
	var s string
	for y := range height {
		s += screentest.Row(screen, y, 1)
	}
	return s
}

func TestWidgetDraw(t *testing.T) {
	glyphs := LegacyComputingGlyphSet()
	t.Run("thumb at the top", func(t *testing.T) {
		if got := column(t, New().GlyphSet(glyphs).Lengths(8, 2), 4); got != "█│││" {
			t.Fatalf("column = %q", got)
		}
	})
	t.Run("thumb at the bottom with arrows", func(t *testing.T) {
		if got := column(t, New().GlyphSet(glyphs).Arrows(ArrowsBoth).Lengths(8, 2).Offset(6), 4); got != "▲│█▼" {
			t.Fatalf("column = %q", got)
		}
	})
	t.Run("hidden when everything fits", func(t *testing.T) {
		if got := column(t, New().GlyphSet(glyphs).Lengths(2, 4), 4); got != "    " {
			t.Fatalf("column = %q", got)
		}
	})
}

func TestWidgetThumb(t *testing.T) {
	start, size := New().Lengths(8, 2).Offset(3).Thumb(4)
	if start != 12 || size != 8 {
		t.Fatalf("thumb = %d+%d, want 12+8", start, size)
	}
}
