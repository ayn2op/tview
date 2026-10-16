package layout_test

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/box"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/layout"
	"github.com/ayn2op/tview/row"
	"github.com/ayn2op/tview/text"
)

// A sidebar of fixed width beside content that takes the rest, and a preview that takes half as much as the content.
func ExampleFillPortion() {
	panes := row.New(
		box.New(text.New("nav")).Borders(tview.BordersAll).Width(layout.Fixed(5)),
		box.New(text.New("body")).Borders(tview.BordersAll).Width(layout.FillPortion(2)),
		box.New(text.New("info")).Borders(tview.BordersAll).Width(layout.FillPortion(1)),
	)
	screentest.Print(panes, 23, 3)
	// Output:
	// ┌───┐┌──────────┐┌────┐
	// │nav││body      ││info│
	// └───┘└──────────┘└────┘
}
