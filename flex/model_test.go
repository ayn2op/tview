package flex

import (
	"testing"

	"github.com/ayn2op/tview"
)

func TestLocalFocus(t *testing.T) {
	first := tview.NewBox()
	second := tview.NewBox()
	m := NewModel().
		AddItem(first, 0, 1, true).
		AddItem(second, 0, 1, false)

	if m.Focused() != 0 {
		t.Fatal("first item was not selected")
	}

	m.SetFocus(1)
	if m.Focused() != 1 {
		t.Fatal("second item was not selected")
	}
}
