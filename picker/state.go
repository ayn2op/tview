package picker

import (
	"github.com/ayn2op/tview/list"
	"github.com/ayn2op/tview/textinput"
)

// SearchState is a picker's query, the items that match it, and the selection among them. The model owns it and applies the Changes a picker produces with Apply.
type SearchState struct {
	query textinput.EditState
	list  list.SelectionState
	// matches are the indexes of the items that match the query, used while the query is not empty.
	matches []int
}

// NewSearchState returns an empty query with the first item selected.
func NewSearchState() SearchState {
	var s SearchState
	s.Reset()
	return s
}

// Reset clears the query and selects the first item. Call it when the items change, since matches refer to them by index.
func (s *SearchState) Reset() {
	s.query = textinput.EditState{}
	s.list = list.NewSelectionState()
	s.list.SetCursor(0)
	s.matches = nil
}

// Selected returns the selected item of items.
func (s *SearchState) Selected(items Items) (Item, bool) {
	i := s.list.Cursor()
	if i < 0 || i >= s.count(items) {
		return Item{}, false
	}
	return items[s.index(i)], true
}

// count returns the number of items that match the query.
func (s *SearchState) count(items Items) int {
	if s.query.Value() == "" {
		return len(items)
	}
	return len(s.matches)
}

// index returns the index in items of the ith match.
func (s *SearchState) index(i int) int {
	if s.query.Value() == "" {
		return i
	}
	return s.matches[i]
}

// Change is an update to SearchState produced by a picker, such as typing in the query or moving the selection.
type Change struct {
	query   textinput.Change
	matches []int
	list    list.Change
	isQuery bool
	// filtered is set when the query changed, so matches and cursor are for the new query.
	filtered bool
	cursor   int
}

func (s *SearchState) Apply(change Change) {
	if !change.isQuery {
		s.list.Apply(change.list)
		return
	}
	s.query.Apply(change.query)
	if change.filtered {
		s.matches = change.matches
		s.list.SetCursor(change.cursor)
	}
}
