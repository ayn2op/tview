package stack_test

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/box"
	"github.com/ayn2op/tview/center"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/layout"
	"github.com/ayn2op/tview/richtext"
	"github.com/ayn2op/tview/stack"
	"github.com/ayn2op/tview/text"
	"github.com/ayn2op/tview/textview"
	"github.com/gdamore/tcell/v3"
)

// A popup drawn over the content behind it.
func ExampleNew() {
	content := textview.New(richtext.New(
		richtext.NewLine(richtext.NewSegment("first line of content", tcell.StyleDefault)),
		richtext.NewLine(richtext.NewSegment("second line of content", tcell.StyleDefault)),
		richtext.NewLine(richtext.NewSegment("third line of content", tcell.StyleDefault)),
	))
	popup := center.New(box.New(text.New("Saved")).Borders(tview.BordersAll).Width(layout.Shrink).Height(layout.Shrink))
	screentest.Print(stack.New(content, popup), 22, 3)
	// Output:
	// first l┌─────┐content
	// second │Saved│ content
	// third l└─────┘content
}
