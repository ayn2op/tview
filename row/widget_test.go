package row

import (
	"fmt"
	"testing"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/layout"
	"github.com/ayn2op/tview/text"
)

func TestWidgetDraw(t *testing.T) {
	screen := screentest.New(t, 7, 1)
	New(text.New("ab"), nil, text.New("cd")).Spacing(1).Draw(screen, tview.Rectangle{Width: 7, Height: 1})
	if got := screentest.Row(screen, 0, 7); got != "ab cd  " {
		t.Fatalf("row = %q", got)
	}
}

func TestWidgetLayout(t *testing.T) {
	limits := layout.Limits{Max: layout.Size{Width: 10, Height: 10}}
	got := New(text.New("ab"), text.New("cde")).Width(layout.Shrink).Height(layout.Shrink).Spacing(1).Layout(limits)
	if got != (layout.Size{Width: 6, Height: 1}) {
		t.Fatalf("size = %+v", got)
	}
}

func TestWidgetWrap(t *testing.T) {
	screen := screentest.New(t, 5, 3)
	New(text.New("ab"), text.New("cd"), text.New("e")).Spacing(1).Wrap().VerticalSpacing(0).Draw(screen, tview.Rectangle{Width: 5, Height: 3})
	for y, want := range []string{"ab cd", "e    ", "     "} {
		t.Run(fmt.Sprintf("row %d", y), func(t *testing.T) {
			if got := screentest.Row(screen, y, 5); got != want {
				t.Fatalf("got %q, want %q", got, want)
			}
		})
	}

	t.Run("layout", func(t *testing.T) {
		limits := layout.Limits{Max: layout.Size{Width: 4, Height: 10}}
		if got := New(text.New("ab"), text.New("cd")).Spacing(1).Height(layout.Shrink).Wrap().Layout(limits); got != (layout.Size{Width: 4, Height: 3}) {
			t.Fatalf("size = %+v, want two lines one apart", got)
		}
	})
	t.Run("infinite height", func(t *testing.T) {
		// A list measures its items with an infinite height and no height to hint at.
		limits := layout.Limits{Max: layout.Size{Width: 4}, Infinite: layout.Axes{Height: true}}
		if got := New(text.New("ab"), text.New("cd")).Spacing(1).Height(layout.Shrink).Wrap().Layout(limits); got.Height != 3 {
			t.Fatalf("size = %+v, want three rows within an infinite height", got)
		}
	})
}

// drawRows draws widget in an area of width and height and returns its rows.
func drawRows(t *testing.T, widget tview.Widget, width, height int) []string {
	t.Helper()
	screen := screentest.New(t, width, height)
	widget.Draw(screen, tview.Rectangle{Width: width, Height: height})
	rows := make([]string, height)
	for y := range rows {
		rows[y] = screentest.Row(screen, y, width)
	}
	return rows
}

func TestWrappingAlignX(t *testing.T) {
	got := drawRows(t, New(text.New("ab"), text.New("cd"), text.New("e")).Spacing(1).Wrap().VerticalSpacing(0).AlignX(tview.HorizontalAlignmentCenter), 5, 2)
	if got[1] != "  e  " {
		t.Fatalf("rows = %q, want the second line centered within the first", got)
	}
}

func TestWidgetAlignY(t *testing.T) {
	got := drawRows(t, New(text.New("x\ny"), text.New("a")).Spacing(1).AlignY(tview.VerticalAlignmentBottom).Wrap(), 3, 2)
	if got[0] != "x  " || got[1] != "y a" {
		t.Fatalf("rows = %q, want the shorter child at the bottom of its line", got)
	}
}
