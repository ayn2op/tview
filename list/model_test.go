package list

import (
	"testing"

	"github.com/ayn2op/tview"
)

type fixedHeightItem struct {
	*tview.Box
	height int
}

func (i *fixedHeightItem) Height(int) int {
	return i.height
}

func TestScrollBarMetricsUsesKnownContentHeight(t *testing.T) {
	builderCalls := 0
	model := NewModel().SetBuilder(func(index int) Item {
		builderCalls++
		return &fixedHeightItem{Box: tview.NewBox(), height: 1}
	})
	children := []drawnItem{{index: 0, row: 0, height: 1}}

	_, contentLength, viewportLength := model.scrollBarMetrics(20, 5, children, 10)

	if builderCalls != 0 {
		t.Fatalf("builder calls: got %d, want 0", builderCalls)
	}
	if contentLength != 10 || viewportLength != 5 {
		t.Fatalf("lengths: got (%d, %d), want (10, 5)", contentLength, viewportLength)
	}
}

var _ Item = (*fixedHeightItem)(nil)
