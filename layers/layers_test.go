package layers

import (
	"slices"
	"testing"

	"github.com/ayn2op/tview"
)

func TestLayerOrder(t *testing.T) {
	layers := New()
	for _, name := range []string{"a", "b", "c"} {
		layers.AddLayer(tview.NewBox(), WithName(name))
	}

	assertNames := func(want ...string) {
		t.Helper()
		if got := layers.GetLayerNames(false); !slices.Equal(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	}

	assertNames("c", "b", "a")
	layers.SendToFront("a")
	assertNames("a", "c", "b")
	layers.SendToBack("a")
	assertNames("c", "b", "a")
	layers.RemoveLayer("b")
	assertNames("c", "a")
	layers.AddLayer(tview.NewBox(), WithName("a"))
	assertNames("a", "c")
}
