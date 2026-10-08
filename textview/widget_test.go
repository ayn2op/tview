package textview

import (
	"github.com/ayn2op/tview/layout"
	"testing"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/richtext"
	"github.com/gdamore/tcell/v3"
)

func lines(strs ...string) richtext.Text {
	text := make(richtext.Text, len(strs))
	for i, s := range strs {
		text[i] = richtext.NewLine(richtext.NewSegment(s, tcell.StyleDefault))
	}
	return text
}

func TestWidgetDraw(t *testing.T) {
	t.Run("wraps on words", func(t *testing.T) {
		screen := screentest.New(t, 6, 2)
		New(lines("ab cd ef")).Draw(screen, tview.Rectangle{Width: 6, Height: 2})
		for y, want := range []string{"ab cd ", "ef    "} {
			if got := screentest.Row(screen, y, 6); got != want {
				t.Fatalf("row %d = %q, want %q", y, got, want)
			}
		}
	})
	t.Run("centers lines", func(t *testing.T) {
		screen := screentest.New(t, 6, 1)
		New(lines("ab")).Alignment(tview.AlignmentCenter).Draw(screen, tview.Rectangle{Width: 6, Height: 1})
		if got := screentest.Row(screen, 0, 6); got != "  ab  " {
			t.Fatalf("row = %q", got)
		}
	})
}

func TestWidgetLayout(t *testing.T) {
	limits := layout.Limits{Max: layout.Size{Width: 6, Height: 10}}
	if got := New(lines("ab cd ef", "g")).Height(layout.Shrink).Layout(limits); got != (layout.Size{Width: 6, Height: 3}) {
		t.Fatalf("size = %+v, want 6 by 3", got)
	}
}
