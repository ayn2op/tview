// Package tree shows a tree of nodes that can be expanded, collapsed, and selected.
package tree

import (
	"slices"

	"github.com/ayn2op/tview/layout"

	"github.com/ayn2op/tview"
	"github.com/gdamore/tcell/v3"
	"github.com/rivo/uniseg"
)

const (
	expandedMarker  = "▾ "
	collapsedMarker = "▸ "
)

// Markers are drawn before a node's text depending on whether it is expanded.
type Markers struct {
	Expanded  string
	Collapsed string
	Leaf      string
}

// Widget draws the nodes under a root and turns keys and the mouse into Changes.
type Widget struct {
	root           *Node
	selectionState SelectionState
	width, height  layout.Length
	topLevel       int
	markers        Markers
	graphics       bool
	graphicsSet    tview.BorderSet
	graphicsStyle  tview.Style
	keybind        func(tview.KeyMsg) Action
	focused        bool
	onChange       func(Change) tview.Msg
	onSelect       func(*Node) tview.Msg
}

var _ tview.Widget = Widget{}

// New returns a tree of the nodes under root, with selectionState as its current node. It is interactive only once OnChange is set.
func New(root *Node, selectionState SelectionState) Widget {
	return Widget{
		root:           root,
		selectionState: selectionState,
		width:          layout.Fill,
		height:         layout.Fill,
		markers:        Markers{Expanded: expandedMarker, Collapsed: collapsedMarker},
		graphics:       true,
		graphicsSet:    tview.BorderSetPlain(),
		keybind:        DefaultKeybind,
	}
}

// Width sets the width of the tree.
func (w Widget) Width(width layout.Length) Widget {
	w.width = width
	return w
}

// Height sets the height of the tree.
func (w Widget) Height(height layout.Length) Widget {
	w.height = height
	return w
}

// TopLevel sets the first level shown, where the root is level 0, its children level 1, and so on.
func (w Widget) TopLevel(level int) Widget {
	w.topLevel = level
	return w
}

// Markers sets the markers drawn before nodes.
func (w Widget) Markers(markers Markers) Widget {
	w.markers = markers
	return w
}

// Graphics sets whether lines are drawn between nodes.
func (w Widget) Graphics(graphics bool) Widget {
	w.graphics = graphics
	return w
}

// GraphicsSet sets the characters the lines are drawn with.
func (w Widget) GraphicsSet(set tview.BorderSet) Widget {
	w.graphicsSet = set
	return w
}

// GraphicsStyle sets the style of the lines.
func (w Widget) GraphicsStyle(style tview.Style) Widget {
	w.graphicsStyle = style
	return w
}

// Keybind sets the function that turns keys into Actions, or ActionNone for keys it does not bind, DefaultKeybind unless set.
func (w Widget) Keybind(f func(tview.KeyMsg) Action) Widget {
	w.keybind = f
	return w
}

// Focused sets whether the tree receives keys.
func (w Widget) Focused(focused bool) Widget {
	w.focused = focused
	return w
}

// OnChange makes the tree interactive, turning keys and the mouse into the message f returns for the Change, which the model applies with SelectionState.Apply.
func (w Widget) OnChange(f func(Change) tview.Msg) Widget {
	w.onChange = f
	return w
}

// OnSelect sets the function that turns a node selected with ActionSelect or a click into a message.
func (w Widget) OnSelect(f func(*Node) tview.Msg) Widget {
	w.onSelect = f
	return w
}

// Size returns the width and height of the tree.
func (w Widget) Size() (width, height layout.Length) {
	return w.width, w.height
}

// Layout returns the size of the tree within limits.
func (w Widget) Layout(limits layout.Limits) layout.Size {
	return layout.Sized(limits, w.width, w.height, func(layout.Limits) layout.Size {
		rows := w.rows()
		return layout.Size{Width: w.widest(rows), Height: len(rows)}
	})
}

// Target returns the row of the current node, or no rows for none.
func (w Widget) Target(int) (top, height int) {
	if current := w.resolve().current; current >= 0 {
		return current, 1
	}
	return 0, 0
}

// RowsWidth returns the width of the widest of height rows from top.
func (w Widget) RowsWidth(top, height int) int {
	rows := w.rows()
	top = min(max(top, 0), len(rows))
	return w.widest(rows[top:min(top+max(height, 0), len(rows))])
}

// row is a node shown on one line: the row of its parent, and where its lines and text start.
type row struct {
	node   *Node
	parent int
	gx, tx int
}

// rows returns the shown nodes, with the children of expanded nodes below them.
func (w Widget) rows() (rows []row) {
	var walk func(node *Node, parent, level, parentX int)
	walk = func(node *Node, parent, level, parentX int) {
		gx, tx := parentX, parentX+node.indent
		if w.graphics {
			tx++
		}
		if level == w.topLevel || level == 0 {
			gx, tx = 0, 0
		}
		if level >= w.topLevel {
			rows = append(rows, row{node: node, parent: parent, gx: gx, tx: tx})
			parent = len(rows) - 1
		}
		if node.expanded {
			for _, child := range node.children {
				walk(child, parent, level+1, tx)
			}
		}
	}
	if w.root != nil {
		walk(w.root, -1, 0, 0)
	}
	return rows
}

