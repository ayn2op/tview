package tview

import (
	"testing"

	"github.com/gdamore/tcell/v3"
)

func TestFormNavigation(t *testing.T) {
	for _, key := range []tcell.Key{tcell.KeyTab, tcell.KeyEnter} {
		form := NewForm().AddInputField("", "", 0).AddCheckbox("", false)
		form.Update(tcell.NewEventKey(key, "", 0))
		if form.focused != 1 {
			t.Fatalf("%v: focused item %d", key, form.focused)
		}
	}
}
