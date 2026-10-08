package viewport

import (
	"slices"
	"strconv"
	"testing"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/layout"
	"github.com/ayn2op/tview/list"
	"github.com/ayn2op/tview/scrollbar"
	"github.com/ayn2op/tview/text"
	"github.com/ayn2op/tview/tree"
	"github.com/gdamore/tcell/v3"
)

func key(k tcell.Key) tview.KeyMsg { return tcell.NewEventKey(k, "", tcell.ModNone) }

func mouse(x, y int, action tview.MouseAction) tview.MouseMsg {
	return tview.MouseMsg{EventMouse: tcell.NewEventMouse(x, y, tcell.ButtonNone, tcell.ModNone), Action: action}
}

// treeModel is a tree of four rows, the third wider than the area, in a viewport that scrolls both ways.
type treeModel struct {
	root        *tree.Node
	selection   tree.SelectionState
	scrollState ScrollState
}

func newTreeModel() *treeModel {
	root := tree.NewNode("root")
	for _, name := range []string{"a", "b", "long name", "d"} {
		root.AddChild(tree.NewNode(name))
	}
	return &treeModel{root: root}
}

var treeArea = tview.Rectangle{X: 1, Y: 1, Width: 6, Height: 2}

func (m *treeModel) view() Widget {
	t := tree.New(m.root, m.selection).TopLevel(1).Markers(tree.Markers{}).Focused(true).
		OnChange(func(c tree.Change) tview.Msg { return c }).
		OnSelect(func(n *tree.Node) tview.Msg { return n })
	return New(t, m.scrollState).
		Target(t.Target).
		Axes(layout.Axes{Width: true, Height: true}).
		Focused(true).
		OnChange(func(c Change) tview.Msg { return c })
}

func (m *treeModel) send(msg tview.Msg) tview.Msg {
	switch out := m.view().Handle(msg, treeArea).(type) {
	case tree.Change:
		m.selection.Apply(out)
		m.scrollState.ScrollToTarget()
	case Change:
		m.scrollState.Apply(out)
	default:
		return out
	}
	return nil
}

func (m *treeModel) draw(t *testing.T) []string {
	screen := screentest.New(t, 8, 4)
	m.view().Draw(screen, treeArea)
	rows := make([]string, 4)
	for y := range rows {
		rows[y] = screentest.Row(screen, y, 8)
	}
	return rows
}

