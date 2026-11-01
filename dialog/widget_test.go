package dialog

import (
	"testing"

	"github.com/ayn2op/tview"
	"github.com/gdamore/tcell/v3"
)

var area = tview.Rectangle{Width: 30, Height: 9}

type focusMsg int

type doneMsg struct {
	index int
	label string
}

func key(k tcell.Key) tview.KeyMsg { return tcell.NewEventKey(k, "", tcell.ModNone) }

func dialog(focus int) Widget {
	return New().
		Buttons("Yes", "No").
		Focus(focus).
		OnFocus(func(i int) tview.Msg { return focusMsg(i) }).
		OnDone(func(i int, label string) tview.Msg { return doneMsg{i, label} })
}

func TestWidgetHandle(t *testing.T) {
	for _, tt := range []struct {
		name  string
		focus int
		msg   tview.Msg
		want  tview.Msg
	}{
		{"enter presses the focused button", 1, key(tcell.KeyEnter), doneMsg{1, "No"}},
		{"tab moves the focus", 1, key(tcell.KeyTab), focusMsg(0)},
		{"backtab moves the focus back", 0, key(tcell.KeyBacktab), focusMsg(1)},
		{"escape cancels", 0, key(tcell.KeyEscape), doneMsg{-1, ""}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := dialog(tt.focus).Handle(tt.msg, area); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}
