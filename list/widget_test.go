package list

import (
	"github.com/ayn2op/tview/layout"
	"strconv"
	"testing"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/scrollbar"
	"github.com/gdamore/tcell/v3"
)

// row is an item one line tall showing its label.
type row string

func (row) Size() (width, height layout.Length) { return layout.Fill, layout.Fixed(1) }

func (row) Layout(limits layout.Limits) layout.Size {
	return layout.Atomic(limits, layout.Fill, layout.Fixed(1))
}

func (r row) Draw(screen tview.Screen, area tview.Rectangle) {
	tview.Print(screen, string(r), area.X, area.Y, area.Width, tview.AlignmentLeft, tcell.StyleDefault)
}

func (row) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg { return msg }

// numbers returns a list of count rows labeled 0, 1, and so on.
func numbers(selectionState SelectionState, count int) Widget {
	return New(selectionState, count, func(i int) tview.Widget { return row(strconv.Itoa(i)) }).
		ScrollBar(scrollbar.New().BeginSymbol("").EndSymbol(""), ScrollBarVisibilityAutomatic).
		Focused(true).
		OnChange(func(a Change) tview.Msg { return a })
}

// apply sends msgs through a list of count rows in area, applying each Change.
func apply(t *testing.T, selectionState SelectionState, count int, area tview.Rectangle, msgs ...tview.Msg) SelectionState {
	t.Helper()
	for _, msg := range msgs {
		out := numbers(selectionState, count).Handle(msg, area)
		change, ok := out.(Change)
		if !ok {
			t.Fatalf("%v: got %v, want a Change", msg, out)
		}
		selectionState.Apply(change)
	}
	return selectionState
}

func key(k tcell.Key) tview.KeyMsg { return tcell.NewEventKey(k, "", tcell.ModNone) }

func mouse(x, y int, action tview.MouseAction) tview.MouseMsg {
	return tview.MouseMsg{EventMouse: tcell.NewEventMouse(x, y, tcell.ButtonNone, tcell.ModNone), Action: action}
}

func TestWidgetHandle(t *testing.T) {
	area := tview.Rectangle{Width: 4, Height: 4}
	t.Run("select down centers the cursor", func(t *testing.T) {
		got := apply(t, NewSelectionState(), 10, area, key(tcell.KeyDown), key(tcell.KeyDown), key(tcell.KeyDown), key(tcell.KeyDown), key(tcell.KeyDown))
		if got.cursor != 4 || got.offset != 2 {
			t.Fatalf("cursor %d offset %d, want 4 and 2", got.cursor, got.offset)
		}
	})
	t.Run("actions do what their keys do", func(t *testing.T) {
		got := apply(t, NewSelectionState(), 10, area, ActionMsg(ActionSelectBottom), ActionMsg(ActionScrollTop))
		if got.cursor != 9 || got.offset != 0 {
			t.Fatalf("cursor %d offset %d, want 9 and 0", got.cursor, got.offset)
		}
	})
	t.Run("wheel scrolls without moving the cursor", func(t *testing.T) {
		got := apply(t, NewSelectionState(), 10, area, mouse(0, 0, tview.MouseScrollDown))
		if got.cursor != -1 || got.offset != 1 {
			t.Fatalf("cursor %d offset %d, want -1 and 1", got.cursor, got.offset)
		}
	})
	t.Run("click selects the item", func(t *testing.T) {
		got := apply(t, NewSelectionState(), 10, area, mouse(0, 2, tview.MouseLeftDown))
		if got.cursor != 2 {
			t.Fatalf("cursor = %d, want 2", got.cursor)
		}
	})
	t.Run("track click pages", func(t *testing.T) {
		got := apply(t, NewSelectionState(), 10, area, mouse(3, 3, tview.MouseLeftDown))
		if got.offset != 4 {
			t.Fatalf("offset = %d, want 4", got.offset)
		}
	})
	t.Run("dragging the thumb scrolls", func(t *testing.T) {
		got := apply(t, NewSelectionState(), 10, area, mouse(3, 0, tview.MouseLeftDown), mouse(3, 3, tview.MouseMove), mouse(5, 3, tview.MouseLeftUp))
		if got.offset != 6 || got.dragging {
			t.Fatalf("offset %d dragging %v, want 6 and not dragging", got.offset, got.dragging)
		}
	})
	t.Run("holding the thumb still keeps the scroll position", func(t *testing.T) {
		wheel := mouse(0, 0, tview.MouseScrollDown)
		got := apply(t, NewSelectionState(), 100, area, wheel, wheel, wheel, mouse(3, 0, tview.MouseLeftDown), mouse(9, 0, tview.MouseMove))
		if got.offset != 3 {
			t.Fatalf("offset = %d, want 3", got.offset)
		}
	})
	t.Run("a scroll bar too short to draw takes no clicks", func(t *testing.T) {
		short := tview.Rectangle{Width: 4, Height: 2}
		msg := numbers(NewSelectionState(), 10).ScrollBar(scrollbar.New(), ScrollBarVisibilityAutomatic).Handle(mouse(3, 1, tview.MouseLeftDown), short)
		if a, ok := msg.(Change); !ok || a.offset != 0 {
			t.Fatalf("got %v, want offset 0", msg)
		}
	})
	t.Run("the end symbol scrolls a step", func(t *testing.T) {
		selectionState := NewSelectionState()
		msg := numbers(selectionState, 10).ScrollBar(scrollbar.New(), ScrollBarVisibilityAutomatic).Handle(mouse(3, 3, tview.MouseLeftDown), area)
		if a, ok := msg.(Change); !ok || a.offset != 1 {
			t.Fatalf("got %v, want offset 1", msg)
		}
	})
	t.Run("passes other keys through", func(t *testing.T) {
		selectionState := NewSelectionState()
		enter := key(tcell.KeyEnter)
		if got := numbers(selectionState, 10).Handle(enter, area); got != enter {
			t.Fatalf("got %v", got)
		}
	})
}

