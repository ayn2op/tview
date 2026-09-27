package picker

import (
	"github.com/ayn2op/tview"
	"github.com/gdamore/tcell/v3"
)

// entry is a list item that draws a single line of text.
type entry string

// Draw draws the text on the first row of area.
func (e entry) Draw(screen tview.Screen, area tview.Rectangle) {
	if area.Width <= 0 {
		return
	}
	tview.Print(screen, string(e), area.X, area.Y, area.Width, tview.AlignmentLeft, tcell.StyleDefault)
}

// Handle passes msg through unchanged.
func (entry) Handle(msg tview.Msg, _ tview.Rectangle) tview.Msg { return msg }

func (entry) Rows(int) int { return 1 }
