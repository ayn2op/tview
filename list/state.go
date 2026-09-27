package list

// SelectionState holds a list's cursor and scroll position. The model owns it and applies the Actions a list produces with Perform.
type SelectionState struct {
	// cursor is the selected item, or -1 for none.
	cursor int
	// offset is the number of rows scrolled off the top.
	offset int
	// center asks the next view to scroll the cursor to the middle.
	center bool
	// atEnd reports whether the view was scrolled to the last row, and trackEnd whether it then stays there as items are added.
	atEnd, trackEnd bool
	// grab is where the scroll bar thumb was grabbed, in subcells from its top, or -1 when it is not being dragged.
	grab int
}

// NewSelectionState returns a selection state with no item selected.
func NewSelectionState() SelectionState {
	return SelectionState{cursor: -1, grab: -1}
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

// Action is a change to SelectionState produced by a list, such as moving the cursor or scrolling.
type Action struct {
	cursor, offset, grab int
	atEnd                bool
}

// Perform applies action.
func (s *SelectionState) Perform(action Action) {
	s.cursor, s.offset, s.grab, s.atEnd = action.cursor, action.offset, action.grab, action.atEnd
	s.center = false
}
