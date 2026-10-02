package tview

import (
	"github.com/ayn2op/tview/layout"
	"testing"

	"github.com/gdamore/tcell/v3"
)

type childMsg struct{ n int }

type parentMsg struct{ child childMsg }

// emitter returns childMsg{1} for any mouse message and passes anything else through.
type emitter struct{}

func (emitter) Draw(Screen, Rectangle) {}

func (emitter) Handle(msg Msg, area Rectangle) Msg {
	if _, ok := msg.(MouseMsg); ok {
		return childMsg{1}
	}
	return msg
}

type sliceMsg []int

// sliceEmitter returns sliceMsg{1, 2} for paste and passes anything else through.
type sliceEmitter struct{}

func (sliceEmitter) Draw(Screen, Rectangle) {}

func (sliceEmitter) Handle(msg Msg, area Rectangle) Msg {
	if _, ok := msg.(PasteMsg); ok {
		return sliceMsg{1, 2}
	}
	return msg
}

func TestMapHandle(t *testing.T) {
	m := Map(emitter{}, func(c childMsg) Msg { return parentMsg{c} })
	area := Rectangle{Width: 1, Height: 1}
	t.Run("maps produced messages", func(t *testing.T) {
		click := MouseMsg{EventMouse: tcell.NewEventMouse(0, 0, tcell.Button1, tcell.ModNone), Action: MouseLeftClick}
		if got := m.Handle(click, area); got != (parentMsg{childMsg{1}}) {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("passes input through", func(t *testing.T) {
		enter := tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone)
		if got := m.Handle(enter, area); got != enter {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("maps uncomparable messages", func(t *testing.T) {
		m := Map(sliceEmitter{}, func(s sliceMsg) Msg { return parentMsg{childMsg{len(s)}} })
		if got := m.Handle(PasteMsg("x"), area); got != (parentMsg{childMsg{2}}) {
			t.Fatalf("got %v", got)
		}
		if got, ok := m.Handle(sliceMsg{1}, area).(sliceMsg); !ok || len(got) != 1 {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("leaves passed through messages of type T", func(t *testing.T) {
		if got := m.Handle(childMsg{2}, area); got != (childMsg{2}) {
			t.Fatalf("got %v", got)
		}
	})
}

func (emitter) Size() (width, height layout.Length) { return layout.Fill, layout.Fill }

func (emitter) Layout(limits layout.Limits) layout.Size {
	return layout.Atomic(limits, layout.Fill, layout.Fill)
}

func (sliceEmitter) Size() (width, height layout.Length) { return layout.Fill, layout.Fill }

func (sliceEmitter) Layout(limits layout.Limits) layout.Size {
	return layout.Atomic(limits, layout.Fill, layout.Fill)
}
