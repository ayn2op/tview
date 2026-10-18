package inert_test

import (
	"fmt"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/button"
	"github.com/ayn2op/tview/inert"
	"github.com/ayn2op/tview/text"
	"github.com/gdamore/tcell/v3"
)

type sendMsg struct{}

// A button behind a dialog is still drawn, but enter no longer presses it.
func ExampleNew() {
	send := button.New(text.New("Send")).OnClick(sendMsg{}).Focused(true)
	area := tview.Rectangle{Width: 6, Height: 1}
	enter := tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone)

	fmt.Printf("%T\n", send.Handle(enter, area))
	fmt.Printf("%T\n", inert.New(send).Handle(enter, area))
	// Output:
	// inert_test.sendMsg
	// *tcell.EventKey
}
