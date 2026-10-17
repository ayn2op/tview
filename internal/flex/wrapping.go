package flex

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/layout"
)

// Wrapping lays out the children of a layout along its axis, starting a new line where the next would overflow it.
// Within a line, children that do not fill it across the axis go where the layout's Align puts them.
type Wrapping struct {
	flex        Widget
	lineSpacing int
	alignLines  tview.Alignment
}

var _ tview.Widget = Wrapping{}

// Wrap returns the layout wrapping its children, with as many cells between lines as between children.
func (w Widget) Wrap() Wrapping {
	return Wrapping{flex: w, lineSpacing: w.spacing}
}

// LineSpacing sets the number of cells between lines.
func (w Wrapping) LineSpacing(spacing int) Wrapping {
	w.lineSpacing = spacing
	return w
}

// AlignLines sets where each line goes along the axis, within the longest line: at its start unless set.
func (w Wrapping) AlignLines(alignment tview.Alignment) Wrapping {
	w.alignLines = alignment
	return w
}

// Size returns the width and height of the layout.
func (w Wrapping) Size() (width, height layout.Length) {
	return w.flex.Size()
}

// Layout returns the size of the layout within limits, laying out its children where it shrinks to them.
func (w Wrapping) Layout(limits layout.Limits) layout.Size {
	return layout.Sized(limits, w.flex.width, w.flex.height, func(limits layout.Limits) layout.Size {
		var size layout.Size
		for _, area := range w.areas(0, 0, limits) {
			size.Width, size.Height = max(size.Width, area.X+area.Width), max(size.Height, area.Y+area.Height)
		}
		return size
	})
}

func (w Wrapping) Draw(screen tview.Screen, area tview.Rectangle) {
	for i, childArea := range w.areas(area.X, area.Y, layout.Limits{Max: layout.Size{Width: area.Width, Height: area.Height}}) {
		w.flex.children[i].Draw(screen, childArea)
	}
}

// Handle passes msg through each child in turn, with the child's area. It stops when a child drops the message.
func (w Wrapping) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	for i, childArea := range w.areas(area.X, area.Y, layout.Limits{Max: layout.Size{Width: area.Width, Height: area.Height}}) {
		if msg = w.flex.children[i].Handle(msg, childArea); msg == nil {
			return nil
		}
	}
	return msg
}

// areas returns the area of each child laid out within limits, from x and y.
// Children are laid out within the same limits, so they keep an axis that is infinite, as along the one a parent scrolls.
func (w Wrapping) areas(x, y int, limits layout.Limits) []tview.Rectangle {
	limits = limits.Loose()
	bounds := limits.Bounds()
	available, _ := pack(w.flex.horizontal, bounds.Width, bounds.Height)

	// line is the children from start to end, the length they take along the axis and the widest of them across it.
	type line struct{ start, end, main, cross int }
	var lines []line
	var current, longest int
	sizes := make([]layout.Size, len(w.flex.children))
	for i, child := range w.flex.children {
		sizes[i] = child.Layout(limits)
		main, cross := pack(w.flex.horizontal, sizes[i].Width, sizes[i].Height)
		if len(lines) == 0 || lines[current].end > lines[current].start && lines[current].main+w.flex.spacing+main > available {
			lines = append(lines, line{start: i, end: i})
			current = len(lines) - 1
		}

		l := &lines[current]
		if l.end > l.start {
			l.main += w.flex.spacing
		}
		l.end, l.main, l.cross = i+1, l.main+main, max(l.cross, cross)
		longest = max(longest, l.main)
	}

	areas := make([]tview.Rectangle, len(w.flex.children))
	along := 0
	for _, l := range lines {
		pos := offset(longest-l.main, w.alignLines)
		for i := l.start; i < l.end; i++ {
			main, cross := pack(w.flex.horizontal, sizes[i].Width, sizes[i].Height)
			childX, childY := pack(w.flex.horizontal, pos, along+offset(l.cross-cross, w.flex.alignment))
			areas[i] = tview.Rectangle{X: x + childX, Y: y + childY, Width: sizes[i].Width, Height: sizes[i].Height}
			pos += main + w.flex.spacing
		}

		along += l.cross + w.lineSpacing
	}

	return areas
}
