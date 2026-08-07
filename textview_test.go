package tview

import (
	"testing"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/vt"
)

func TestTextViewReadsDoNotMutate(t *testing.T) {
	view := NewTextView().SetText("one two three").ScrollTo(10, 10)
	view.SetRect(0, 0, 5, 2)
	state := func() [6]int {
		return [6]int{len(view.lines), len(view.wrapped), view.longestLine, view.lastWidth, view.lineOffset, view.columnOffset}
	}
	want := state()

	screen, err := tcell.NewTerminfoScreenFromTty(vt.NewMockTerm(vt.MockOptSize{X: 5, Y: 2}))
	if err != nil {
		t.Fatal(err)
	}
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(screen.Fini)

	view.Height(2)
	view.GetWrappedLineCount()
	view.View(screen)
	if got := state(); got != want {
		t.Fatalf("state changed: got %v, want %v", got, want)
	}
}

func TestTextViewExitMsg(t *testing.T) {
	cmd := NewTextView().Update(tcell.NewEventKey(tcell.KeyEscape, "", tcell.ModNone))
	msg, ok := cmd().(TextViewExitMsg)
	if !ok || msg.Key != tcell.KeyEscape {
		t.Fatalf("got %#v", msg)
	}
}
