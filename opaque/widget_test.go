package opaque

import (
	"github.com/ayn2op/tview/layout"
	"testing"

	"github.com/ayn2op/tview"
	"github.com/gdamore/tcell/v3"
)

type clicked struct{}

// clicker returns clicked for a left click and passes anything else through.
type clicker struct{}

func (clicker) Draw(tview.Screen, tview.Rectangle) {}

func (clicker) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	if mouse, ok := msg.(tview.MouseMsg); ok && mouse.Action == tview.MouseLeftClick {
		return clicked{}
	}
	return msg
}

func mouse(x, y int, action tview.MouseAction) tview.MouseMsg {
	return tview.MouseMsg{EventMouse: tcell.NewEventMouse(x, y, tcell.ButtonNone, tcell.ModNone), Action: action}
}

func TestWidgetHandle(t *testing.T) {
	area := tview.Rectangle{Width: 4, Height: 2}
	o := New(clicker{})
	t.Run("drops mouse inside", func(t *testing.T) {
		if got := o.Handle(mouse(1, 1, tview.MouseScrollUp), area); got != nil {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("passes mouse outside", func(t *testing.T) {
		outside := mouse(9, 1, tview.MouseScrollUp)
		if got := o.Handle(outside, area); got != outside {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("keeps child message", func(t *testing.T) {
		if got := o.Handle(mouse(1, 1, tview.MouseLeftClick), area); got != (clicked{}) {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("passes keys", func(t *testing.T) {
		enter := tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone)
		if got := o.Handle(enter, area); got != enter {
			t.Fatalf("got %v", got)
		}
	})
}

func (clicker) Size() (width, height layout.Length) { return layout.Fill, layout.Fill }

func (clicker) Layout(limits layout.Limits) layout.Size {
	return layout.Atomic(limits, layout.Fill, layout.Fill)
}
