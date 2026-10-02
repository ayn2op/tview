package inert

import (
	"github.com/ayn2op/tview/layout"
	"testing"

	"github.com/ayn2op/tview"
	"github.com/gdamore/tcell/v3"
)

type taken struct{}

// taker turns every message into taken, and records whether it was drawn.
type taker struct {
	drawn *bool
	width layout.Length
}

func (t taker) Size() (width, height layout.Length) { return t.width, layout.Fill }

func (t taker) Draw(tview.Screen, tview.Rectangle) { *t.drawn = true }

func (taker) Handle(tview.Msg, tview.Rectangle) tview.Msg { return taken{} }

func TestWidget(t *testing.T) {
	var drawn bool
	w := New(taker{drawn: &drawn, width: layout.Fixed(3)})
	area := tview.Rectangle{Width: 4, Height: 2}

	t.Run("draws the child", func(t *testing.T) {
		w.Draw(nil, area)
		if !drawn {
			t.Fatal("child was not drawn")
		}
	})
	t.Run("keeps the child's size", func(t *testing.T) {
		if width, _ := w.Size(); width != layout.Fixed(3) {
			t.Fatalf("width = %v", width)
		}
	})
	t.Run("gives the child no input", func(t *testing.T) {
		for _, msg := range []tview.Msg{
			tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone),
			tview.MouseMsg{EventMouse: tcell.NewEventMouse(1, 1, tcell.ButtonPrimary, tcell.ModNone), Action: tview.MouseLeftClick},
			tview.PasteMsg("text"),
		} {
			if got := w.Handle(msg, area); got != msg {
				t.Fatalf("%T: got %v", msg, got)
			}
		}
	})
}

func (t taker) Layout(limits layout.Limits) layout.Size {
	return layout.Atomic(limits, t.width, layout.Fill)
}