func TestWidget(t *testing.T) {
	blank := "        "
	check := func(t *testing.T, m *treeModel, want ...string) {
		t.Helper()
		if got := m.draw(t); !slices.Equal(got, want) {
			t.Fatalf("got %q, want %q", got, want)
		}
	}

	t.Run("draws the top of the child within the area", func(t *testing.T) {
		check(t, newTreeModel(), blank, " a      ", " b      ", blank)
	})
	t.Run("keys the child leaves scroll", func(t *testing.T) {
		m := newTreeModel()
		m.send(key(tcell.KeyPgDn))
		check(t, m, blank, " long n ", " d      ", blank)
		m.send(key(tcell.KeyRight))
		m.send(key(tcell.KeyRight))
		check(t, m, blank, " ng nam ", "        ", blank)
		// The child is 9 cells wide, so the view stops 3 cells in.
		m.send(mouse(2, 1, tview.MouseScrollRight))
		check(t, m, blank, " g name ", "        ", blank)
	})
	t.Run("the wheel scrolls within the area only", func(t *testing.T) {
		m := newTreeModel()
		m.send(mouse(2, 1, tview.MouseScrollDown))
		check(t, m, blank, " b      ", " long n ", blank)
		if out := m.send(mouse(0, 0, tview.MouseScrollDown)); out == nil {
			t.Fatal("the wheel outside the area was used")
		}
	})
	t.Run("the target is kept in the middle", func(t *testing.T) {
		m := newTreeModel()
		m.selection.SetCurrentNode(m.root.Children()[0])
		for range 3 {
			m.send(key(tcell.KeyDown))
		}
		if got := m.selection.CurrentNode(); got != m.root.Children()[3] {
			t.Fatalf("current = %v", got.Line())
		}
		check(t, m, blank, " long n ", " d      ", blank)
	})
	t.Run("clicks reach the scrolled child", func(t *testing.T) {
		m := newTreeModel()
		m.send(mouse(2, 1, tview.MouseScrollDown))
		if got := m.send(mouse(2, 2, tview.MouseLeftClick)); got != tview.Msg(m.root.Children()[2]) {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("scrolls sideways only as far as the rows in view", func(t *testing.T) {
		m := newTreeModel()
		tr := tree.New(m.root, m.selection).TopLevel(1).Markers(tree.Markers{})
		view := m.view().ContentWidth(tr.RowsWidth)
		if out := view.Handle(key(tcell.KeyRight), treeArea); out != tview.Msg(Change{}) {
			t.Fatalf("got %+v, want no scrolling past a and b", out)
		}
		m.send(key(tcell.KeyPgDn))
		if out := m.view().ContentWidth(tr.RowsWidth).Handle(key(tcell.KeyRight), treeArea); out != tview.Msg(Change{x: 1, y: 2}) {
			t.Fatalf("got %+v, want a cell right at long name", out)
		}
	})
	t.Run("scrolling away from wide rows resets the sideways position", func(t *testing.T) {
		m := newTreeModel()
		tr := tree.New(m.root, m.selection).TopLevel(1).Markers(tree.Markers{})
		var got Change
		for _, msg := range []tview.Msg{key(tcell.KeyPgDn), key(tcell.KeyRight), key(tcell.KeyPgUp), key(tcell.KeyPgDn)} {
			got = m.view().ContentWidth(tr.RowsWidth).Handle(msg, treeArea).(Change)
			m.scrollState.Apply(got)
		}
		if got != (Change{y: 2}) {
			t.Fatalf("got %+v, want back at long name and not scrolled right", got)
		}
	})
	t.Run("following the selection away from wide rows resets the sideways position", func(t *testing.T) {
		type moved struct {
			scroll Change
			msg    tview.Msg
		}
		m := newTreeModel()
		for i := range 8 {
			m.root.AddChild(tree.NewNode(strconv.Itoa(i)))
		}
		m.selection.SetCurrentNode(m.root.Children()[2])
		send := func(msg tview.Msg) {
			tr := tree.New(m.root, m.selection).TopLevel(1).Markers(tree.Markers{})
			out := m.view().ContentWidth(tr.RowsWidth).OnChildMsg(func(c Change, msg tview.Msg) tview.Msg { return moved{c, msg} }).Handle(msg, treeArea)
			if joined, ok := out.(moved); ok {
				m.scrollState.Apply(joined.scroll)
				out = joined.msg
			}
			switch out := out.(type) {
			case tree.Change:
				m.selection.Apply(out)
				m.scrollState.ScrollToTarget()
			case Change:
				m.scrollState.Apply(out)
			}
		}
		send(key(tcell.KeyPgDn))
		send(key(tcell.KeyRight))
		for range 4 {
			send(key(tcell.KeyDown))
		}
		for range 4 {
			send(key(tcell.KeyUp))
		}
		if m.scrollState.x != 0 || m.selection.CurrentNode() != m.root.Children()[2] {
			t.Fatalf("x = %d at %v, want 0 at long name", m.scrollState.x, m.selection.CurrentNode().Line())
		}
	})
	t.Run("does not scroll without OnChange", func(t *testing.T) {
		down := key(tcell.KeyPgDn)
		if got := New(text.New("a"), ScrollState{}).Focused(true).Handle(down, treeArea); got != down {
			t.Fatalf("got %v", got)
		}
	})
}

// numbers returns a list of count rows labeled 0, 1, and so on, in a viewport with a scroll bar without end symbols.
func numbers(selection list.SelectionState, scrollState ScrollState, count int) Widget {
	l := list.New(selection, count, func(i int) tview.Widget { return text.New(strconv.Itoa(i)) })
	return New(l, scrollState).
		Target(l.Target).
		ScrollBar(scrollbar.New().BeginSymbol("").EndSymbol(""), ScrollBarVisibilityAutomatic).
		TrackEnd(true).
		OnChange(func(c Change) tview.Msg { return c })
}

// apply sends msgs through numbers in area, applying each Change.
func apply(t *testing.T, scrollState ScrollState, count int, area tview.Rectangle, msgs ...tview.Msg) ScrollState {
	t.Helper()
	for _, msg := range msgs {
		change, ok := numbers(list.NewSelectionState(), scrollState, count).Handle(msg, area).(Change)
		if !ok {
			t.Fatalf("%v: want a Change", msg)
		}
		scrollState.Apply(change)
	}
	return scrollState
}

func TestScrollBar(t *testing.T) {
	area := tview.Rectangle{Width: 4, Height: 4}
	t.Run("track click pages", func(t *testing.T) {
		if got := apply(t, ScrollState{}, 10, area, mouse(3, 3, tview.MouseLeftDown)); got.y != 4 {
			t.Fatalf("y = %d, want 4", got.y)
		}
	})
	t.Run("dragging the thumb scrolls", func(t *testing.T) {
		got := apply(t, ScrollState{}, 10, area, mouse(3, 0, tview.MouseLeftDown), mouse(3, 3, tview.MouseMove), mouse(5, 3, tview.MouseLeftUp))
		if got.y != 6 || got.dragging {
			t.Fatalf("y %d dragging %v, want 6 and not dragging", got.y, got.dragging)
		}
	})
	t.Run("holding the thumb still keeps the scroll position", func(t *testing.T) {
		wheel := mouse(0, 0, tview.MouseScrollDown)
		got := apply(t, ScrollState{}, 100, area, wheel, wheel, wheel, mouse(3, 0, tview.MouseLeftDown), mouse(9, 0, tview.MouseMove))
		if got.y != 3 {
			t.Fatalf("y = %d, want 3", got.y)
		}
	})
	t.Run("a scroll bar too short to draw takes no clicks", func(t *testing.T) {
		short := tview.Rectangle{Width: 4, Height: 2}
		msg := numbers(list.NewSelectionState(), ScrollState{}, 10).ScrollBar(scrollbar.New(), ScrollBarVisibilityAutomatic).Handle(mouse(3, 1, tview.MouseLeftDown), short)
		if a, ok := msg.(Change); !ok || a.y != 0 {
			t.Fatalf("got %v, want y 0", msg)
		}
	})
	t.Run("the end symbol scrolls a step", func(t *testing.T) {
		msg := numbers(list.NewSelectionState(), ScrollState{}, 10).ScrollBar(scrollbar.New(), ScrollBarVisibilityAutomatic).Handle(mouse(3, 3, tview.MouseLeftDown), area)
		if a, ok := msg.(Change); !ok || a.y != 1 {
			t.Fatalf("got %v, want y 1", msg)
		}
	})
	t.Run("moving the mouse without dragging passes through", func(t *testing.T) {
		move := mouse(1, 3, tview.MouseMove)
		if got := numbers(list.NewSelectionState(), ScrollState{}, 50).Handle(move, area); got != tview.Msg(move) {
			t.Fatalf("got %v", got)
		}
	})
}

func TestScrollTo(t *testing.T) {
	area := tview.Rectangle{Width: 3, Height: 2}
	rows := func(t *testing.T, selection list.SelectionState, scrollState ScrollState, count int) string {
		t.Helper()
		screen := screentest.New(t, 3, 2)
		numbers(selection, scrollState, count).ScrollBar(scrollbar.New(), ScrollBarVisibilityNever).Draw(screen, area)
		return screentest.Row(screen, 0, 1) + screentest.Row(screen, 1, 1)
	}
	// atEnd returns the state scrolled to the end of 5 rows, with msgs applied.
	atEnd := func(t *testing.T, msgs ...tview.Msg) ScrollState {
		t.Helper()
		var scrollState ScrollState
		scrollState.ScrollToEnd()
		return apply(t, scrollState, 5, area, msgs...)
	}
	t.Run("the target stays in view once deselected", func(t *testing.T) {
		selection := list.NewSelectionState()
		selection.SetCursor(3)
		selection.SetCursor(-1)
		var scrollState ScrollState
		scrollState.ScrollToTarget()
		if got := rows(t, selection, scrollState, 9); got != "23" {
			t.Fatalf("rows = %q", got)
		}
	})
	t.Run("stays at the end when an item is added", func(t *testing.T) {
		if got := rows(t, list.NewSelectionState(), atEnd(t), 6); got != "45" {
			t.Fatalf("rows = %q", got)
		}
	})
	t.Run("scrolling back to the end tracks it again", func(t *testing.T) {
		scrollState := atEnd(t, mouse(0, 0, tview.MouseScrollUp), mouse(0, 0, tview.MouseScrollDown))
		if got := rows(t, list.NewSelectionState(), scrollState, 6); got != "45" {
			t.Fatalf("rows = %q", got)
		}
	})
	t.Run("scrolled up stays put", func(t *testing.T) {
		scrollState := atEnd(t, mouse(0, 0, tview.MouseScrollUp))
		if got := rows(t, list.NewSelectionState(), scrollState, 6); got != "23" {
			t.Fatalf("rows = %q", got)
		}
	})
}
