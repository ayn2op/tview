package grid

import (
	"slices"
	"testing"
)

func TestLayoutAxis(t *testing.T) {
	for _, tt := range []struct {
		name                           string
		defs                           []int
		count, available, minimum, gap int
		bordered                       bool
		wantPos, wantSizes             []int
	}{
		{"fixed proportional implicit gaps", []int{30, 0, -2}, 4, 100, 0, 2, false, []int{0, 32, 50, 84}, []int{30, 16, 32, 16}},
		{"borders", []int{10}, 2, 20, 0, 5, true, []int{1, 12}, []int{10, 7}},
		{"minimum fixed track", []int{5, 0}, 2, 30, 10, 0, false, []int{0, 10}, []int{10, 20}},
	} {
		pos, sizes := layoutAxis(tt.defs, tt.count, tt.available, tt.minimum, tt.gap, tt.bordered)
		if !slices.Equal(pos, tt.wantPos) || !slices.Equal(sizes, tt.wantSizes) {
			t.Fatalf("%s: got pos=%v sizes=%v, want pos=%v sizes=%v", tt.name, pos, sizes, tt.wantPos, tt.wantSizes)
		}
	}
}
