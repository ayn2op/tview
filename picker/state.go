package picker

import (
	"github.com/ayn2op/tview/viewport"
)

// SearchState is a picker's query, the items that match it, and the selection among them. The model owns it and replaces it with the one a picker produces.
type SearchState struct {
	query string
	// queryCursor is the byte offset of the cursor in query.
	queryCursor int
	// cursor is the selected match.
	cursor int
	// scroll is how far the list is scrolled.
	scroll viewport.ScrollState
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
	s.query, s.queryCursor = "", 0
	s.cursor = 0
	s.scroll = viewport.ScrollState{}
	s.matches = nil
}

// Selected returns the selected item of items.
func (s *SearchState) Selected(items Items) (Item, bool) {
	i := s.cursor
	if i < 0 || i >= s.count(items) {
		return Item{}, false
	}
	return items[s.index(i)], true
}

// count returns the number of items that match the query.
func (s *SearchState) count(items Items) int {
	if s.query == "" {
		return len(items)
	}
	return len(s.matches)
}

// index returns the index in items of the ith match.
func (s *SearchState) index(i int) int {
	if s.query == "" {
		return i
	}
	return s.matches[i]
}
