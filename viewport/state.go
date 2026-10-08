package viewport

// ScrollState is how far a viewport is scrolled. The model owns it and applies the Changes a viewport produces with Apply.
type ScrollState struct {
	x, y int
	// toTarget keeps the target in the middle, and atEnd the last row at the bottom.
	toTarget, atEnd bool
	// dragging is set while the scroll bar thumb is dragged, which was grabbed grab cells from its top.
	dragging bool
	grab     int
}

// ScrollToTarget keeps the target of the viewport in the middle until the user scrolls.
func (s *ScrollState) ScrollToTarget() {
	s.toTarget, s.atEnd = true, false
}

// ScrollToEnd scrolls to the last row.
func (s *ScrollState) ScrollToEnd() {
	s.toTarget, s.atEnd = false, true
}

// Change is an update to ScrollState produced by a viewport.
type Change struct {
	x, y, grab                int
	dragging, toTarget, atEnd bool
}

func (s *ScrollState) Apply(change Change) {
	*s = ScrollState{x: change.x, y: change.y, toTarget: change.toTarget, atEnd: change.atEnd, dragging: change.dragging, grab: change.grab}
}
