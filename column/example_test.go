package column_test

import (
	"github.com/ayn2op/tview/column"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/text"
	"github.com/ayn2op/tview/textinput"
)

// A form of labeled inputs, one under another with a row between them.
func ExampleNew() {
	form := column.New(
		text.New("Email"),
		textinput.New("me@example.com", 0),
		text.New("Password"),
		textinput.New("hunter2", 0).Mask("*"),
	).Spacing(1)
	screentest.Print(form, 16, 7)
	// Output:
	// Email
	//
	// me@example.com
	//
	// Password
	//
	// *******
}
