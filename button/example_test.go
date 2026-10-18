package button_test

import (
	"fmt"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/button"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/text"
	"github.com/gdamore/tcell/v3"
)

type saveMsg struct{}

// A focused button turns enter into its OnClick message, which Update receives.
func ExampleNew() {
	save := button.New(text.New("Save")).OnClick(saveMsg{}).Focused(true)
	screentest.Print(save, 8, 1)

	enter := tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone)
	msg := save.Handle(enter, tview.Rectangle{Width: 8, Height: 1})
	fmt.Printf("%T\n", msg)
	// Output:
	// Save
	// button_test.saveMsg
}
