package tview

import (
	"testing"

	"github.com/gdamore/tcell/v3"
)

func TestFormNavigation(t *testing.T) {
	for _, key := range []tcell.Key{tcell.KeyTab, tcell.KeyEnter} {
		form := NewForm().AddInputField("", "", 0).AddCheckbox("", false)
		form.items[0].Focus(func(Model) {})
		msg := form.Update(tcell.NewEventKey(key, "", 0))()
		focus, ok := msg.(setFocusMsg)
		if !ok || focus.target != form.items[1] {
			t.Fatalf("%v: %#v", key, msg)
		}
	}
}
