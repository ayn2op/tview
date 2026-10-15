// Package spacer fills space with nothing, as between or beside other widgets.
package spacer

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/layout"
)

// Widget is an empty area that draws nothing and handles no messages.
type Widget struct {
	width, height layout.Length
}

var _ tview.Widget = Widget{}

// New returns a spacer that takes no space until it is given a width or height.
func New() Widget {
	return Widget{width: layout.Shrink, height: layout.Shrink}
}

// Width sets the width of the spacer.
func (w Widget) Width(width layout.Length) Widget {
	w.width = width
	return w
}

// Height sets the height of the spacer.
func (w Widget) Height(height layout.Length) Widget {
	w.height = height
	return w
}

// Size returns the width and height of the spacer.
func (w Widget) Size() (width, height layout.Length) {
	return w.width, w.height
}

// Layout returns the size of the spacer within limits.
func (w Widget) Layout(limits layout.Limits) layout.Size {
	return layout.Atomic(limits, w.width, w.height)
}

// Draw draws nothing.
func (Widget) Draw(tview.Screen, tview.Rectangle) {}

// Handle returns msg unchanged.
func (Widget) Handle(msg tview.Msg, _ tview.Rectangle) tview.Msg { return msg }
