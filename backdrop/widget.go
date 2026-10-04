// Package backdrop restyles what is drawn behind it, typically to dim the content behind a dialog.
package backdrop

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/layout"
	"github.com/gdamore/tcell/v3"
)

// Widget merges its style into everything already drawn within its area and hides the cursor.
type Widget struct {
	style tview.Style
}

var _ tview.Widget = Widget{}

// New returns a backdrop that dims what is behind it.
func New() Widget {
	return Widget{style: tcell.StyleDefault.Dim(true)}
}

// Style sets the style merged into the cells behind the backdrop with tview.MergeStyle.
func (w Widget) Style(style tview.Style) Widget {
	w.style = style
	return w
}

// Size returns Fill, as the backdrop takes its whole area.
func (Widget) Size() (width, height layout.Length) {
	return layout.Fill, layout.Fill
}

// Layout returns the size of limits, as the backdrop takes its whole area.
func (Widget) Layout(limits layout.Limits) layout.Size {
	return layout.Atomic(limits, layout.Fill, layout.Fill)
}

// Draw restyles the cells already drawn within area.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	screen.HideCursor()
	for y := area.Y; y < area.Y+area.Height; y++ {
		for x := area.X; x < area.X+area.Width; x++ {
			str, style, width := screen.Get(x, y)
			screen.Put(x, y, str, tview.MergeStyle(style, w.style))
			// Wide characters occupy the following cells too.
			x += max(width, 1) - 1
		}
	}
}

// Handle passes msg through unchanged.
func (Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	return msg
}