// view is the rows of the tree and the one of the current node, or -1.
type view struct {
	rows    []row
	current int
}

// resolve returns the rows and the row of the current node, falling back to the first selectable one.
func (w Widget) resolve() view {
	v := view{rows: w.rows(), current: -1}
	if c := w.selectionState.current; c != nil {
		v.current = slices.IndexFunc(v.rows, func(r row) bool { return r.node == c && c.selectable })
		if v.current < 0 {
			v.current = slices.IndexFunc(v.rows, func(r row) bool { return r.node.selectable })
		}
	}
	return v
}

// widest returns the width of the widest of rows.
func (w Widget) widest(rows []row) (widest int) {
	for _, r := range rows {
		widest = max(widest, r.tx+uniseg.StringWidth(w.marker(r.node))+r.node.lineWidth)
	}
	return widest
}

// marker returns the marker drawn before node's text.
func (w Widget) marker(node *Node) string {
	if node.expandable || len(node.children) > 0 {
		if node.expanded {
			return w.markers.Expanded
		}
		return w.markers.Collapsed
	}
	return w.markers.Leaf
}

// node returns the node of row index, or nil if there is none.
func (v view) node(index int) *Node {
	if index < 0 || index >= len(v.rows) {
		return nil
	}
	return v.rows[index].node
}

// Draw draws the rows with their lines, markers, and text, the current node in its selected style.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	v := w.resolve()
	x, width, set := area.X, area.Width, w.graphicsSet
	// Rows off the screen, as in a viewport, are skipped.
	_, bottom := screen.Size()
	for index := max(-area.Y, 0); index < len(v.rows) && index < min(area.Height, bottom-area.Y); index++ {
		current, y := v.rows[index], area.Y+index
		node := current.node
		if w.graphics {
			// Branches of ancestors that are not last children continue past this row.
			for ancestor := current.parent; ancestor >= 0 && v.rows[ancestor].parent >= 0; ancestor = v.rows[ancestor].parent {
				a := v.rows[ancestor]
				parent := v.rows[a.parent].node
				if a.gx < width && parent.children[len(parent.children)-1] != a.node {
					screen.Put(x+a.gx, y, set.Right, w.graphicsStyle)
				}
			}
			if current.tx > current.gx && current.gx < width {
				connector := set.BottomLeft
				if current.parent >= 0 {
					if siblings := v.rows[current.parent].node.children; siblings[len(siblings)-1] != node {
						connector = set.LeftT
					}
				}
				screen.Put(x+current.gx, y, connector, w.graphicsStyle)
				for pos := current.gx + 1; pos < current.tx && pos < width; pos++ {
					screen.Put(x+pos, y, set.Top, w.graphicsStyle)
				}
			}
		}
		if current.tx >= width {
			continue
		}

		markerStyle := tcell.StyleDefault
		if len(node.line) > 0 {
			markerStyle = node.line[0].Style
		}
		textX := current.tx
		textX += tview.Print(screen, w.marker(node), x+textX, y, width-textX, tview.AlignmentLeft, markerStyle)
		for _, segment := range node.line {
			if textX >= width {
				break
			}
			style := segment.Style
			if index == v.current {
				style = tview.MergeStyle(style, node.selectedTextStyle)
			}
			textX += tview.Print(screen, segment.Text, x+textX, y, width-textX, tview.AlignmentLeft, style)
		}
	}
}

// Handle turns keys and ActionMsgs (while focused) into a Change once OnChange is set, and ActionSelect and clicks on nodes into the OnSelect message. Other messages pass through unchanged.
func (w Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	if w.onChange == nil {
		return msg
	}
	if key, ok := msg.(tview.KeyMsg); ok && w.focused {
		if action := w.keybind(key); action != ActionNone {
			msg = ActionMsg(action)
		}
	}
	switch m := msg.(type) {
	case ActionMsg:
		if !w.focused {
			return msg
		}
		v := w.resolve()
		current := v.current
		switch Action(m) {
		case ActionDown:
			current = v.step(current, 1)
		case ActionUp:
			current = v.step(current, -1)
		case ActionTop:
			current = v.step(-1, 1)
		case ActionBottom:
			current = v.step(len(v.rows), -1)
		case ActionMoveToParent:
			if current >= 0 {
				if parent := v.rows[current].parent; parent >= 0 && v.rows[parent].node.selectable {
					current = parent
				}
			}
		case ActionSelect:
			return w.selectNode(v.node(current))
		default:
			return msg
		}
		return w.onChange(Change{current: v.node(current)})
	case tview.MouseMsg:
		x, y := m.Position()
		if m.Action != tview.MouseLeftClick || !area.Contains(x, y) {
			return msg
		}
		if node := w.resolve().node(y - area.Y); node != nil && node.selectable {
			return w.selectNode(node)
		}
		return nil
	}
	return msg
}

// step returns the next selectable row after index in direction, or index if there is none.
func (v view) step(index, direction int) int {
	for next := index + direction; next >= 0 && next < len(v.rows); next += direction {
		if v.rows[next].node.selectable {
			return next
		}
	}
	return index
}

// selectNode returns the OnSelect message for node, or nil if node is nil or OnSelect is not set.
func (w Widget) selectNode(node *Node) tview.Msg {
	if node == nil || w.onSelect == nil {
		return nil
	}
	return w.onSelect(node)
}