func TestWidgetDraw(t *testing.T) {
	area := tview.Rectangle{Width: 3, Height: 2}
	t.Run("centers the cursor with the selected style", func(t *testing.T) {
		screen := screentest.New(t, 3, 2)
		selectionState := NewSelectionState()
		selectionState.SetCursor(1)
		numbers(selectionState, 3).ScrollBar(scrollbar.New(), ScrollBarVisibilityNever).SelectedStyle(tcell.StyleDefault.Reverse(true)).Draw(screen, area)
		if got := screentest.Row(screen, 0, 3) + screentest.Row(screen, 1, 3); got != "0  1  " {
			t.Fatalf("rows = %q", got)
		}
		if _, style, _ := screen.Get(0, 1); !style.HasReverse() {
			t.Fatal("selected item is not reversed")
		}
	})
	t.Run("deselecting leaves the view where it is", func(t *testing.T) {
		screen := screentest.New(t, 3, 2)
		selectionState := NewSelectionState()
		selectionState.SetCursor(3)
		selectionState.SetCursor(-1)
		numbers(selectionState, 9).ScrollBar(scrollbar.New(), ScrollBarVisibilityNever).Draw(screen, area)
		if got := screentest.Row(screen, 0, 3) + screentest.Row(screen, 1, 3); got != "2  3  " {
			t.Fatalf("rows = %q", got)
		}
	})
	t.Run("follows the end", func(t *testing.T) {
		screen := screentest.New(t, 3, 2)
		selectionState := NewSelectionState()
		selectionState.ScrollToEnd()
		numbers(selectionState, 5).Draw(screen, area)
		if got := screentest.Row(screen, 1, 2); got != "4 " {
			t.Fatalf("last row = %q", got)
		}
	})
	t.Run("clips fills of items partly out of view", func(t *testing.T) {
		screen := screentest.New(t, 3, 3)
		selectionState := NewSelectionState()
		selectionState.SetCursor(-1)
		New(selectionState, 1, func(int) tview.Widget { return filled{} }).ScrollBar(scrollbar.New(), ScrollBarVisibilityNever).Draw(screen, tview.Rectangle{Y: 1, Width: 3, Height: 1})
		if got := screentest.Row(screen, 0, 3) + screentest.Row(screen, 1, 3) + screentest.Row(screen, 2, 3); got != "   xxx   " {
			t.Fatalf("rows = %q", got)
		}
	})
	t.Run("lays items that fit out at one width", func(t *testing.T) {
		widths := map[int]bool{}
		w := New(NewSelectionState(), 1, func(int) tview.Widget { return measured{widths: widths} }).OnChange(func(c Change) tview.Msg { return c })
		w.Draw(screentest.New(t, 3, 2), area)
		w.Handle(mouse(0, 0, tview.MouseMove), area)
		if len(widths) != 1 || !widths[2] {
			t.Fatalf("widths = %v", widths)
		}
	})
}

