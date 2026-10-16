package center_test

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/box"
	"github.com/ayn2op/tview/center"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/layout"
	"github.com/ayn2op/tview/text"
)

// A box that shrinks to fit its text, in the middle of a window.
func ExampleNew() {
	b := box.New(text.New("loading")).Borders(tview.BordersAll).Width(layout.Shrink).Height(layout.Shrink)
	window := box.New(center.New(b)).Borders(tview.BordersAll)
	screentest.Print(window, 15, 7)
	// Output:
	// ┌─────────────┐
	// │             │
	// │  ┌───────┐  │
	// │  │loading│  │
	// │  └───────┘  │
	// │             │
	// └─────────────┘
}
