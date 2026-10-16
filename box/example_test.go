package box_test

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/box"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/text"
)

func ExampleNew() {
	b := box.New(text.New("hello")).
		Borders(tview.BordersAll).
		Title("Inbox").
		Footer("1 new").
		FooterAlignment(tview.AlignmentRight)
	screentest.Print(b, 12, 3)
	// Output:
	// ┌───Inbox──┐
	// │hello     │
	// └─────1 new┘
}
