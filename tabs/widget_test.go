package tabs

import (
	"testing"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/layout"
	"github.com/gdamore/tcell/v3"
)

type selectMsg int

func TestWidgetHandle(t *testing.T) {
	area := tview.Rectangle{X: 1, Y: 2, Width: 10, Height: 5}
	tabs := New("tab", "tab").
		OnSelect(func(i int) tview.Msg { return selectMsg(i) })
	next := tcell.NewEventKey(tcell.KeyRune, "l", tcell.ModCtrl)
	// The labels "tab tab" are centered: the second starts at x 1 + (10-7)/2 + 4.
	click := tview.MouseMsg{EventMouse: tcell.NewEventMouse(6, 2, tcell.ButtonNone, tcell.ModNone), Action: tview.MouseLeftDown}

	for _, tt := range []struct {
		name   string
		widget Widget
		msg    tview.Msg
		want   tview.Msg
	}{
		{"next", tabs, next, selectMsg(1)},
		{"next at the last tab passes through", tabs.Active(1), next, next},
		{"next at the last tab wraps to the first", tabs.Active(1).Wrap(true), next, selectMsg(0)},
		{"clicking a label", tabs, click, selectMsg(1)},
		{"clicking the active label", tabs.Active(1), click, nil},
		{"scrolling over the space between labels", tabs, tview.MouseMsg{EventMouse: tcell.NewEventMouse(5, 2, tcell.ButtonNone, tcell.ModNone), Action: tview.MouseScrollDown}, selectMsg(1)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.widget.Handle(tt.msg, area); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("clicking an arrow selects the neighbor of the active tab", func(t *testing.T) {
		// "one two three" is wider than the area, so it scrolls between arrows at x 1 and 10.
		arrow := tview.MouseMsg{EventMouse: tcell.NewEventMouse(1, 2, tcell.ButtonNone, tcell.ModNone), Action: tview.MouseLeftDown}
		w := New("one", "two", "three").Arrows("◀", "▶").Active(2).OnSelect(func(i int) tview.Msg { return selectMsg(i) })
		if got := w.Handle(arrow, area); got != selectMsg(1) {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("clicking an arrow that is not clickable does nothing", func(t *testing.T) {
		arrow := tview.MouseMsg{EventMouse: tcell.NewEventMouse(1, 2, tcell.ButtonNone, tcell.ModNone), Action: tview.MouseLeftDown}
		w := New("one", "two", "three").Arrows("◀", "▶").ClickableArrows(false).Active(2).OnSelect(func(i int) tview.Msg { return selectMsg(i) })
		if got := w.Handle(arrow, area); got != nil {
			t.Fatalf("got %v", got)
		}
	})
}

func TestWidgetDraw(t *testing.T) {
	tabs := New("one", "two", "three")
	for _, tt := range []struct {
		name   string
		widget Widget
		width  int
		want   string
	}{
		{"labels that fit are centered", tabs, 15, " one two three "},
		{"the first tab stays at the start", tabs.Arrows("◀", "▶"), 9, " one two▶"},
		{"the last tab stays at the end", tabs.Arrows("◀", "▶").Active(2), 9, "◀o three "},
		{"the active tab is centered", tabs.Arrows("◀", "▶").Active(1), 9, "◀e two t▶"},
		{"the active tab is centered without arrows", tabs.Active(1), 7, "e two t"},
		{"labels aligned left with padding and a separator", New("a", "b").Alignment(tview.HorizontalAlignmentLeft).Padding("[", "]").Separator("|"), 9, "[a]|[b]  "},
	} {
		t.Run(tt.name, func(t *testing.T) {
			screen := screentest.New(t, tt.width, 1)
			tt.widget.Draw(screen, tview.Rectangle{Width: tt.width, Height: 1})
			if got := screentest.Row(screen, 0, tt.width); got != tt.want {
				t.Fatalf("row = %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("redrawing clears labels and arrows left over from the last draw", func(t *testing.T) {
		screen := screentest.New(t, 9, 1)
		area := tview.Rectangle{Width: 9, Height: 1}
		tabs.Arrows("◀", "▶").Draw(screen, area)
		tabs.Arrows("◀", "▶").Active(2).Draw(screen, area)
		if got, want := screentest.Row(screen, 0, 9), "◀o three "; got != want {
			t.Fatalf("row = %q, want %q", got, want)
		}
	})
}

func TestWidgetLayout(t *testing.T) {
	limits := layout.Limits{Max: layout.Size{Width: 20, Height: 1}}
	for name, tt := range map[string]struct {
		width layout.Length
		want  int
	}{
		"fill takes the row":            {layout.Fill, 20},
		"shrink is as wide as the tabs": {layout.Shrink, len("[a] - [bc]")},
	} {
		t.Run(name, func(t *testing.T) {
			tabs := New("a", "bc").Separator(" - ").Padding("[", "]").Width(tt.width)
			if got := tabs.Layout(limits).Width; got != tt.want {
				t.Fatalf("width = %d, want %d", got, tt.want)
			}
		})
	}
}
