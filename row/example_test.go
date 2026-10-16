package row_test

import (
	"github.com/ayn2op/tview/button"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/layout"
	"github.com/ayn2op/tview/row"
	"github.com/ayn2op/tview/text"
)

// A status bar with a label at the start and a button that as wide as its label and padding.
func ExampleNew() {
	bar := row.New(
		text.New("3 unread"),
		button.New().Label("Mark read").Width(layout.Fixed(11)),
	).Spacing(1)
	screentest.Print(bar, 24, 1)
	// Output:
	// 3 unread  Mark read
}
