package stack_test

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/box"
	"github.com/ayn2op/tview/center"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/layout"
	"github.com/ayn2op/tview/stack"
	"github.com/ayn2op/tview/text"
)

// A popup drawn over the content behind it.
func ExampleNew() {
	content := text.New("first line of content\nsecond line of content\nthird line of content")
	popup := center.New(box.New(text.New("Saved")).Borders(tview.BordersAll).Width(layout.Shrink).Height(layout.Shrink))
	screentest.Print(stack.New(content, popup), 22, 3)
	// Output:
	// first l┌─────┐content
	// second │Saved│ content
	// third l└─────┘content
}
