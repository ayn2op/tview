package inert

import (
	"fmt"
	"testing"

	"github.com/ayn2op/tview/layout"

	"github.com/ayn2op/tview"
	"github.com/gdamore/tcell/v3"
)

type taken struct{}

// taker turns every message into taken.
type taker struct{}

func (taker) Size() (width, height layout.Length) { return layout.Fill, layout.Fill }

func (taker) Layout(limits layout.Limits) layout.Size {
	return layout.Atomic(limits, layout.Fill, layout.Fill)
}

func (taker) Draw(tview.Screen, tview.Rectangle) {}

func (taker) Handle(tview.Msg, tview.Rectangle) tview.Msg { return taken{} }

func TestWidgetHandle(t *testing.T) {
	w := New(taker{})
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
