// Package tree shows a tree of nodes that can be expanded, collapsed, and selected.
package tree

import (
	"slices"

	"github.com/ayn2op/tview"
	"github.com/gdamore/tcell/v3"
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
	width, height  tview.Length
	topLevel       int
	markers        Markers
	graphics       bool
	graphicsSet    tview.BorderSet
	graphicsStyle  tcell.Style
	keybind        func(tview.KeyMsg) (Action, bool)
	focused        bool
	onChange       func(Change) tview.Msg
	onSelect       func(*Node) tview.Msg
}

var _ tview.Element = Widget{}

// New returns a tree of the nodes under root, with selectionState as its current node and scroll position, that draws lines between nodes and fills its parent. It is interactive only once OnChange is set.
func New(root *Node, selectionState SelectionState) Widget {
	return Widget{
		root:           root,
		selectionState: selectionState,
		width:          tview.Fill,
		height:         tview.Fill,
		markers:        Markers{Expanded: "▾ ", Collapsed: "▸ "},
		graphics:       true,
		graphicsSet:    tview.BorderSetPlain(),
		keybind:        DefaultKeybind,
	}
}

// Width sets the width of the tree.
func (w Widget) Width(width tview.Length) Widget {
	w.width = width
	return w
}

// Height sets the height of the tree.
func (w Widget) Height(height tview.Length) Widget {
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
func (w Widget) GraphicsStyle(style tcell.Style) Widget {
	w.graphicsStyle = style
	return w
}

// Keybind sets the function that turns keys into Actions, DefaultKeybind unless set.
func (w Widget) Keybind(f func(tview.KeyMsg) (Action, bool)) Widget {
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
func (w Widget) Size() (width, height tview.Length) {
	return w.width, w.height
}

// row is a node shown on one line: the row of its parent, its level, and where its lines and text start.
type row struct {
	node          *Node
	parent        int
	level, gx, tx int
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
			rows = append(rows, row{node: node, parent: parent, level: level, gx: gx, tx: tx})
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

// view is the tree laid out in an area.
type view struct {
	rows    []row
	current int
	offset  int
	height  int
}

// resolve lays the tree out for height rows: the row of the current node, falling back to the first selectable one, and the scroll position.
func (w Widget) resolve(height int) view {
	v := view{rows: w.rows(), current: -1, height: height}
	if c := w.selectionState.current; c != nil {
		v.current = slices.IndexFunc(v.rows, func(r row) bool { return r.node == c && c.selectable })
		if v.current < 0 {
			v.current = slices.IndexFunc(v.rows, func(r row) bool { return r.node.selectable })
		}
	}
	v.offset = w.selectionState.offset
	if w.selectionState.center && v.current >= 0 {
		v.offset = v.current - height/2
	}
	v.offset = v.clamp(v.offset)
	return v
}

func (v view) clamp(offset int) int {
	return min(max(offset, 0), max(len(v.rows)-v.height, 0))
}

// node returns the node of row index, or nil if there is none.
func (v view) node(index int) *Node {
	if index < 0 || index >= len(v.rows) {
		return nil
	}
	return v.rows[index].node
}

// Draw draws the visible rows with their lines, markers, and text, the current node in its selected style.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	v := w.resolve(area.Height)
	x, width, set := area.X, area.Width, w.graphicsSet
	for index := v.offset; index < len(v.rows) && index-v.offset < area.Height; index++ {
		current, y := v.rows[index], area.Y+index-v.offset
		node := current.node
		if w.graphics {
			// Branches of ancestors that are not last children continue past this row.
			for ancestor := current.parent; ancestor >= 0 && v.rows[ancestor].parent >= 0; ancestor = v.rows[ancestor].parent {
				a := v.rows[ancestor]
				parent := v.rows[a.parent].node
				if a.gx < width && parent.children[len(parent.children)-1] != a.node {
					if y-1 >= area.Y && a.tx > a.gx {
						tview.PrintJoinedSemigraphics(screen, x+a.gx, y-1, set.Left, w.graphicsStyle)
					}
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
				tview.PrintJoinedSemigraphics(screen, x+current.gx, y, connector, w.graphicsStyle)
				for pos := current.gx + 1; pos < current.tx && pos < width; pos++ {
					screen.Put(x+pos, y, set.Top, w.graphicsStyle)
				}
			}
		}
		if current.tx >= width {
			continue
		}

		marker := w.markers.Leaf
		if node.expandable || len(node.children) > 0 {
			marker = w.markers.Collapsed
			if node.expanded {
				marker = w.markers.Expanded
			}
		}
		markerStyle := tcell.StyleDefault
		if len(node.line) > 0 {
			markerStyle = node.line[0].Style
		}
		textX := current.tx
		if marker != "" {
			markerWidth := tview.Print(screen, marker, x+textX, y, width-textX, tview.AlignmentLeft, markerStyle)
			textX += markerWidth
		}
		for _, segment := range node.line {
			if textX >= width {
				break
			}
			style := segment.Style
			if index == v.current {
				style = tview.MergeStyle(style, node.selectedTextStyle)
			}
			segmentWidth := tview.Print(screen, segment.Text, x+textX, y, width-textX, tview.AlignmentLeft, style)
			textX += segmentWidth
		}
	}
}

// Handle turns keys and ActionMsgs (while focused) and the mouse within area into a Change once OnChange is set, and ActionSelect and clicks on nodes into the OnSelect message. Other messages pass through unchanged.
func (w Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	if w.onChange == nil {
		return msg
	}
	if key, ok := msg.(tview.KeyMsg); ok && w.focused {
		if action, ok := w.keybind(key); ok {
			msg = ActionMsg(action)
		}
	}
	// Laying the tree out is the costly part, so messages the tree ignores return before it.
	switch msg.(type) {
	case ActionMsg:
		if !w.focused {
			return msg
		}
	case tview.MouseMsg:
	default:
		return msg
	}
	v := w.resolve(area.Height)
	a := Change{current: v.node(v.current), offset: v.offset, dragging: w.selectionState.dragging, dragY: w.selectionState.dragY}
	center := false
	switch m := msg.(type) {
	case ActionMsg:
		switch Action(m) {
		case ActionDown:
			a.current, center = v.node(v.step(v.current, 1)), true
		case ActionUp:
			a.current, center = v.node(v.step(v.current, -1)), true
		case ActionTop:
			a.current, center = v.node(v.step(-1, 1)), true
		case ActionBottom:
			a.current, center = v.node(v.step(len(v.rows), -1)), true
		case ActionMoveToParent:
			if v.current >= 0 {
				if parent := v.rows[v.current].parent; parent >= 0 && v.rows[parent].node.selectable {
					a.current, center = v.rows[parent].node, true
				}
			}
		case ActionSelect:
			return w.selectNode(a.current)
		default:
			return msg
		}
	case tview.MouseMsg:
		x, y := m.Position()
		if a.dragging && m.Action == tview.MouseMove && m.Buttons()&tcell.Button1 != 0 {
			a.offset, a.dragY = a.offset+a.dragY-y, y
			break
		}
		if !area.Contains(x, y) {
			return msg
		}
		switch m.Action {
		case tview.MouseLeftDown:
			a.dragging, a.dragY = true, y
		case tview.MouseLeftUp:
			a.dragging = false
		case tview.MouseLeftClick:
			if node := v.node(v.offset + y - area.Y); node != nil && node.selectable {
				return w.selectNode(node)
			}
			return nil
		case tview.MouseScrollUp:
			a.offset--
		case tview.MouseScrollDown:
			a.offset++
		default:
			return msg
		}
	default:
		return msg
	}
	if index := slices.IndexFunc(v.rows, func(r row) bool { return r.node == a.current }); center && index >= 0 {
		a.offset = index - v.height/2
	}
	a.offset = v.clamp(a.offset)
	return w.onChange(a)
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
