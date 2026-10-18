package opaque_test

import (
	"fmt"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/box"
	"github.com/ayn2op/tview/button"
	"github.com/ayn2op/tview/opaque"
	"github.com/ayn2op/tview/stack"
	"github.com/ayn2op/tview/text"
	"github.com/gdamore/tcell/v3"
)

type deleteMsg struct{}

// A click on a panel stacked over a button would reach the button below it, unless the panel is opaque.
func ExampleNew() {
	del := button.New(text.New("Delete")).OnClick(deleteMsg{})
	panel := box.New(text.New("notes")).Borders(tview.BordersAll)
	area := tview.Rectangle{Width: 10, Height: 3}
	click := tview.MouseMsg{EventMouse: tcell.NewEventMouse(4, 1, tcell.ButtonPrimary, tcell.ModNone), Action: tview.MouseLeftDown}

	fmt.Printf("%T\n", stack.New(del, panel).Handle(click, area))
	fmt.Printf("%v\n", stack.New(del, opaque.New(panel)).Handle(click, area))
	// Output:
	// opaque_test.deleteMsg
	// <nil>
}
