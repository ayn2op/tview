package tree

import (
	"testing"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/vt"
)

func testModel() (*Model, []*Node) {
	nodes := []*Node{NewNode("root"), NewNode("one"), NewNode("two")}
	nodes[0].AddChild(nodes[1]).AddChild(nodes[2])
	return NewModel().SetRoot(nodes[0]).SetCurrentNode(nodes[0]), nodes
}

func TestUpdateMovesCursor(t *testing.T) {
	m, nodes := testModel()
	m.Update(tcell.NewEventKey(tcell.KeyDown, "", tcell.ModNone))
	if m.CurrentNode() != nodes[1] {
		t.Fatal("cursor did not move during Update")
	}
}

func TestSetCurrentNodeScrolls(t *testing.T) {
	m, nodes := testModel()
	m.SetRect(0, 0, 20, 2)
	m.SetCurrentNode(nodes[2])
	if m.GetScrollOffset() != 1 {
		t.Fatal("selected node is off-screen")
	}
}

func TestViewPreservesState(t *testing.T) {
	m, _ := testModel()
	m.SetRect(0, 0, 20, 2)
	m.Move(2)
	wantNode, wantOffset := m.CurrentNode(), m.GetScrollOffset()

	screen, err := tcell.NewTerminfoScreenFromTty(vt.NewMockTerm(vt.MockOptSize{X: 20, Y: 2}))
	if err != nil {
		t.Fatal(err)
	}
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(screen.Fini)

	m.View(screen)
	m.View(screen)
	if m.CurrentNode() != wantNode || m.GetScrollOffset() != wantOffset {
		t.Fatal("View mutated tree state")
	}
}
