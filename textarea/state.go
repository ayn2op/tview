package textarea

// EditState holds a text area's value, cursor, how far it is scrolled, and earlier values for Undo. The model owns it and applies the Changes a text area produces with Apply.
type EditState struct {
	value string
	// cursor is the byte offset of the cursor in value, always at a grapheme boundary.
	cursor int
	// row is the first visible wrapped line.
	row     int
	history []snapshot
}

type snapshot struct {
	value  string
	cursor int
}

// NewEditState returns an edit state holding value with the cursor at its end.
func NewEditState(value string) EditState {
	return EditState{value: value, cursor: len(value)}
}

// Value returns the text.
func (s *EditState) Value() string {
	return s.value
}

// Cursor returns the byte offset of the cursor in the value.
func (s *EditState) Cursor() int {
	return s.cursor
}

// SetValue replaces the text and moves the cursor to its end.
func (s *EditState) SetValue(value string) {
	s.Replace(0, len(s.value), value)
}

// Replace replaces the bytes from start to end of the value with text and moves the cursor after it.
func (s *EditState) Replace(start, end int, text string) {
	s.save()
	s.value, s.cursor = s.value[:start]+text+s.value[end:], start+len(text)
}

// Undo restores the value and cursor from before the last change.
func (s *EditState) Undo() {
	if len(s.history) == 0 {
		return
	}
	last := s.history[len(s.history)-1]
	s.history = s.history[:len(s.history)-1]
	s.value, s.cursor = last.value, last.cursor
}

func (s *EditState) save() {
	s.history = append(s.history, snapshot{value: s.value, cursor: s.cursor})
}

// Change is an update to EditState produced by a text area, such as typing or moving the cursor.
type Change struct {
	value       string
	cursor, row int
}

// Apply applies change, remembering the previous value for Undo if it changes.
func (s *EditState) Apply(change Change) {
	if change.value != s.value {
		s.save()
	}
	s.value, s.cursor, s.row = change.value, change.cursor, change.row
}
