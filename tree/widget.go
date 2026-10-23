// Package tree shows a tree of nodes that can be expanded, collapsed, and selected.
package tree

import (
	"slices"

	"github.com/ayn2op/tview/layout"

	"github.com/ayn2op/tview"
	"github.com/gdamore/tcell/v3"
	"github.com/rockorager/go-uucode"
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

// Widget draws the nodes under a root and turns keys and the mouse into the node to make current.
type Widget struct {
	root    *Node
	current *Node

	width, height layout.Length
	topLevel      int
	spacing       func(node *Node) int

	markers       Markers
	graphics      bool
	graphicsSet   tview.BorderSet
	graphicsStyle tview.Style

	keybind func(tview.KeyMsg) Action
	focused bool

	onChange func(*Node) tview.Msg
	onSelect func(*Node) tview.Msg
}

var _ tview.Widget = Widget{}

// New returns a tree of the nodes under root, with current as its current node, or nil for none. The model owns the current node. It is interactive only once OnChange is set.
func New(root, current *Node) Widget {
	return Widget{
		root:        root,
		current:     current,
		width:       layout.Fill,
		height:      layout.Fill,
		markers:     Markers{Expanded: expandedMarker, Collapsed: collapsedMarker},
		graphics:    true,
		graphicsSet: tview.BorderSetPlain(),
		keybind:     DefaultKeybind,
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

// Spacing sets the function that returns the number of empty rows above each shown node after the first. The lines between nodes continue through them.
func (w Widget) Spacing(f func(node *Node) int) Widget {
	w.spacing = f
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

// OnChange makes the tree interactive, turning keys and the mouse into the message f returns for the node to make current.
func (w Widget) OnChange(f func(current *Node) tview.Msg) Widget {
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
		return layout.Size{Width: w.widest(rows), Height: height(rows)}
	})
}

// Target returns the row of the current node, or no rows for none.
func (w Widget) Target(int) (top, height int) {
	if v := w.resolve(); v.current >= 0 {
		return v.rows[v.current].y, 1
	}
	return 0, 0
}

// RowsWidth returns the width of the widest of the nodes within height rows from top.
func (w Widget) RowsWidth(top, height int) int {
	rows := w.rows()
	start := slices.IndexFunc(rows, func(r row) bool { return r.y >= top })
	if start < 0 {
		return 0
	}
	end := slices.IndexFunc(rows[start:], func(r row) bool { return r.y >= top+height })
	if end < 0 {
		end = len(rows) - start
	}
	return w.widest(rows[start : start+end])
}

// row is a node shown on one line: the row of its parent, where its lines and text start, and the line it is on, below the spacing above it.
type row struct {
	node   *Node
	parent int
	gx, tx int
	y      int
}

// height returns the number of lines rows take.
func height(rows []row) int {
	if len(rows) == 0 {
		return 0
	}
	return rows[len(rows)-1].y + 1
}

// rows returns the shown nodes, with the children of expanded nodes below them.
func (w Widget) rows() []row {
	if w.root == nil {
		return nil
	}

	rows := make([]row, 0, w.shown(w.root, 0))
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
			y := height(rows)
			if len(rows) > 0 && w.spacing != nil {
				y += w.spacing(node)
			}
			rows = append(rows, row{node: node, parent: parent, gx: gx, tx: tx, y: y})
			parent = len(rows) - 1
		}
		if node.expanded {
			for _, child := range node.children {
				walk(child, parent, level+1, tx)
			}
		}
	}
	walk(w.root, -1, 0, 0)
	return rows
}

// shown returns how many rows node at level and the children of it that are expanded take.
func (w Widget) shown(node *Node, level int) (count int) {
	if level >= w.topLevel {
		count++
	}
	if node.expanded {
		for _, child := range node.children {
			count += w.shown(child, level+1)
		}
	}
	return count
}

// view is the rows of the tree and the one of the current node, or -1.
type view struct {
	rows    []row
	current int
}

// resolve returns the rows and the row of the current node, falling back to the first selectable one.
func (w Widget) resolve() view {
	v := view{rows: w.rows(), current: -1}
	if c := w.current; c != nil {
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
		widest = max(widest, r.tx+uucode.StringWidth(w.marker(r.node))+r.node.spansWidth)
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
	top, end := max(-area.Y, 0), min(area.Height, bottom-area.Y)
	for index, current := range v.rows {
		if current.y >= end {
			break
		}
		node := current.node
		if w.graphics {
			// The lines into this row run through the spacing above it.
			for gap := max(height(v.rows[:index]), top); gap < current.y; gap++ {
				w.drawBranches(screen, v, current, x, area.Y+gap, width)
				if current.tx > current.gx && current.gx < width {
					screen.Put(x+current.gx, area.Y+gap, set.Right, w.graphicsStyle)
				}
			}
		}
		if current.y < top {
			continue
		}
		y := area.Y + current.y
		if w.graphics {
			w.drawBranches(screen, v, current, x, y, width)
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
		if len(node.spans) > 0 {
			markerStyle = node.spans[0].Style
		}
		textX := current.tx
		textX += tview.Print(screen, w.marker(node), x+textX, y, width-textX, tview.HorizontalAlignmentLeft, markerStyle)
		for _, span := range node.spans {
			if textX >= width {
				break
			}
			style := span.Style
			if index == v.current {
				style = tview.MergeStyle(style, node.selectedTextStyle)
			}
			textX += tview.Print(screen, span.Text, x+textX, y, width-textX, tview.HorizontalAlignmentLeft, style)
		}
	}
}

// drawBranches draws, on line y, the branches of the ancestors of current that are not last children, which continue past it.
func (w Widget) drawBranches(screen tview.Screen, v view, current row, x, y, width int) {
	for ancestor := current.parent; ancestor >= 0 && v.rows[ancestor].parent >= 0; ancestor = v.rows[ancestor].parent {
		a := v.rows[ancestor]
		parent := v.rows[a.parent].node
		if a.gx < width && parent.children[len(parent.children)-1] != a.node {
			screen.Put(x+a.gx, y, w.graphicsSet.Right, w.graphicsStyle)
		}
	}
}

// Handle turns keys (while focused) into the OnChange message, and ActionSelect and clicks on nodes into the OnSelect message. Other messages pass through unchanged.
func (w Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	if w.onChange == nil {
		return msg
	}
	switch m := msg.(type) {
	case tview.KeyMsg:
		if !w.focused {
			return msg
		}
		v := w.resolve()
		current := v.current
		switch w.keybind(m) {
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
		return w.onChange(v.node(current))
	case tview.MouseMsg:
		x, y := m.Position()
		if m.Action != tview.MouseLeftClick || !area.Contains(x, y) {
			return msg
		}
		v := w.resolve()
		index := slices.IndexFunc(v.rows, func(r row) bool { return r.y == y-area.Y })
		if node := v.node(index); node != nil && node.selectable {
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
