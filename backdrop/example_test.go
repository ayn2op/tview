package backdrop_test

import (
	"github.com/ayn2op/tview/backdrop"
	"github.com/ayn2op/tview/dialog"
	"github.com/ayn2op/tview/inert"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/stack"
	"github.com/ayn2op/tview/text"
)

// A modal dialog: the content behind it is dimmed by the backdrop and takes no input, being inert.
func ExampleNew() {
	content := text.New("draft: hello world")
	modal := stack.New(
		inert.New(content),
		backdrop.New(),
		dialog.New().Text("Discard draft?").Buttons("Discard", "Keep"),
	)
	screentest.Print(modal, 30, 11)
	// Output:
	// draft: hello world
	//
	// ┌────────────────────────────┐
	// │                            │
	// │       Discard draft?       │
	// │                            │
	// │     Discard      Keep      │
	// │                            │
	// └────────────────────────────┘
}
