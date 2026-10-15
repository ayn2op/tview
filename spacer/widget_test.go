package spacer

import (
	"testing"

	"github.com/ayn2op/tview/layout"
)

func TestWidgetLayout(t *testing.T) {
	limits := layout.Limits{Max: layout.Size{Width: 10, Height: 5}}
	if got := New().Layout(limits); got != (layout.Size{}) {
		t.Fatalf("New().Layout = %+v, want no size", got)
	}
	if got := New().Width(layout.Fixed(3)).Height(layout.Fill).Layout(limits); got != (layout.Size{Width: 3, Height: 5}) {
		t.Fatalf("Layout = %+v, want 3x5", got)
	}
}
