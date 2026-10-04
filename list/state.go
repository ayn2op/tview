package list

// SelectionState holds a list's cursor and scroll position. The model owns it and applies the Changes a list produces with Apply.
type SelectionState struct {
	// cursor is the selected item, or -1 for none.
	cursor int
	// offset is the number of rows scrolled off the top.
	offset int
	// center is the item the next view scrolls to the middle, or -1 for none. It outlives the cursor, so that deselecting leaves the view where it is.
	center int
	// atEnd keeps the view scrolled to the last row as items are added.
	atEnd bool
	// dragging is set while the scroll bar thumb is being dragged, which was grabbed grab cells from its top.
	dragging bool
	grab     int
}

// NewSelectionState returns a selection state with no item selected.
func NewSelectionState() SelectionState {
	return SelectionState{cursor: -1, center: -1}
}

// Cursor returns the selected item, or -1 for none.
func (s *SelectionState) Cursor() int {
	return s.cursor
}

// SetCursor selects item index and scrolls it to the middle of the view if it changed, or selects none if it is negative, which leaves the view where it is.
func (s *SelectionState) SetCursor(index int) {
	if index = max(index, -1); index == s.cursor {
		return
	}
	s.cursor = index
	if index >= 0 {
		s.center, s.atEnd = index, false
	}
}

// ScrollToEnd scrolls the view to the last row.
func (s *SelectionState) ScrollToEnd() {
	s.atEnd, s.center = true, -1
}

// Change is an update to SelectionState produced by a list, such as moving the cursor or scrolling.
type Change struct {
	cursor, offset, grab int
	dragging             bool
	atEnd                bool
}

func (s *SelectionState) Apply(change Change) {
	s.cursor, s.offset, s.dragging, s.grab, s.atEnd = change.cursor, change.offset, change.dragging, change.grab, change.atEnd
	s.center = -1
}
