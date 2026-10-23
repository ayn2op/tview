package tree

import (
	"slices"
	"testing"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/layout"
	"github.com/ayn2op/tview/viewport"
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
	nodes["a1"].SetSpans(NewNode("long name").Spans()...)
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
			t.Fatalf("current = %v", current.Spans())
		}
	})
	t.Run("move to parent", func(t *testing.T) {
		nodes := testTree()
		current := nodes["a1"]
		send(nodes, &current, key(tcell.KeyRune, "K"))
		if current != nodes["a"] {
			t.Fatalf("current = %v", current.Spans())
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
			t.Fatalf("current = %v", current.Spans())
		}
	})
}

func TestWidgetSpacing(t *testing.T) {
	// nodes is a with children a1, which has child x, and a2, with a row above every node after the first.
	nodes := map[string]*Node{"root": NewNode("root"), "a": NewNode("a"), "a1": NewNode("a1"), "x": NewNode("x"), "a2": NewNode("a2")}
	nodes["root"].AddChild(nodes["a"])
	nodes["a"].AddChild(nodes["a1"]).AddChild(nodes["a2"])
	nodes["a1"].AddChild(nodes["x"])
	tree := func(current *Node) Widget {
		return interactive(nodes["root"], current).Spacing(func(*Node) int { return 1 })
	}

	t.Run("lines continue through the spacing", func(t *testing.T) {
		screen := screentest.New(t, 7, 7)
		tree(nil).Draw(screen, tview.Rectangle{Width: 7, Height: 7})
		for y, want := range []string{"a      ", "│      ", "├──a1  ", "│  │   ", "│  └──x", "│      ", "└──a2  "} {
			if got := screentest.Row(screen, y, 7); got != want {
				t.Fatalf("row %d = %q, want %q", y, got, want)
			}
		}
	})
	t.Run("drawn scrolled past the first rows", func(t *testing.T) {
		screen := screentest.New(t, 6, 2)
		tree(nil).Draw(screen, tview.Rectangle{Y: -3, Width: 6, Height: 7})
		for y, want := range []string{"│  │  ", "│  └──"} {
			if got := screentest.Row(screen, y, 6); got != want {
				t.Fatalf("row %d = %q, want %q", y, got, want)
			}
		}
	})
	t.Run("layout includes the spacing", func(t *testing.T) {
		limits := layout.Limits{Max: layout.Size{Width: 6, Height: 2}, Infinite: layout.Axes{Width: true, Height: true}}
		if got := tree(nil).Layout(limits).Height; got != 7 {
			t.Fatalf("height = %d, want 7", got)
		}
	})
	t.Run("target is the line of the current node", func(t *testing.T) {
		if top, height := tree(nodes["a2"]).Target(0); top != 6 || height != 1 {
			t.Fatalf("target = %d, %d", top, height)
		}
	})
	t.Run("rows width counts the nodes on the lines", func(t *testing.T) {
		if got := tree(nil).RowsWidth(3, 2); got != 7 {
			t.Fatalf("width = %d, want 7", got)
		}
		if got := tree(nil).RowsWidth(5, 1); got != 0 {
			t.Fatalf("width of spacing = %d, want 0", got)
		}
	})
	t.Run("click on a node selects it", func(t *testing.T) {
		click := tview.MouseMsg{EventMouse: tcell.NewEventMouse(1, 2, tcell.ButtonNone, tcell.ModNone), Action: tview.MouseLeftClick}
		if got := tree(nil).Handle(click, tview.Rectangle{Width: 6, Height: 7}); got != nodes["a1"] {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("click on spacing selects nothing", func(t *testing.T) {
		click := tview.MouseMsg{EventMouse: tcell.NewEventMouse(1, 1, tcell.ButtonNone, tcell.ModNone), Action: tview.MouseLeftClick}
		if got := tree(nil).Handle(click, tview.Rectangle{Width: 6, Height: 7}); got != nil {
			t.Fatalf("got %v", got)
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

// BenchmarkWidgetDraw draws a tree of 50 expanded nodes with 30 children each in a viewport, which asks for its target and width too.
func BenchmarkWidgetDraw(b *testing.B) {
	root := NewNode("")
	for range 50 {
		parent := NewNode("parent")
		for range 30 {
			parent.AddChild(NewNode("child"))
		}
		root.AddChild(parent)
	}

	tree := New(root, root.children[10].children[5]).TopLevel(1).Graphics(true)
	widget := viewport.New(tree, viewport.ScrollState{}).Target(tree.Target).ContentWidth(tree.RowsWidth)
	screen := screentest.New(b, 40, 50)
	b.ReportAllocs()
	for b.Loop() {
		widget.Draw(screen, tview.Rectangle{Width: 40, Height: 50})
	}
}
