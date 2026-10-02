// Package inert draws an element without letting it take input, like HTML's inert attribute, such as for the content behind a dialog.
package inert

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/layout"
)

// Widget draws its child but passes every message by it unchanged.
type Widget struct {
	child tview.Element
}

var _ tview.Element = Widget{}

// New wraps child so that it is drawn but takes no input.
func New(child tview.Element) Widget {
	return Widget{child: child}
}

// Size returns the size of the child.
func (w Widget) Size() (width, height tview.Length) {
	return tview.SizeOf(w.child)
}

// Layout lays out the child.
func (w Widget) Layout(limits layout.Limits) tview.Size {
	return w.child.Layout(limits)
}

// Draw draws the child.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	w.child.Draw(screen, area)
}

// Handle returns msg without passing it to the child.
func (Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	return msg
}
