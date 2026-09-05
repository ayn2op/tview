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

func TestInputFieldMask(t *testing.T) {
	i := NewInputField().SetMaskCharacter('*')
	if i.textArea.mask != "*" || i.textArea.maskWidth <= 0 {
		t.Fatalf("mask = %q width %d", i.textArea.mask, i.textArea.maskWidth)
	}
	i.SetMaskCharacter(0)
	if i.textArea.mask != "" || i.textArea.maskWidth != 0 {
		t.Fatalf("mask not disabled: %q width %d", i.textArea.mask, i.textArea.maskWidth)
	}
}
