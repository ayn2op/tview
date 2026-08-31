package text

import (
	"testing"

	"github.com/gdamore/tcell/v3"
)

func TestLineClone(t *testing.T) {
	line := Line{{Text: "original"}}
	clone := line.Clone()
	clone[0].Text = "clone"
	if line[0].Text != "original" {
		t.Fatal("clone shares backing storage")
	}
}

func TestDimensions(t *testing.T) {
	text := New(
		NewLine(NewSegment("ab", tcell.StyleDefault)),
		NewLine(NewSegment("界界", tcell.StyleDefault)),
	)
	if text.Width() != 4 || text.Height() != 2 {
		t.Fatalf("got %dx%d", text.Width(), text.Height())
	}
}

func TestWrap(t *testing.T) {
	lines := Wrap(NewLine(NewSegment("ab界c", tcell.StyleDefault)), 3)
	if len(lines) != 2 || lines[0][0].Text != "ab" || lines[1][0].Text != "界c" {
		t.Fatalf("got %#v", lines)
	}
}

func TestBuilderWrite(t *testing.T) {
	var builder Builder
	builder.Write("one\ntwo", tcell.StyleDefault)
	text := builder.Finish()
	if len(text) != 2 || text[0][0].Text != "one" || text[1][0].Text != "two" {
		t.Fatalf("got %#v", text)
	}
}
