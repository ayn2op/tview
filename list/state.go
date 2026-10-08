package list

// SelectionState holds a list's cursor. The model owns it and applies the Changes a list produces with Apply.
type SelectionState struct {
	// cursor is the selected item, or -1 for none.
	cursor int
	// target is the item selected last. It outlives the cursor, so that deselecting leaves a viewport where it is.
	target int
}

// NewSelectionState returns a selection state with no item selected.
func NewSelectionState() SelectionState {
	return SelectionState{cursor: -1, target: -1}
}

// Cursor returns the selected item, or -1 for none.
func (s *SelectionState) Cursor() int {
	return s.cursor
}

// SetCursor selects item index, or none if it is negative.
func (s *SelectionState) SetCursor(index int) {
	s.cursor = max(index, -1)
	if index >= 0 {
		s.target = index
	}
}

// Change is an update to SelectionState produced by a list.
type Change struct {
	cursor int
}

func (s *SelectionState) Apply(change Change) {
	s.SetCursor(change.cursor)
}
