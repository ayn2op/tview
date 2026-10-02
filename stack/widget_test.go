package stack

import (
	"github.com/ayn2op/tview/layout"
	"testing"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/gdamore/tcell/v3"
)

// letter fills its area with str, records that it saw a message, and drops Escape.
type letter struct {
	str  string
	seen *[]string
}

func (l letter) Draw(screen tview.Screen, area tview.Rectangle) {
	for y := area.Y; y < area.Y+area.Height; y++ {
		for x := area.X; x < area.X+area.Width; x++ {
			screen.Put(x, y, l.str, tcell.StyleDefault)
		}
	}
}

func (l letter) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	*l.seen = append(*l.seen, l.str)
	if key, ok := msg.(tview.KeyMsg); ok && key.Key() == tcell.KeyEscape {
		return nil
	}
	return msg
}

var area = tview.Rectangle{Width: 4, Height: 2}

func TestWidgetDraw(t *testing.T) {
	screen := screentest.New(t, 4, 2)
	var seen []string
	New(letter{"a", &seen}, letter{"b", &seen}).Draw(screen, area)
	if str, _, _ := screen.Get(1, 1); str != "b" {
		t.Fatalf("top cell = %q, want %q", str, "b")
	}
}

func TestWidgetHandle(t *testing.T) {
	t.Run("top first", func(t *testing.T) {
		var seen []string
		enter := tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone)
		if got := New(letter{"a", &seen}, letter{"b", &seen}).Handle(enter, area); got != enter {
			t.Fatalf("got %v", got)
		}
		if len(seen) != 2 || seen[0] != "b" || seen[1] != "a" {
			t.Fatalf("order = %v", seen)
		}
	})
	t.Run("stops when dropped", func(t *testing.T) {
		var seen []string
		escape := tcell.NewEventKey(tcell.KeyEscape, "", tcell.ModNone)
		if got := New(letter{"a", &seen}, letter{"b", &seen}).Handle(escape, area); got != nil {
			t.Fatalf("got %v", got)
		}
		if len(seen) != 1 {
			t.Fatalf("seen = %v", seen)
		}
	})
}

func (letter) Layout(limits layout.Limits) tview.Size {
	return layout.Atomic(limits, tview.Fill, tview.Fill)
}
