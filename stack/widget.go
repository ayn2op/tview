// Package stack draws elements on top of each other, for overlays such as modal dialogs.
package stack

import "github.com/ayn2op/tview"

// Widget draws its children on top of each other in the same area.
type Widget struct {
	children []tview.Element
}

var _ tview.Element = Widget{}

// New returns a stack of children, the first at the bottom, skipping nil ones.
func New(children ...tview.Element) Widget {
	var w Widget
	for _, child := range children {
		w = w.Push(child)
	}
	return w
}

// Push adds child on top of the others, unless it is nil.
func (w Widget) Push(child tview.Element) Widget {
	if child != nil {
		w.children = append(w.children[:len(w.children):len(w.children)], child)
	}
	return w
}

// Draw draws the children from bottom to top.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	for _, child := range w.children {
		child.Draw(screen, area)
	}
}

// Handle passes msg through the children from top to bottom. It stops when a child drops the message.
func (w Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	for i := len(w.children) - 1; i >= 0; i-- {
		if msg = w.children[i].Handle(msg, area); msg == nil {
			return nil
		}
	}
	return msg
}
