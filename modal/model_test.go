package modal

import (
	"testing"

	"github.com/ayn2op/tview"
	"github.com/gdamore/tcell/v3"
)

func TestModelDone(t *testing.T) {
	for _, test := range []struct {
		key   tcell.Key
		index int
		label string
	}{
		{tcell.KeyEnter, 0, "Yes"},
		{tcell.KeyEscape, -1, ""},
	} {
		m := NewModel().AddButtons([]string{"Yes", "No"})
		var focused tview.Model
		var focus func(tview.Model)
		focus = func(next tview.Model) {
			if focused != nil {
				focused.Blur()
			}
			focused = next
			next.Focus(focus)
		}
		focus(m)

		msg := tview.Msg(tcell.NewEventKey(test.key, "", tcell.ModNone))
		for {
			cmd := m.Update(msg)
			if cmd == nil {
				t.Fatalf("key %v produced no command", test.key)
			}
			msg = cmd()
			done, ok := msg.(DoneMsg)
			if !ok {
				continue
			}
			if done.ButtonIndex != test.index || done.ButtonLabel != test.label {
				t.Fatalf("key %v: got %+v, want index %d label %q", test.key, done, test.index, test.label)
			}
			break
		}
	}
}
