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
		Title(text.New("Inbox")).
		Footer(text.New("1 new")).
		FooterAlignment(tview.HorizontalAlignmentRight)
	screentest.Print(b, 12, 3)
	// Output:
	// ┌───Inbox──┐
	// │hello     │
	// └─────1 new┘
}
