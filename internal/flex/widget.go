package flex

import "github.com/ayn2op/tview"

// Widget lays out its children along one axis. Children take their width and height from tview.SizeOf.
type Widget struct {
	layout Layout
}

var _ tview.Element = Widget{}

// New returns children laid out left to right if horizontal and top to bottom otherwise, skipping nil ones. It fills its parent in both directions by default.
func New(horizontal bool, children ...tview.Element) Widget {
	w := Widget{layout: Layout{Horizontal: horizontal, Width: tview.Fill, Height: tview.Fill}}
	for _, child := range children {
		w = w.Push(child)
	}
	return w
}

// Push adds child after the others, unless it is nil.
func (w Widget) Push(child tview.Element) Widget {
	if child != nil {
		w.layout.Children = append(w.layout.Children[:len(w.layout.Children):len(w.layout.Children)], child)
	}
	return w
}

// Width sets the width of the layout.
func (w Widget) Width(width tview.Length) Widget {
	w.layout.Width = width
	return w
}

// Height sets the height of the layout.
func (w Widget) Height(height tview.Length) Widget {
	w.layout.Height = height
	return w
}

// Spacing sets the number of cells between children.
func (w Widget) Spacing(spacing int) Widget {
	w.layout.Spacing = spacing
	return w
}

// Size returns the width and height of the layout, resolving tview.Shrink to the size of its fixed-size children.
func (w Widget) Size() (width, height tview.Length) {
	return w.layout.Size()
}

// Draw draws each child in its area.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	w.layout.Draw(screen, area)
}

// Handle passes msg through each child in turn, with the child's area. It stops when a child drops the message.
func (w Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	return w.layout.Handle(msg, area)
}
