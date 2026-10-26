package inert

import (
	"fmt"
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

func TestWidgetDraw(t *testing.T) {
	var drawn bool
	New(taker{drawn: &drawn}).Draw(nil, tview.Rectangle{Width: 4, Height: 2})
	if !drawn {
		t.Fatal("child was not drawn")
	}
}

func TestWidgetSize(t *testing.T) {
	if width, _ := New(taker{drawn: new(bool), width: layout.Fixed(3)}).Size(); width != layout.Fixed(3) {
		t.Fatalf("width = %v, want the child's", width)
	}
}

func TestWidgetHandle(t *testing.T) {
	w := New(taker{drawn: new(bool)})
	area := tview.Rectangle{Width: 4, Height: 2}
	// The child gets no input.
	for _, msg := range []tview.Msg{
		tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone),
		tview.MouseMsg{EventMouse: tcell.NewEventMouse(1, 1, tcell.ButtonPrimary, tcell.ModNone), Action: tview.MouseLeftClick},
		tview.PasteMsg("text"),
	} {
		t.Run(fmt.Sprintf("%T", msg), func(t *testing.T) {
			if got := w.Handle(msg, area); got != msg {
				t.Fatalf("got %v", got)
			}
		})
	}
}

func (t taker) Layout(limits layout.Limits) layout.Size {
	return layout.Atomic(limits, t.width, layout.Fill)
}
