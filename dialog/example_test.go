package dialog_test

import (
	"fmt"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/dialog"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/gdamore/tcell/v3"
)

type doneMsg struct {
	index int
	label string
}

// A confirmation whose OnDone message tells Update which button was pressed.
func ExampleNew() {
	confirm := dialog.New().
		Text("Leave the server?").
		Buttons("Leave", "Cancel").
		Focus(1).
		OnDone(func(index int, label string) tview.Msg { return doneMsg{index, label} })
	screentest.Print(confirm, 26, 7)

	enter := tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone)
	fmt.Printf("%+v\n", confirm.Handle(enter, tview.Rectangle{Width: 26, Height: 7}))
	// Output:
	// ┌────────────────────────┐
	// │                        │
	// │   Leave the server?    │
	// │                        │
	// │   Leave      Cancel    │
	// │                        │
	// └────────────────────────┘
	// {index:1 label:Cancel}
}
