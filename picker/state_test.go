package picker

import (
	"slices"
	"strconv"
	"testing"
)

func TestSearchStateVisible(t *testing.T) {
	var items Items
	for i := range 10 {
		items = append(items, Item{FilterText: strconv.Itoa(i)})
	}

	for _, test := range []struct {
		name   string
		items  Items
		cursor int
		want   []string
	}{
		{"top", items, 0, []string{"0", "1", "2"}},
		{"centered", items, 5, []string{"4", "5", "6"}},
		{"bottom", items, 9, []string{"7", "8", "9"}},
		{"fewer items than rows", items[:2], 0, []string{"0", "1"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			s := NewSearchState()
			s.Select(test.cursor)
			var got []string
			for _, item := range s.Visible(test.items, 3) {
				got = append(got, item.FilterText)
			}
			if !slices.Equal(got, test.want) {
				t.Errorf("Visible = %v, want %v", got, test.want)
			}
		})
	}
}
