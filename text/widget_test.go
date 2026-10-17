package text

import (
	"testing"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/layout"
	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

func TestWidgetDraw(t *testing.T) {
	t.Run("keeps the style beneath", func(t *testing.T) {
		screen := screentest.New(t, 5, 1)
		style := tcell.StyleDefault.Background(color.Blue)
		screen.Put(1, 0, " ", style)
		New("ab").Draw(screen, tview.Rectangle{Width: 5, Height: 1})
		if _, got, _ := screen.Get(1, 0); got != style {
			t.Fatalf("style = %v, want %v", got, style)
		}
	})
	t.Run("wraps between words", func(t *testing.T) {
		screen := screentest.New(t, 6, 2)
		New("ab cd ef").Draw(screen, tview.Rectangle{Width: 6, Height: 2})
		for y, want := range []string{"ab cd ", "ef    "} {
			if got := screentest.Row(screen, y, 6); got != want {
				t.Fatalf("row %d = %q, want %q", y, got, want)
			}
		}
	})
	t.Run("centers lines", func(t *testing.T) {
		screen := screentest.New(t, 6, 1)
		New("ab").AlignX(tview.AlignmentCenter).Draw(screen, tview.Rectangle{Width: 6, Height: 1})
		if got := screentest.Row(screen, 0, 6); got != "  ab  " {
			t.Fatalf("row = %q", got)
		}
	})
	t.Run("ends lines wider than area at its right edge", func(t *testing.T) {
		screen := screentest.New(t, 4, 2)
		New("abcdef\nab").Wrapping(WrappingNone).AlignX(tview.AlignmentRight).Draw(screen, tview.Rectangle{Width: 4, Height: 2})
		for y, want := range []string{"cdef", "  ab"} {
			if got := screentest.Row(screen, y, 4); got != want {
				t.Fatalf("row %d = %q, want %q", y, got, want)
			}
		}
	})
}

func TestWidgetLayout(t *testing.T) {
	limits := layout.Limits{Max: layout.Size{Width: 6, Height: 10}}
	if got := New("ab cd ef\ng").Layout(limits); got != (layout.Size{Width: 5, Height: 3}) {
		t.Fatalf("size = %+v, want 5 by 3", got)
	}
}
