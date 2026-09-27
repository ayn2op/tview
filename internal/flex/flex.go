// Package flex lays out elements along one axis. It implements the row and column packages.
package flex

import "github.com/ayn2op/tview"

// Layout lays out Children along one axis, left to right if Horizontal and top to bottom otherwise.
type Layout struct {
	Horizontal    bool
	Width, Height tview.Length
	Spacing       int
	Children      []tview.Element
}

// Size returns the width and height of the layout, resolving Shrink to the size of the fixed-size children.
func (l Layout) Size() (width, height tview.Length) {
	width, height = l.Width, l.Height
	if width.IsShrink() || height.IsShrink() {
		contentWidth, contentHeight := l.contentSize()
		if width.IsShrink() {
			width = tview.Fixed(contentWidth)
		}
		if height.IsShrink() {
			height = tview.Fixed(contentHeight)
		}
	}
	return width, height
}

// contentSize returns the space the fixed-size children need: their sum along the axis plus spacing, and their maximum across it.
func (l Layout) contentSize() (width, height int) {
	var along, across int
	for _, child := range l.Children {
		childAlong, childAcross := l.axes(child)
		along += childAlong.Cells()
		across = max(across, childAcross.Cells())
	}
	along += l.Spacing * max(len(l.Children)-1, 0)
	if l.Horizontal {
		return along, across
	}
	return across, along
}

// axes returns the length of child along the layout's axis and across it.
func (l Layout) axes(child tview.Element) (along, across tview.Length) {
	width, height := tview.SizeOf(child)
	if l.Horizontal {
		return width, height
	}
	return height, width
}

// Draw draws each child in its area.
func (l Layout) Draw(screen tview.Screen, area tview.Rectangle) {
	for i, childArea := range l.Areas(area) {
		l.Children[i].Draw(screen, childArea)
	}
}

// Handle passes msg through each child in turn, with the child's area. It stops when a child drops the message.
func (l Layout) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	for i, childArea := range l.Areas(area) {
		if msg = l.Children[i].Handle(msg, childArea); msg == nil {
			return nil
		}
	}
	return msg
}

// Areas returns the area of each child within area.
func (l Layout) Areas(area tview.Rectangle) []tview.Rectangle {
	length := area.Height
	if l.Horizontal {
		length = area.Width
	}

	lengths := make([]tview.Length, len(l.Children))
	free, portions := length-l.Spacing*max(len(l.Children)-1, 0), 0
	for i, child := range l.Children {
		lengths[i], _ = l.axes(child)
		if lengths[i].Cells() > 0 {
			free -= lengths[i].Cells()
		} else {
			portions += lengths[i].Portion()
		}
	}
	free = max(free, 0)

	areas := make([]tview.Rectangle, len(l.Children))
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
		if l.Horizontal {
			areas[i] = tview.Rectangle{X: area.X + pos, Y: area.Y, Width: size, Height: area.Height}
		} else {
			areas[i] = tview.Rectangle{X: area.X, Y: area.Y + pos, Width: area.Width, Height: size}
		}
		pos = min(pos+size+l.Spacing, length)
	}
	return areas
}
