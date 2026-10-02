package list

// SelectionState holds a list's cursor and scroll position. The model owns it and applies the Changes a list produces with Apply.
type SelectionState struct {
	// cursor is the selected item, or -1 for none.
	cursor int
	// offset is the number of rows scrolled off the top.
	offset int
	// center asks the next view to scroll the cursor to the middle.
	center bool
	// atEnd reports whether the view was scrolled to the last row, and trackEnd whether it then stays there as items are added.
	atEnd, trackEnd bool
	// dragging is set while the scroll bar thumb is being dragged, which was grabbed grab cells from its top.
	dragging bool
	grab     int
}

// NewSelectionState returns a selection state with no item selected.
func NewSelectionState() SelectionState {
	return SelectionState{cursor: -1}
}

// Cursor returns the selected item, or -1 for none.
func (s *SelectionState) Cursor() int {
	return s.cursor
}

// SetCursor selects item index, or none if it is negative, and scrolls it to the middle of the view if it changed.
func (s *SelectionState) SetCursor(index int) {
	if index = max(index, -1); index != s.cursor {
		s.cursor, s.center, s.atEnd = index, true, false
	}
}

// SetTrackEnd sets whether the view stays scrolled to the last row as items are added while it is there.
func (s *SelectionState) SetTrackEnd(track bool) {
	s.trackEnd = track
}

// ScrollToEnd scrolls the view to the last row.
func (s *SelectionState) ScrollToEnd() {
	s.atEnd, s.center = true, false
}

// Change is an update to SelectionState produced by a list, such as moving the cursor or scrolling.
type Change struct {
	cursor, offset, grab int
	dragging             bool
	atEnd                bool
}

func (s *SelectionState) Apply(change Change) {
	s.cursor, s.offset, s.dragging, s.grab, s.atEnd = change.cursor, change.offset, change.dragging, change.grab, change.atEnd
	s.center = false
}
