package tree

import (
	"slices"
	"testing"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/gdamore/tcell/v3"
)

// testTree returns the nodes of a root with children a and b, where a has child a1, all expanded.
func testTree() map[string]*Node {
	nodes := map[string]*Node{"root": NewNode("root"), "a": NewNode("a"), "a1": NewNode("a1"), "b": NewNode("b")}
	nodes["root"].AddChild(nodes["a"]).AddChild(nodes["b"])
	nodes["a"].AddChild(nodes["a1"])
	return nodes
}

func interactive(root *Node, selectionState *SelectionState) Widget {
	return New(root, selectionState).TopLevel(1).Markers(Markers{}).Focused(true).OnAction(func(a Action) tview.Msg { return a })
}

func TestWidgetDraw(t *testing.T) {
	screen := screentest.New(t, 6, 3)
	var selectionState SelectionState
	interactive(testTree()["root"], &selectionState).Draw(screen, tview.Rectangle{Width: 6, Height: 3})
	for y, want := range []string{"a     ", "└──a1 ", "b     "} {
		if got := screentest.Row(screen, y, 6); got != want {
			t.Fatalf("row %d = %q, want %q", y, got, want)
		}
	}
}

func TestWidgetHandle(t *testing.T) {
	area := tview.Rectangle{Width: 6, Height: 3}
	key := func(k tcell.Key, str string) tview.KeyMsg { return tcell.NewEventKey(k, str, tcell.ModNone) }
	// send passes msg through the tree and applies the Action it produces, if any, returning the other result.
	send := func(nodes map[string]*Node, selectionState *SelectionState, msg tview.Msg) tview.Msg {
		out := interactive(nodes["root"], selectionState).Handle(msg, area)
		if action, ok := out.(Action); ok {
			selectionState.Perform(action)
			return nil
		}
		return out
	}

	t.Run("down selects the next node", func(t *testing.T) {
		nodes := testTree()
		var selectionState SelectionState
		selectionState.SetCurrentNode(nodes["a"])
		send(nodes, &selectionState, key(tcell.KeyDown, ""))
		if selectionState.CurrentNode() != nodes["a1"] {
			t.Fatalf("current = %v", selectionState.CurrentNode().Line())
		}
	})
	t.Run("move to parent", func(t *testing.T) {
		nodes := testTree()
		var selectionState SelectionState
		selectionState.SetCurrentNode(nodes["a1"])
		send(nodes, &selectionState, key(tcell.KeyRune, "K"))
		if selectionState.CurrentNode() != nodes["a"] {
			t.Fatalf("current = %v", selectionState.CurrentNode().Line())
		}
	})
	t.Run("select", func(t *testing.T) {
		nodes := testTree()
		var selectionState SelectionState
		selectionState.SetCurrentNode(nodes["b"])
		if got := send(nodes, &selectionState, key(tcell.KeyEnter, "")); got != (SelectedMsg{Node: nodes["b"]}) {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("click selects the node", func(t *testing.T) {
		nodes := testTree()
		var selectionState SelectionState
		click := tview.MouseMsg{EventMouse: tcell.NewEventMouse(1, 1, tcell.ButtonNone, tcell.ModNone), Action: tview.MouseLeftClick}
		if got := send(nodes, &selectionState, click); got != (SelectedMsg{Node: nodes["a1"]}) {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("collapsed children are skipped", func(t *testing.T) {
		nodes := testTree()
		var selectionState SelectionState
		nodes["a"].SetExpanded(false)
		selectionState.SetCurrentNode(nodes["a"])
		send(nodes, &selectionState, key(tcell.KeyDown, ""))
		if selectionState.CurrentNode() != nodes["b"] {
			t.Fatalf("current = %v", selectionState.CurrentNode().Line())
		}
	})
}

func TestNodePathTo(t *testing.T) {
	nodes := testTree()
	root := nodes["root"]
	t.Run("descendant", func(t *testing.T) {
		if got := root.PathTo(nodes["a1"]); !slices.Equal(got, []*Node{root, nodes["a"], nodes["a1"]}) {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("itself", func(t *testing.T) {
		if got := root.PathTo(root); !slices.Equal(got, []*Node{root}) {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("not under", func(t *testing.T) {
		if got := nodes["b"].PathTo(nodes["a1"]); got != nil {
			t.Fatalf("got %v", got)
		}
	})
}
