package tree

import (
	"slices"
	"testing"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/layout"
	"github.com/gdamore/tcell/v3"
)

// testTree returns the nodes of a root with children a and b, where a has child a1, all expanded.
func testTree() map[string]*Node {
	nodes := map[string]*Node{"root": NewNode("root"), "a": NewNode("a"), "a1": NewNode("a1"), "b": NewNode("b")}
	nodes["root"].AddChild(nodes["a"]).AddChild(nodes["b"])
	nodes["a"].AddChild(nodes["a1"])
	return nodes
}

// changed is the node a tree made current.
type changed struct{ node *Node }

func interactive(root, current *Node) Widget {
	return New(root, current).TopLevel(1).Markers(Markers{}).Focused(true).OnChange(func(n *Node) tview.Msg { return changed{n} }).OnSelect(func(n *Node) tview.Msg { return n })
}

func TestWidgetDraw(t *testing.T) {
	screen := screentest.New(t, 6, 3)
	var current *Node
	interactive(testTree()["root"], current).Draw(screen, tview.Rectangle{Width: 6, Height: 3})
	for y, want := range []string{"a     ", "└──a1 ", "b     "} {
		if got := screentest.Row(screen, y, 6); got != want {
			t.Fatalf("row %d = %q, want %q", y, got, want)
		}
	}
}

func TestWidgetLayout(t *testing.T) {
	nodes := testTree()
	nodes["a1"].SetLine(NewNode("long name").Line())
	limits := layout.Limits{Max: layout.Size{Width: 6, Height: 2}, Infinite: layout.Axes{Width: true, Height: true}}
	if got := interactive(nodes["root"], nil).Layout(limits); got != (layout.Size{Width: 12, Height: 3}) {
		t.Fatalf("size = %+v, want 12 by 3", got)
	}
}

func TestWidgetTarget(t *testing.T) {
	nodes := testTree()
	current := nodes["b"]
	if top, height := interactive(nodes["root"], current).Target(0); top != 2 || height != 1 {
		t.Fatalf("target = %d, %d", top, height)
	}
}

func TestWidgetHandle(t *testing.T) {
	area := tview.Rectangle{Width: 6, Height: 3}
	key := func(k tcell.Key, str string) tview.KeyMsg { return tcell.NewEventKey(k, str, tcell.ModNone) }
	// send passes msg through the tree and makes current the node it changes to, if any, returning the other result.
	send := func(nodes map[string]*Node, current **Node, msg tview.Msg) tview.Msg {
		out := interactive(nodes["root"], *current).Handle(msg, area)
		if c, ok := out.(changed); ok {
			*current = c.node
			return nil
		}
		return out
	}

	t.Run("down selects the next node", func(t *testing.T) {
		nodes := testTree()
		current := nodes["a"]
		send(nodes, &current, key(tcell.KeyDown, ""))
		if current != nodes["a1"] {
			t.Fatalf("current = %v", current.Line())
		}
	})
	t.Run("move to parent", func(t *testing.T) {
		nodes := testTree()
		current := nodes["a1"]
		send(nodes, &current, key(tcell.KeyRune, "K"))
		if current != nodes["a"] {
			t.Fatalf("current = %v", current.Line())
		}
	})
	t.Run("select", func(t *testing.T) {
		nodes := testTree()
		current := nodes["b"]
		if got := send(nodes, &current, key(tcell.KeyEnter, "")); got != nodes["b"] {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("click selects the node", func(t *testing.T) {
		nodes := testTree()
		var current *Node
		click := tview.MouseMsg{EventMouse: tcell.NewEventMouse(1, 1, tcell.ButtonNone, tcell.ModNone), Action: tview.MouseLeftClick}
		if got := send(nodes, &current, click); got != nodes["a1"] {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("unfocused passes keys through", func(t *testing.T) {
		nodes := testTree()
		var current *Node
		down := key(tcell.KeyDown, "")
		if got := interactive(nodes["root"], current).Focused(false).Handle(down, area); got != down {
			t.Fatalf("got %v, want %v", got, down)
		}
	})
	t.Run("collapsed children are skipped", func(t *testing.T) {
		nodes := testTree()
		var current *Node
		nodes["a"].SetExpanded(false)
		current = nodes["a"]
		send(nodes, &current, key(tcell.KeyDown, ""))
		if current != nodes["b"] {
			t.Fatalf("current = %v", current.Line())
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
