package tview

import (
	"testing"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/vt"
)

func TestLineClone(t *testing.T) {
	var nilLine Line
	if nilLine.Clone() != nil {
		t.Fatal("nil line cloned as non-nil")
	}

	line := Line{{Text: "original"}}
	clone := line.Clone()
	clone[0].Text = "clone"
	if line[0].Text != "original" {
		t.Fatal("clone shares backing storage")
	}
}

func TestTextViewLinesReturnsCopy(t *testing.T) {
	view := NewTextView().SetLines([]Line{{{Text: "original"}}})
	lines := view.Lines()
	lines[0][0].Text = "copy"
	if view.Lines()[0][0].Text != "original" {
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
