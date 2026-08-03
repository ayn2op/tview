package tview

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v3"
)

func TestAppendPasteKey(t *testing.T) {
	t.Parallel()

	var buffer strings.Builder
	for _, event := range []*tcell.EventKey{
		tcell.NewEventKey(tcell.KeyRune, "a", tcell.ModNone),
		tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, "b", tcell.ModNone),
		tcell.NewEventKey(tcell.KeyCtrlJ, "", tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, "c", tcell.ModNone),
		tcell.NewEventKey(tcell.KeyTab, "", tcell.ModNone),
	} {
		appendPasteKey(&buffer, event)
	}

	if got, want := buffer.String(), "a\nb\nc\t"; got != want {
		t.Fatalf("appendPasteKey() = %q, want %q", got, want)
	}
}
