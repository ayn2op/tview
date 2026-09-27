// Package center places an element in the middle of its area.
package center

import "github.com/ayn2op/tview"

// Widget places its child in the middle of its area.
type Widget struct {
	child tview.Element
}

var _ tview.Element = Widget{}

// New places child in the middle of its area at the child's Fixed size. A child that fills along an axis takes the whole area along it.
func New(child tview.Element) Widget {
	return Widget{child: child}
}

// Draw draws the child in the middle of area.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	w.child.Draw(screen, w.area(area))
}

// Handle passes msg to the child with the child's area.
func (w Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	return w.child.Handle(msg, w.area(area))
}

func (w Widget) area(area tview.Rectangle) tview.Rectangle {
	width, height := tview.SizeOf(w.child)
	if cells := width.Cells(); cells > 0 && cells < area.Width {
		area.X += (area.Width - cells) / 2
		area.Width = cells
	}
	if cells := height.Cells(); cells > 0 && cells < area.Height {
		area.Y += (area.Height - cells) / 2
		area.Height = cells
	}
	return area
}
