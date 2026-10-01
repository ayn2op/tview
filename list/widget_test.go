package list

import (
	"strconv"
	"testing"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/scrollbar"
	"github.com/gdamore/tcell/v3"
)

// row is an item one line tall showing its label.
type row string

func (r row) Rows(int) int { return 1 }

func (r row) Draw(screen tview.Screen, area tview.Rectangle) {
	tview.Print(screen, string(r), area.X, area.Y, area.Width, tview.AlignmentLeft, tcell.StyleDefault)
}

func (row) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg { return msg }

// numbers returns a list of count rows labeled 0, 1, and so on.
func numbers(selectionState *SelectionState, count int) Widget {
	return New(selectionState, count, func(i int) Item { return row(strconv.Itoa(i)) }).
		Focused(true).
		OnChange(func(a Change) tview.Msg { return a })
}

// apply sends msgs through a list of count rows in area, applying each Change.
func apply(t *testing.T, selectionState SelectionState, count int, area tview.Rectangle, msgs ...tview.Msg) SelectionState {
	t.Helper()
	for _, msg := range msgs {
		out := numbers(&selectionState, count).Handle(msg, area)
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
		got := apply(t, NewSelectionState(), 10, area, mouse(0, 2, tview.MouseLeftClick))
		if got.cursor != 2 {
			t.Fatalf("cursor = %d, want 2", got.cursor)
		}
	})
	t.Run("track click pages", func(t *testing.T) {
		got := apply(t, NewSelectionState(), 10, area, mouse(3, 3, tview.MouseLeftClick))
		if got.offset != 4 {
			t.Fatalf("offset = %d, want 4", got.offset)
		}
	})
	t.Run("dragging the thumb scrolls", func(t *testing.T) {
		got := apply(t, NewSelectionState(), 10, area, mouse(3, 0, tview.MouseLeftDown), mouse(3, 3, tview.MouseMove), mouse(5, 3, tview.MouseLeftUp))
		if got.offset != 6 || got.grab != -1 {
			t.Fatalf("offset %d grab %d, want 6 and -1", got.offset, got.grab)
		}
	})
	t.Run("arrows scroll a step", func(t *testing.T) {
		selectionState := NewSelectionState()
		bar := scrollbar.New().Arrows(scrollbar.ArrowsBoth)
		msg := numbers(&selectionState, 10).ScrollBar(bar, ScrollBarVisibilityAutomatic).Handle(mouse(3, 3, tview.MouseLeftClick), area)
		if a, ok := msg.(Change); !ok || a.offset != 1 {
			t.Fatalf("got %v, want offset 1", msg)
		}
	})
	t.Run("passes other keys through", func(t *testing.T) {
		selectionState := NewSelectionState()
		enter := key(tcell.KeyEnter)
		if got := numbers(&selectionState, 10).Handle(enter, area); got != enter {
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
		numbers(&selectionState, 3).ScrollBar(scrollbar.New(), ScrollBarVisibilityNever).SelectedStyle(tcell.StyleDefault.Reverse(true)).Draw(screen, area)
		if got := screentest.Row(screen, 0, 3) + screentest.Row(screen, 1, 3); got != "0  1  " {
			t.Fatalf("rows = %q", got)
		}
		if _, style, _ := screen.Get(0, 1); !style.HasReverse() {
			t.Fatal("selected item is not reversed")
		}
	})
	t.Run("follows the end", func(t *testing.T) {
		screen := screentest.New(t, 3, 2)
		selectionState := NewSelectionState()
		selectionState.ScrollToEnd()
		numbers(&selectionState, 5).Draw(screen, area)
		if got := screentest.Row(screen, 1, 2); got != "4 " {
			t.Fatalf("last row = %q", got)
		}
	})
	t.Run("clips fills of items partly out of view", func(t *testing.T) {
		screen := screentest.New(t, 3, 3)
		selectionState := NewSelectionState()
		selectionState.SetCursor(-1)
		New(&selectionState, 1, func(int) Item { return filled{} }).Draw(screen, tview.Rectangle{Y: 1, Width: 3, Height: 1})
		if got := screentest.Row(screen, 0, 3) + screentest.Row(screen, 1, 3) + screentest.Row(screen, 2, 3); got != "   xxx   " {
			t.Fatalf("rows = %q", got)
		}
	})
}

// filled is an item that fills more than the area it is given.
type filled struct{}

func (filled) Rows(int) int { return 1 }

func (filled) Draw(screen tview.Screen, area tview.Rectangle) {
	screen.FillArea(area.X-1, area.Y-1, area.Width+2, area.Height+2, 'x', tcell.StyleDefault)
}

func (filled) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg { return msg }

func TestStateSetTrackEnd(t *testing.T) {
	area := tview.Rectangle{Width: 3, Height: 2}
	lastRow := func(t *testing.T, selectionState SelectionState, count int) string {
		t.Helper()
		screen := screentest.New(t, 3, 2)
		numbers(&selectionState, count).ScrollBar(scrollbar.New(), ScrollBarVisibilityNever).Draw(screen, area)
		return screentest.Row(screen, 1, 1)
	}
	t.Run("stays at the end when an item is added and the cursor is kept", func(t *testing.T) {
		selectionState := NewSelectionState()
		selectionState.SetTrackEnd(true)
		selectionState.ScrollToEnd()
		selectionState.SetCursor(-1)
		if got := lastRow(t, selectionState, 6); got != "5" {
			t.Fatalf("last row = %q, want %q", got, "5")
		}
	})
	t.Run("scrolling back to the end tracks it again", func(t *testing.T) {
		selectionState := NewSelectionState()
		selectionState.SetTrackEnd(true)
		selectionState.ScrollToEnd()
		selectionState = apply(t, selectionState, 5, area, mouse(0, 0, tview.MouseScrollUp), mouse(0, 0, tview.MouseScrollDown))
		if got := lastRow(t, selectionState, 6); got != "5" {
			t.Fatalf("last row = %q, want %q", got, "5")
		}
	})
	t.Run("scrolled up stays put", func(t *testing.T) {
		selectionState := NewSelectionState()
		selectionState.SetTrackEnd(true)
		selectionState.ScrollToEnd()
		selectionState = apply(t, selectionState, 5, area, mouse(0, 0, tview.MouseScrollUp))
		if got := lastRow(t, selectionState, 6); got != "3" {
			t.Fatalf("last row = %q, want %q", got, "3")
		}
	})
}
