package textview

// ScrollState is how far a text view is scrolled. The model owns it and applies the Actions a text view produces with Perform.
type ScrollState struct {
	row, column int
	// followEnd keeps the view scrolled to the last line as text is added.
	followEnd bool
}

// ScrollToEnd keeps the view scrolled to the last line, also as text is added, until the user scrolls up.
func (s *ScrollState) ScrollToEnd() {
	s.followEnd = true
}

// Action is a change to ScrollState produced by a text view, such as scrolling.
type Action struct {
	row, column int
	followEnd   bool
}

// Perform applies action.
func (s *ScrollState) Perform(action Action) {
	s.row, s.column, s.followEnd = action.row, action.column, action.followEnd
}
