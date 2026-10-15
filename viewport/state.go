package viewport

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/layout"
)

// ScrollState is how far a viewport is scrolled. The model owns it and replaces it with the one a viewport produces.
type ScrollState struct {
	// x and y are how far the viewport is scrolled from its Anchor on each axis.
	x, y int
	// toTarget keeps the target in the middle, and atEnd the last row at the bottom.
	toTarget, atEnd bool
	// dragging is set while the scroll bar thumb is dragged, which was grabbed grab cells from its top.
	dragging bool
	grab     int
	// viewport is what the viewport showed when it produced the state, and origin what it showed before.
	viewport, origin Viewport
}

// ScrollToTarget keeps the target of the viewport in the middle until the user scrolls.
func (s *ScrollState) ScrollToTarget() {
	s.toTarget, s.atEnd = true, false
}

// ScrollToEnd scrolls to the last row.
func (s *ScrollState) ScrollToEnd() {
	s.toTarget, s.atEnd = false, true
}

// Viewport returns what the viewport showed when it produced the state, or the zero Viewport if a viewport did not produce it.
func (s ScrollState) Viewport() Viewport {
	return s.viewport
}

// Origin returns what the viewport showed before it produced the state, such as before scrolling, or the zero Viewport if a viewport did not produce it.
func (s ScrollState) Origin() Viewport {
	return s.origin
}

// Viewport is the part of its child a viewport shows.
type Viewport struct {
	// x and y are how far the viewport is scrolled from its Anchor on each axis.
	x, y int
	// bounds is the area the child shows in, and content the size of the child.
	bounds  tview.Rectangle
	content layout.Size
}

// AbsoluteOffset is how far a viewport is scrolled, in cells.
type AbsoluteOffset struct {
	X, Y int
}

// RelativeOffset is how far a viewport is scrolled through its child, from 0 at the start to 1 at the end.
type RelativeOffset struct {
	X, Y float64
}

// Bounds returns the area the child shows in.
func (v Viewport) Bounds() tview.Rectangle {
	return v.bounds
}

// Content returns the size of the child.
func (v Viewport) Content() layout.Size {
	return v.content
}

// End returns the offset of the end of the viewport, the furthest it scrolls from its Anchor.
func (v Viewport) End() AbsoluteOffset {
	return AbsoluteOffset{X: max(v.content.Width-v.bounds.Width, 0), Y: max(v.content.Height-v.bounds.Height, 0)}
}

// DistanceToEnd returns how much of the child is left to scroll into view away from the Anchor, which is zero along an axis where it fits or is scrolled to the end.
func (v Viewport) DistanceToEnd() AbsoluteOffset {
	end := v.End()
	return AbsoluteOffset{X: end.X - v.x, Y: end.Y - v.y}
}

// AbsoluteOffset returns how far the viewport is scrolled from its Anchor.
func (v Viewport) AbsoluteOffset() AbsoluteOffset {
	return AbsoluteOffset{X: v.x, Y: v.y}
}

// AbsoluteOffsetReversed returns how far the viewport is scrolled from the edge opposite its Anchor, which keeps the scroll position when switching the Anchor.
func (v Viewport) AbsoluteOffsetReversed() AbsoluteOffset {
	return v.DistanceToEnd()
}

// RelativeOffset returns how far the viewport is scrolled through its child from its Anchor, which is 0 along an axis where it fits.
func (v Viewport) RelativeOffset() RelativeOffset {
	var offset RelativeOffset
	if end := v.End(); end.X > 0 {
		offset.X = float64(v.x) / float64(end.X)
	}
	if end := v.End(); end.Y > 0 {
		offset.Y = float64(v.y) / float64(end.Y)
	}
	return offset
}
