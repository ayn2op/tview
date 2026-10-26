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

// current is the node a tree made current.
type current struct{ *tree.Node }

// treeModel is a tree of four rows, the third wider than the area, in a viewport that scrolls both ways.
type treeModel struct {
	root        *tree.Node
	selection   *tree.Node
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
		OnChange(func(n *tree.Node) tview.Msg { return current{n} }).
		OnSelect(func(n *tree.Node) tview.Msg { return n })
	return New(t, m.scrollState).
		Target(t.Target).
		Axes(layout.Axes{Width: true, Height: true}).
		Focused(true).
		OnChange(func(s *ScrollState) tview.Msg { return *s })
}

func (m *treeModel) send(msg tview.Msg) tview.Msg {
	switch out := m.view().Handle(msg, treeArea).(type) {
	case current:
		m.selection = out.Node
		m.scrollState.ScrollToTarget()
	case ScrollState:
		m.scrollState = out
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

func TestWidgetHandle(t *testing.T) {
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
		t.Run("page down", func(t *testing.T) {
			m.send(key(tcell.KeyPgDn))
			check(t, m, blank, " long n ", " d      ", blank)
		})
		t.Run("right", func(t *testing.T) {
			m.send(key(tcell.KeyRight))
			m.send(key(tcell.KeyRight))
			check(t, m, blank, " ng nam ", "        ", blank)
		})
		t.Run("the wheel stops at the child's edge", func(t *testing.T) {
			// The child is 9 cells wide, so the view stops 3 cells in.
			m.send(mouse(2, 1, tview.MouseScrollRight))
			check(t, m, blank, " g name ", "        ", blank)
		})
	})
	t.Run("the wheel scrolls within the area", func(t *testing.T) {
		m := newTreeModel()
		m.send(mouse(2, 1, tview.MouseScrollDown))
		check(t, m, blank, " b      ", " long n ", blank)
	})
	t.Run("the wheel outside the area passes through", func(t *testing.T) {
		if out := newTreeModel().send(mouse(0, 0, tview.MouseScrollDown)); out == nil {
			t.Fatal("the wheel outside the area was used")
		}
	})
	t.Run("the target is kept in the middle", func(t *testing.T) {
		m := newTreeModel()
		m.selection = m.root.Children()[0]
		for range 3 {
			m.send(key(tcell.KeyDown))
		}
		if got := m.selection; got != m.root.Children()[3] {
			t.Fatalf("current = %v", got.Spans())
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
		t.Run("not past narrow rows", func(t *testing.T) {
			if out := m.view().ContentWidth(tr.RowsWidth).Handle(key(tcell.KeyRight), treeArea).(ScrollState); out.x != 0 || out.y != 0 {
				t.Fatalf("got %+v, want no scrolling past a and b", out)
			}
		})
		t.Run("along a wide row", func(t *testing.T) {
			m.send(key(tcell.KeyPgDn))
			if out := m.view().ContentWidth(tr.RowsWidth).Handle(key(tcell.KeyRight), treeArea).(ScrollState); out.x != 1 || out.y != 2 {
				t.Fatalf("got %+v, want a cell right at long name", out)
			}
		})
	})
	t.Run("scrolling away from wide rows resets the sideways position", func(t *testing.T) {
		m := newTreeModel()
		tr := tree.New(m.root, m.selection).TopLevel(1).Markers(tree.Markers{})
		var got ScrollState
		for _, msg := range []tview.Msg{key(tcell.KeyPgDn), key(tcell.KeyRight), key(tcell.KeyPgUp), key(tcell.KeyPgDn)} {
			got = m.view().ContentWidth(tr.RowsWidth).Handle(msg, treeArea).(ScrollState)
			m.scrollState = got
		}
		if got.x != 0 || got.y != 2 {
			t.Fatalf("got %+v, want back at long name and not scrolled right", got)
		}
	})
	t.Run("following the selection away from wide rows resets the sideways position", func(t *testing.T) {
		type moved struct {
			scroll ScrollState
			msg    tview.Msg
		}
		m := newTreeModel()
		for i := range 8 {
			m.root.AddChild(tree.NewNode(strconv.Itoa(i)))
		}
		m.selection = m.root.Children()[2]
		send := func(msg tview.Msg) {
			tr := tree.New(m.root, m.selection).TopLevel(1).Markers(tree.Markers{})
			out := m.view().ContentWidth(tr.RowsWidth).OnChildMsg(func(s *ScrollState, msg tview.Msg) tview.Msg { return moved{*s, msg} }).Handle(msg, treeArea)
			if joined, ok := out.(moved); ok {
				m.scrollState = joined.scroll
				out = joined.msg
			}
			switch out := out.(type) {
			case current:
				m.selection = out.Node
				m.scrollState.ScrollToTarget()
			case ScrollState:
				m.scrollState = out
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
		if m.scrollState.x != 0 || m.selection != m.root.Children()[2] {
			t.Fatalf("x = %d at %v, want 0 at long name", m.scrollState.x, m.selection.Spans())
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
func numbers(selection int, scrollState ScrollState, count int) Widget {
	l := list.New(selection, count, func(i int) tview.Widget { return text.New(strconv.Itoa(i)) })
	return New(l, scrollState).
		Target(l.Target(selection)).
		ScrollBar(scrollbar.New().BeginSymbol("").EndSymbol(""), ScrollBarVisibilityAutomatic).
		TrackEnd(true).
		OnChange(func(s *ScrollState) tview.Msg { return *s })
}

// apply sends msgs through numbers in area, scrolling to what each produces, if anything.
func apply(t *testing.T, scrollState ScrollState, count int, area tview.Rectangle, msgs ...tview.Msg) ScrollState {
	t.Helper()
	for _, msg := range msgs {
		switch change := numbers(-1, scrollState, count).Handle(msg, area).(type) {
		case ScrollState:
			scrollState = change
		case nil:
		default:
			t.Fatalf("%v: got %v, want a ScrollState or nothing", msg, change)
		}
	}
	return scrollState
}

func TestWidgetScrollBar(t *testing.T) {
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
		msg := numbers(-1, ScrollState{}, 10).ScrollBar(scrollbar.New(), ScrollBarVisibilityAutomatic).Handle(mouse(3, 1, tview.MouseLeftDown), short)
		if a, ok := msg.(ScrollState); !ok || a.y != 0 {
			t.Fatalf("got %v, want y 0", msg)
		}
	})
	t.Run("the end symbol scrolls a step", func(t *testing.T) {
		msg := numbers(-1, ScrollState{}, 10).ScrollBar(scrollbar.New(), ScrollBarVisibilityAutomatic).Handle(mouse(3, 3, tview.MouseLeftDown), area)
		if a, ok := msg.(ScrollState); !ok || a.y != 1 {
			t.Fatalf("got %v, want y 1", msg)
		}
	})
	t.Run("moving the mouse without dragging passes through", func(t *testing.T) {
		move := mouse(1, 3, tview.MouseMove)
		if got := numbers(-1, ScrollState{}, 50).Handle(move, area); got != tview.Msg(move) {
			t.Fatalf("got %v", got)
		}
	})
}

func TestWidgetTrackEnd(t *testing.T) {
	area := tview.Rectangle{Width: 3, Height: 2}
	rows := func(t *testing.T, selection int, scrollState ScrollState, count int) string {
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
	t.Run("the target is kept in view", func(t *testing.T) {
		var scrollState ScrollState
		scrollState.ScrollToTarget()
		if got := rows(t, 3, scrollState, 9); got != "23" {
			t.Fatalf("rows = %q", got)
		}
	})
	t.Run("stays at the end when an item is added", func(t *testing.T) {
		if got := rows(t, -1, atEnd(t), 6); got != "45" {
			t.Fatalf("rows = %q", got)
		}
	})
	t.Run("scrolling back to the end tracks it again", func(t *testing.T) {
		scrollState := atEnd(t, mouse(0, 0, tview.MouseScrollUp), mouse(0, 0, tview.MouseScrollDown))
		if got := rows(t, -1, scrollState, 6); got != "45" {
			t.Fatalf("rows = %q", got)
		}
	})
	t.Run("scrolled up stays put", func(t *testing.T) {
		scrollState := atEnd(t, mouse(0, 0, tview.MouseScrollUp))
		if got := rows(t, -1, scrollState, 6); got != "23" {
			t.Fatalf("rows = %q", got)
		}
	})
}

func TestWidgetAnchorBottom(t *testing.T) {
	area := tview.Rectangle{Width: 3, Height: 2}
	rows := func(t *testing.T, scrollState ScrollState, count int) string {
		t.Helper()
		screen := screentest.New(t, 3, 2)
		numbers(-1, scrollState, count).AnchorBottom().Draw(screen, area)
		return screentest.Row(screen, 0, 1) + screentest.Row(screen, 1, 1)
	}
	scrollUp := func(t *testing.T, count int) ScrollState {
		t.Helper()
		var scrollState ScrollState
		scrollState.ScrollToEnd()
		return numbers(-1, scrollState, count).AnchorBottom().Handle(mouse(0, 0, tview.MouseScrollUp), area).(ScrollState)
	}
	t.Run("the zero state is at the end", func(t *testing.T) {
		if got := rows(t, ScrollState{}, 5); got != "34" {
			t.Fatalf("rows = %q", got)
		}
	})
	t.Run("rows added at the start keep what is shown in place", func(t *testing.T) {
		// Two rows added before 0 number what showed as 2 and 3 as 4 and 5.
		if got := rows(t, scrollUp(t, 5), 7); got != "45" {
			t.Fatalf("rows = %q", got)
		}
	})
	t.Run("offsets are from the anchor", func(t *testing.T) {
		v := scrollUp(t, 5).Viewport()
		if got := v.AbsoluteOffset(); got != (AbsoluteOffset{Y: 1}) {
			t.Fatalf("absolute offset = %+v, want y 1 from the bottom", got)
		}
		if got := v.AbsoluteOffsetReversed(); got != (AbsoluteOffset{Y: 2}) {
			t.Fatalf("reversed offset = %+v, want y 2 from the top", got)
		}
		if got := v.DistanceToEnd(); got != (AbsoluteOffset{Y: 2}) {
			t.Fatalf("distance to end = %+v, want y 2 to the top", got)
		}
	})
	t.Run("nothing is reported while the child fits", func(t *testing.T) {
		if out := numbers(-1, ScrollState{}, 2).AnchorBottom().Handle(mouse(0, 0, tview.MouseScrollUp), area); out != nil {
			t.Fatalf("got %+v, want nothing", out)
		}
	})
}

func TestScrollStateViewport(t *testing.T) {
	area := tview.Rectangle{Width: 3, Height: 2}
	scrollState := apply(t, ScrollState{}, 5, area, mouse(0, 0, tview.MouseScrollDown))
	v := scrollState.Viewport()
	t.Run("absolute offset", func(t *testing.T) {
		if got := v.AbsoluteOffset(); got != (AbsoluteOffset{Y: 1}) {
			t.Fatalf("got %+v, want y 1", got)
		}
	})
	t.Run("reversed offset", func(t *testing.T) {
		if got := v.AbsoluteOffsetReversed(); got != (AbsoluteOffset{Y: 2}) {
			t.Fatalf("got %+v, want y 2", got)
		}
	})
	t.Run("relative offset", func(t *testing.T) {
		if got := v.RelativeOffset(); got != (RelativeOffset{Y: 1.0 / 3}) {
			t.Fatalf("got %+v, want y 1/3", got)
		}
	})
	t.Run("end", func(t *testing.T) {
		if got := v.End(); got != (AbsoluteOffset{Y: 3}) {
			t.Fatalf("got %+v, want y 3", got)
		}
	})
	t.Run("bounds and content", func(t *testing.T) {
		if v.Bounds().Height != 2 || v.Content().Height != 5 {
			t.Fatalf("bounds %v content %v", v.Bounds(), v.Content())
		}
	})
	t.Run("origin", func(t *testing.T) {
		if got := scrollState.Origin().AbsoluteOffset(); got != (AbsoluteOffset{}) {
			t.Fatalf("got %+v, want 0", got)
		}
	})
	t.Run("scrolling past an edge produces nothing", func(t *testing.T) {
		top := apply(t, ScrollState{}, 5, area, mouse(0, 0, tview.MouseScrollUp))
		if out := numbers(-1, top, 5).Handle(mouse(0, 0, tview.MouseScrollUp), area); out != nil {
			t.Fatalf("got %+v, want nothing", out)
		}
	})
}

func TestWidgetAnchorRight(t *testing.T) {
	area := tview.Rectangle{Width: 4, Height: 1}
	// line returns a viewport of s that scrolls sideways from the right.
	line := func(s string, scrollState ScrollState) Widget {
		return New(text.New(s), scrollState).
			Axes(layout.Axes{Width: true}).
			AnchorRight().
			OnChange(func(s *ScrollState) tview.Msg { return *s })
	}
	row := func(t *testing.T, w Widget) string {
		t.Helper()
		screen := screentest.New(t, 4, 1)
		w.Draw(screen, area)
		return screentest.Row(screen, 0, 4)
	}
	t.Run("the zero state is at the right end", func(t *testing.T) {
		if got := row(t, line("0123456789", ScrollState{})); got != "6789" {
			t.Fatalf("shows %q, want the right end", got)
		}
	})
	left := line("0123456789", ScrollState{}).Handle(mouse(0, 0, tview.MouseScrollLeft), area).(ScrollState)
	t.Run("absolute offset", func(t *testing.T) {
		if got := left.Viewport().AbsoluteOffset(); got != (AbsoluteOffset{X: 2}) {
			t.Fatalf("got %+v, want x 2 from the right", got)
		}
	})
	t.Run("reversed offset", func(t *testing.T) {
		if got := left.Viewport().AbsoluteOffsetReversed(); got != (AbsoluteOffset{X: 4}) {
			t.Fatalf("got %+v, want x 4 from the left", got)
		}
	})
	t.Run("growing on the left", func(t *testing.T) {
		// Two cells added on the left keep what is shown in place.
		if got := row(t, line("ab0123456789", left)); got != "4567" {
			t.Fatalf("shows %q, want 4567", got)
		}
	})
}
