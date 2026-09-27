// Package opaque keeps mouse messages from passing through an element to the ones stacked below it.
package opaque

import "github.com/ayn2op/tview"

// Widget drops the mouse messages within its area that its child does not turn into another message.
type Widget struct {
	child tview.Element
}

var _ tview.Element = Widget{}

// New wraps child so that mouse messages within its area do not reach the children of a stack below it.
func New(child tview.Element) Widget {
	return Widget{child: child}
}

// Size returns the size of the child.
func (w Widget) Size() (width, height tview.Length) {
	return tview.SizeOf(w.child)
}

// Draw draws the child.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	w.child.Draw(screen, area)
}

// Handle passes msg to the child and drops it if it is still a mouse message within area.
func (w Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	msg = w.child.Handle(msg, area)
	if mouse, ok := msg.(tview.MouseMsg); ok && area.Contains(mouse.Position()) {
		return nil
	}
	return msg
}
