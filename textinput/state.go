package textinput

// EditState holds a text input's value, cursor, and how far it is scrolled. The model owns it and applies the Changes a text input produces with Apply.
type EditState struct {
	value string
	// cursor is the byte offset of the cursor in value, always at a grapheme boundary.
	cursor int
	// offset is the number of cells scrolled off the left edge.
	offset int
}

// NewEditState returns an edit state holding value with the cursor at its end.
func NewEditState(value string) EditState {
	return EditState{value: value, cursor: len(value)}
}

// Value returns the text.
func (s *EditState) Value() string {
	return s.value
}

// SetValue replaces the text and moves the cursor to its end.
func (s *EditState) SetValue(value string) {
	s.value, s.cursor = value, len(value)
}

// Change is an update to EditState produced by a text input, such as typing or moving the cursor.
type Change struct {
	value          string
	cursor, offset int
}

func (s *EditState) Apply(change Change) {
	s.value, s.cursor, s.offset = change.value, change.cursor, change.offset
}