// measured is an item that records the widths it is laid out and drawn at.
type measured struct {
	row
	widths map[int]bool
}

func (m measured) Layout(limits layout.Limits) layout.Size {
	m.widths[limits.Max.Width] = true
	return m.row.Layout(limits)
}

func (m measured) Draw(screen tview.Screen, area tview.Rectangle) {
	m.widths[area.Width] = true
}

// filled is an item that fills more than the area it is given.
type filled struct{ row }

func (filled) Draw(screen tview.Screen, area tview.Rectangle) {
	screen.FillArea(area.X-1, area.Y-1, area.Width+2, area.Height+2, 'x', tcell.StyleDefault)
}

func (filled) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg { return msg }

func TestWidgetTrackEnd(t *testing.T) {
	area := tview.Rectangle{Width: 3, Height: 2}
	// scrolled returns the state scrolled to the end of 5 rows, with msgs applied by a list that tracks the end.
	scrolled := func(t *testing.T, msgs ...tview.Msg) SelectionState {
		t.Helper()
		selectionState := NewSelectionState()
		selectionState.ScrollToEnd()
		for _, msg := range msgs {
			selectionState.Apply(numbers(selectionState, 5).TrackEnd(true).Handle(msg, area).(Change))
		}
		return selectionState
	}
	lastRow := func(t *testing.T, selectionState SelectionState, count int) string {
		t.Helper()
		screen := screentest.New(t, 3, 2)
		numbers(selectionState, count).ScrollBar(scrollbar.New(), ScrollBarVisibilityNever).Draw(screen, area)
		return screentest.Row(screen, 1, 1)
	}
	t.Run("stays at the end when an item is added and the cursor is kept", func(t *testing.T) {
		selectionState := scrolled(t)
		selectionState.SetCursor(-1)
		if got := lastRow(t, selectionState, 6); got != "5" {
			t.Fatalf("last row = %q, want %q", got, "5")
		}
	})
	t.Run("scrolling back to the end tracks it again", func(t *testing.T) {
		selectionState := scrolled(t, mouse(0, 0, tview.MouseScrollUp), mouse(0, 0, tview.MouseScrollDown))
		if got := lastRow(t, selectionState, 6); got != "5" {
			t.Fatalf("last row = %q, want %q", got, "5")
		}
	})
	t.Run("scrolled up stays put", func(t *testing.T) {
		selectionState := scrolled(t, mouse(0, 0, tview.MouseScrollUp))
		if got := lastRow(t, selectionState, 6); got != "3" {
			t.Fatalf("last row = %q, want %q", got, "3")
		}
	})
}

// wrapped is an item that takes a row for every width cells of its 6 cells of text.
type wrapped struct{ row }

func (wrapped) Layout(limits layout.Limits) layout.Size {
	return layout.Size{Width: limits.Max.Width, Height: (6 + limits.Max.Width - 1) / limits.Max.Width}
}

// sized is an item of a height.
type sized struct {
	row
	height layout.Length
}

func (s sized) Layout(limits layout.Limits) layout.Size {
	return layout.Atomic(limits, layout.Fill, s.height)
}

func TestZeroSelectionStateIsNotDragging(t *testing.T) {
	var selectionState SelectionState
	move := mouse(1, 3, tview.MouseMove)
	if got := numbers(selectionState, 50).Handle(move, tview.Rectangle{Width: 4, Height: 5}); got != move {
		t.Fatalf("moving the mouse over the list produced %v", got)
	}
}

func TestItemHeight(t *testing.T) {
	for _, tt := range []struct {
		name string
		item tview.Widget
		want int
	}{
		{"one row", row(""), 1},
		{"height for the width", wrapped{}, 2},
		{"fixed height", sized{height: layout.Fixed(3)}, 3},
		{"fill height takes the height of the view", sized{height: layout.Fill}, 10},
	} {
		t.Run(tt.name, func(t *testing.T) {
			list := New(NewSelectionState(), 1, func(int) tview.Widget { return tt.item })
			if _, sizes, _ := list.layout(layout.Size{Width: 4, Height: 10}); sizes[0] != tt.want {
				t.Fatalf("rows = %d, want %d", sizes[0], tt.want)
			}
		})
	}
}
