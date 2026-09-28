package textview

// ScrollState is how far a text view is scrolled. The model owns it and applies the Changes a text view produces with Apply.
type ScrollState struct {
	row, column int
	// followEnd keeps the view scrolled to the last line as text is added.
	followEnd bool
}

// ScrollToEnd keeps the view scrolled to the last line, also as text is added, until the user scrolls up.
func (s *ScrollState) ScrollToEnd() {
	s.followEnd = true
}

// Change is an update to ScrollState produced by a text view, such as scrolling.
type Change struct {
	row, column int
	followEnd   bool
}

func (s *ScrollState) Apply(change Change) {
	s.row, s.column, s.followEnd = change.row, change.column, change.followEnd
}
