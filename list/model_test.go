package list

import (
	"testing"

	"github.com/ayn2op/tview"
	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/vt"
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

func TestViewIsReadOnly(t *testing.T) {
	model := testList(5).ScrollDown()
	model.SetRect(0, 0, 20, 2)
	if model.scroll.top != 1 || model.scroll.pending != 0 {
		t.Fatalf("scroll state: %+v", model.scroll)
	}
	wantScroll, wantEnd := model.scroll, model.atEnd
	screen, err := tcell.NewTerminfoScreenFromTty(vt.NewMockTerm(vt.MockOptSize{X: 20, Y: 2}))
	if err != nil {
		t.Fatal(err)
	}
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(screen.Fini)

	model.View(screen)
	if model.scroll != wantScroll || model.atEnd != wantEnd {
		t.Fatal("View mutated list state")
	}
}

func TestCursorChangedMsg(t *testing.T) {
	cmd := testList(2).SetCursor(0).Update(tcell.NewEventKey(tcell.KeyDown, "", tcell.ModNone))
	msg, ok := cmd().(CursorChangedMsg)
	if !ok || msg.Index != 1 {
		t.Fatalf("got %#v", msg)
	}
}

func testList(count int) *Model {
	model := NewModel().SetScrollBarVisibility(ScrollBarVisibilityNever)
	model.SetBuilder(func(index int) Item {
		if index >= count {
			return nil
		}
		return &fixedHeightItem{Box: tview.NewBox(), height: 1}
	})
	model.SetRect(0, 0, 20, 2)
	return model
}

var _ Item = (*fixedHeightItem)(nil)
