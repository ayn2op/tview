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
		m.form.GetButton(0).Update(tview.FocusMsg{})

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
