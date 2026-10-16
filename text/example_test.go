package text_test

import (
	"github.com/ayn2op/tview/column"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/text"
)

// Text is one line that is cut off at the edge. For text that wraps or has styles, use textview.
func ExampleNew() {
	screentest.Print(column.New(text.New("short"), text.New("a line too long to fit")), 12, 2)
	// Output:
	// short
	// a line too l
}
