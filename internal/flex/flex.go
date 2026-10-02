// Package flex lays out elements along one axis. It implements the row and column packages.
package flex

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/layout"
)

// Widget lays out its children along one axis. Children take their width and height from their Size.
type Widget struct {
	horizontal    bool
	width, height layout.Length
	spacing       int
	children      []tview.Element
}

var _ tview.Element = Widget{}

// New returns children laid out left to right if horizontal and top to bottom otherwise, skipping nil ones. It fills its parent in both directions by default.
func New(horizontal bool, children ...tview.Element) Widget {
	w := Widget{horizontal: horizontal, width: layout.Fill, height: layout.Fill}
	for _, child := range children {
		w = w.Push(child)
	}
	return w
}

// Push adds child after the others, unless it is nil.
func (w Widget) Push(child tview.Element) Widget {
	if child != nil {
		w.children = append(w.children[:len(w.children):len(w.children)], child)
	}
	return w
}

// Width sets the width of the layout.
func (w Widget) Width(width layout.Length) Widget {
	w.width = width
	return w
}

// Height sets the height of the layout.
func (w Widget) Height(height layout.Length) Widget {
	w.height = height
	return w
}

// Spacing sets the number of cells between children.
func (w Widget) Spacing(spacing int) Widget {
	w.spacing = spacing
	return w
}

// Size returns the width and height of the layout, resolving layout.Shrink to the size of its fixed-size children.
func (w Widget) Size() (width, height layout.Length) {
	width, height = w.width, w.height
	if width.IsShrink() || height.IsShrink() {
		contentWidth, contentHeight := w.contentSize()
		if width.IsShrink() {
			width = layout.Fixed(contentWidth)
		}
		if height.IsShrink() {
			height = layout.Fixed(contentHeight)
		}
	}
	return width, height
}

// Layout returns the size of the layout within limits.
func (w Widget) Layout(limits layout.Limits) layout.Size {
	width, height := w.Size()
	return layout.Atomic(limits, width, height)
}

// contentSize returns the space the fixed-size children need: their sum along the axis plus spacing, and their maximum across it.
func (w Widget) contentSize() (width, height int) {
	var along, across int
	for _, child := range w.children {
		childAlong, childAcross := w.axes(child)
		along += childAlong.Cells()
		across = max(across, childAcross.Cells())
	}
	along += w.spacing * max(len(w.children)-1, 0)
	if w.horizontal {
		return along, across
	}
	return across, along
}

// axes returns the length of child along the layout's axis and across it.
func (w Widget) axes(child tview.Element) (along, across layout.Length) {
	width, height := child.Size()
	if w.horizontal {
		return width, height
	}
	return height, width
}

// Draw draws each child in its area.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	for i, childArea := range w.areas(area) {
		w.children[i].Draw(screen, childArea)
	}
}

// Handle passes msg through each child in turn, with the child's area. It stops when a child drops the message.
func (w Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	for i, childArea := range w.areas(area) {
		if msg = w.children[i].Handle(msg, childArea); msg == nil {
			return nil
		}
	}
	return msg
}

// areas returns the area of each child within area.
func (w Widget) areas(area tview.Rectangle) []tview.Rectangle {
	length := area.Height
	if w.horizontal {
		length = area.Width
	}

	lengths := make([]layout.Length, len(w.children))
	free, portions := length-w.spacing*max(len(w.children)-1, 0), 0
	for i, child := range w.children {
		lengths[i], _ = w.axes(child)
		if lengths[i].Cells() > 0 {
			free -= lengths[i].Cells()
		} else {
			portions += lengths[i].Portion()
		}
	}
	free = max(free, 0)

	areas := make([]tview.Rectangle, len(w.children))
	pos := 0
	for i, childLength := range lengths {
		size := childLength.Cells()
		if size <= 0 && portions > 0 {
			// Portions are handed out from what is left so rounding never loses a cell.
			size = free * childLength.Portion() / portions
			free -= size
			portions -= childLength.Portion()
		}
		size = max(min(size, length-pos), 0)
		if w.horizontal {
			areas[i] = tview.Rectangle{X: area.X + pos, Y: area.Y, Width: size, Height: area.Height}
		} else {
			areas[i] = tview.Rectangle{X: area.X, Y: area.Y + pos, Width: area.Width, Height: size}
		}
		pos = min(pos+size+w.spacing, length)
	}
	return areas
}
