package list

import (
	"strconv"
	"testing"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/button"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/layout"
	"github.com/ayn2op/tview/slider"
	"github.com/ayn2op/tview/text"
	"github.com/gdamore/tcell/v3"
)

// row is an item one line tall showing its label.
type row string

func (row) Size() (width, height layout.Length) { return layout.Fill, layout.Fixed(1) }

func (row) Layout(limits layout.Limits) layout.Size {
	return layout.Atomic(limits, layout.Fill, layout.Fixed(1))
}

func (r row) Draw(screen tview.Screen, area tview.Rectangle) {
	tview.Print(screen, string(r), area.X, area.Y, area.Width, tview.HorizontalAlignmentLeft, tcell.StyleDefault)
}

func (row) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg { return msg }

// numbers returns a list of count rows labeled 0, 1, and so on.
func numbers(cursor, count int) Widget {
	return New(cursor, count, func(i int) tview.Widget { return row(strconv.Itoa(i)) }).
		Focused(true).
		OnChange(func(cursor int) tview.Msg { return moved(cursor) })
}

// moved is the item a list moved its cursor to.
type moved int

// apply sends msgs through a list of count rows in area, moving the cursor as each says, and returns where it ends up.
func apply(t *testing.T, cursor, count int, area tview.Rectangle, msgs ...tview.Msg) int {
	t.Helper()
	for _, msg := range msgs {
		out, ok := numbers(cursor, count).Handle(msg, area).(moved)
		if !ok {
			t.Fatalf("%v: got %v, want the cursor moved", msg, out)
		}
		cursor = int(out)
	}
	return cursor
}

func key(k tcell.Key) tview.KeyMsg { return tcell.NewEventKey(k, "", tcell.ModNone) }

func mouse(x, y int, action tview.MouseAction) tview.MouseMsg {
	return tview.MouseMsg{EventMouse: tcell.NewEventMouse(x, y, tcell.ButtonNone, tcell.ModNone), Action: action}
}

func TestWidgetHandle(t *testing.T) {
	area := tview.Rectangle{Width: 4, Height: 4}
	t.Run("keys move the cursor", func(t *testing.T) {
		got := apply(t, -1, 10, area, key(tcell.KeyDown), key(tcell.KeyDown), key(tcell.KeyEnd), key(tcell.KeyUp))
		if got != 8 {
			t.Fatalf("cursor = %d, want 8", got)
		}
	})
	t.Run("click selects the item", func(t *testing.T) {
		got := apply(t, -1, 10, area, mouse(0, 2, tview.MouseLeftDown))
		if got != 2 {
			t.Fatalf("cursor = %d, want 2", got)
		}
	})
	t.Run("passes other messages through", func(t *testing.T) {
		for _, msg := range []tview.Msg{key(tcell.KeyEnter), mouse(1, 3, tview.MouseMove)} {
			if got := numbers(-1, 10).Handle(msg, area); got != msg {
				t.Fatalf("got %v", got)
			}
		}
	})
}

func TestWidgetDraw(t *testing.T) {
	area := tview.Rectangle{Width: 3, Height: 2}
	t.Run("the cursor in the selected style", func(t *testing.T) {
		screen := screentest.New(t, 3, 2)
		numbers(1, 3).Draw(screen, area)
		if got := screentest.Row(screen, 0, 3) + screentest.Row(screen, 1, 3); got != "0  1  " {
			t.Fatalf("rows = %q", got)
		}
		if _, style, _ := screen.Get(0, 1); !style.HasReverse() {
			t.Fatal("selected item is not reversed")
		}
	})
	t.Run("clips fills of items partly out of view", func(t *testing.T) {
		screen := screentest.New(t, 3, 3)
		New(-1, 1, func(int) tview.Widget { return filled{} }).Draw(screen, tview.Rectangle{Y: 1, Width: 3, Height: 1})
		if got := screentest.Row(screen, 0, 3) + screentest.Row(screen, 1, 3) + screentest.Row(screen, 2, 3); got != "   xxx   " {
			t.Fatalf("rows = %q", got)
		}
	})
	t.Run("lays items out at one width", func(t *testing.T) {
		widths := map[int]bool{}
		w := New(-1, 1, func(int) tview.Widget { return measured{widths: widths} }).OnChange(func(cursor int) tview.Msg { return moved(cursor) })
		w.Layout(layout.Limits{Max: area.Size(), Infinite: layout.Axes{Height: true}})
		w.Draw(screentest.New(t, 3, 2), area)
		w.Handle(mouse(0, 0, tview.MouseLeftDown), area)
		if len(widths) != 1 || !widths[3] {
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

func TestWidgetLayout(t *testing.T) {
	limits := layout.Limits{Max: layout.Size{Width: 4, Height: 2}, Infinite: layout.Axes{Height: true}}
	if got := numbers(-1, 5).Spacing(func(int) int { return 1 }).Layout(limits); got != (layout.Size{Width: 4, Height: 9}) {
		t.Fatalf("size = %+v, want 4 by 9", got)
	}
	// Only the odd items have spacing above them.
	if got := numbers(-1, 5).Spacing(func(i int) int { return i % 2 }).Layout(limits); got != (layout.Size{Width: 4, Height: 7}) {
		t.Fatalf("size = %+v, want 4 by 7", got)
	}
}

func TestWidgetTarget(t *testing.T) {
	if top, height := numbers(-1, 5).Spacing(func(int) int { return 1 }).Target(3)(4); top != 6 || height != 1 {
		t.Fatalf("target = %d, %d", top, height)
	}
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

func TestItemHeight(t *testing.T) {
	for _, tt := range []struct {
		name string
		item tview.Widget
		want int
	}{
		{"one row", row(""), 1},
		{"height for the width", wrapped{}, 2},
		{"fixed height", sized{height: layout.Fixed(3)}, 3},
	} {
		t.Run(tt.name, func(t *testing.T) {
			list := New(-1, 1, func(int) tview.Widget { return tt.item })
			if _, sizes, _ := list.layout(4); sizes[0] != tt.want {
				t.Fatalf("rows = %d, want %d", sizes[0], tt.want)
			}
		})
	}
}

func TestWidgetHandleItem(t *testing.T) {
	area := tview.Rectangle{Width: 3, Height: 2}
	w := New(-1, 2, func(i int) tview.Widget {
		return button.New(text.New("")).Height(layout.Fixed(1)).OnClick(i)
	}).OnChange(func(cursor int) tview.Msg { return moved(cursor) })

	// The item clicked takes the click, which then does not select it.
	if got := w.Handle(mouse(0, 1, tview.MouseLeftDown), area); got != 1 {
		t.Fatalf("got %v, want 1", got)
	}

	// A drag reaches the item dragged, wherever the pointer went.
	sliders := New(-1, 2, func(i int) tview.Widget {
		return slider.New(0, 2, 0, func(value int) tview.Msg { return value }).Dragging(i == 1)
	}).OnChange(func(cursor int) tview.Msg { return moved(cursor) })
	drag := tview.MouseMsg{EventMouse: tcell.NewEventMouse(9, 9, tcell.ButtonPrimary, tcell.ModNone), Action: tview.MouseMove}
	if got := sliders.Handle(drag, area); got != 2 {
		t.Fatalf("got %v, want 2", got)
	}
}
