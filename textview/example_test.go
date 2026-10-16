package textview_test

import (
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/richtext"
	"github.com/ayn2op/tview/textview"
	"github.com/gdamore/tcell/v3"
)

// Styled text that wraps on words. Put it in a viewport to scroll it.
func ExampleNew() {
	var b richtext.Builder
	b.Write("Note: ", tcell.StyleDefault.Bold(true))
	b.Write("changes take effect after a restart.", tcell.StyleDefault)
	screentest.Print(textview.New(b.Finish()), 16, 3)
	// Output:
	// Note: changes
	// take effect
	// after a restart.
}
