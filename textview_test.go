package tview

import (
	"testing"

	"github.com/ayn2op/tview/text"
	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/vt"
)

func TestTextViewScrollSetters(t *testing.T) {
	v := NewTextView()
	for range 2 {
		v.ScrollTo(4, 5).ScrollToEnd().ScrollToBeginning()
		row, column := v.GetScrollOffset()
		if row != 0 || column != 0 || v.trackEnd {
			t.Fatalf("got (%d, %d, %v), want (0, 0, false)", row, column, v.trackEnd)
		}
	}

	v.SetScrollable(false)
	v.lineOffset, v.columnOffset, v.trackEnd = 4, 5, false
	v.ScrollToBeginning().ScrollToEnd().ScrollTo(1, 2).SetScrollable(false)
	row, column := v.GetScrollOffset()
	if row != 4 || column != 5 || v.trackEnd {
		t.Fatalf("got (%d, %d, %v), want (4, 5, false)", row, column, v.trackEnd)
	}
}

func TestTextViewContentReturnsCopy(t *testing.T) {
	view := NewTextView().SetContent(text.Text{{{Text: "original"}}})
	lines := view.Content()
	lines[0][0].Text = "copy"
	if view.Content()[0][0].Text != "original" {
		t.Fatal("lines share backing storage")
	}
}

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
