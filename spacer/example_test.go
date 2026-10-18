package spacer_test

import (
	"github.com/ayn2op/tview/button"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/layout"
	"github.com/ayn2op/tview/row"
	"github.com/ayn2op/tview/spacer"
	"github.com/ayn2op/tview/text"
)

// A spacer that fills the width pushes the buttons after it to the right edge, away from the title.
func ExampleNew() {
	buttons := row.New(
		text.New("Draft"),
		spacer.New().Width(layout.Fill),
		button.New(text.New("Cancel")).Width(layout.Fixed(8)),
		button.New(text.New("OK")).Width(layout.Fixed(4)),
	).Spacing(1)
	screentest.Print(buttons, 24, 1)
	// Output:
	// Draft       Cancel   OK
}
