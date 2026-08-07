package tree

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/keybind"
	"github.com/gdamore/tcell/v3"
)

// Markers are glyphs drawn before node text.
type Markers struct {
	Expanded  string
	Collapsed string
	Leaf      string
}

// Model displays tree structures. A tree consists of nodes (Node
// objects) where each node has zero or more child nodes and exactly one parent
// node (except for the root node which has no parent node).
//
// The SetRoot() function is used to specify the root of the tree. Other nodes
// are added locally to the root node or any of its descendents. See the
// Node documentation for details on node attributes. (You can use
// SetReference() to store a reference to nodes of your own tree structure.)
//
// Nodes can be selected by calling SetCurrentNode(). The user can navigate the
// cursor or the tree using the configured [Keybinds]. Selected nodes emit
// [SelectedMsg].
//
// The root node corresponds to level 0, its children correspond to level 1,
// their children to level 2, and so on. Per default, the first level that is
// displayed is 0, i.e. the root node. You can call SetTopLevel() to hide
// levels.
//
// If graphics are turned on (see SetGraphics()), lines indicate the tree's
// hierarchy. Alternative (or additionally), you can set different prefixes
// using SetPrefixes() for different levels, for example to display hierarchical
// bullet point lists.
type Model struct {
	*tview.Box

	// The root node.
	root *Node

	// The currently selected node or nil if no node is selected.
	currentNode *Node

	// The top hierarchical level shown. (0 corresponds to the root level.)
	topLevel int

	// Strings drawn before the nodes, based on their level.
	prefixes []string

	// Markers drawn before the node text depending on expansion state.
	markers Markers

	// Vertical scroll offset.
	offsetY int

	// If set to true, cursor tries to stay centered in the viewport.
	centerCursor bool

	// If set to true, all node texts will be aligned horizontally.
	align bool

	// If set to true, the tree structure is drawn using lines.
	graphics bool

	// The color of the lines.
	graphicsColor tcell.Color

	// Internal mouse track data.
	lastMouseY int

	keybinds Keybinds
}

type row struct {
	node          *Node
	parent        int
	level, gx, tx int
}

// NewModel returns a new tree view.
func NewModel() *Model {
	return &Model{
		Box:           tview.NewBox(),
		centerCursor:  true,
		graphics:      true,
		graphicsColor: tview.Styles.GraphicsColor,
		markers: Markers{
			Expanded:  "▾ ",
			Collapsed: "▸ ",
			Leaf:      "",
		},
		lastMouseY: -1,
		keybinds:   DefaultKeybinds(),
	}
}

// Root returns the root node of the tree. If no such node was previously
// set, nil is returned.
func (t *Model) Root() *Node {
	return t.root
}

// SetRoot sets the root node of the tree.
func (t *Model) SetRoot(root *Node) *Model {
	t.root = root
	return t
}

// CurrentNode returns the currently selected node or nil of no node is
// currently selected.
func (t *Model) CurrentNode() *Node {
	return t.currentNode
}

// SetCurrentNode sets the currently selected node. Provide nil to clear all
// cursors. Invalid nodes select the first visible, selectable node.
func (t *Model) SetCurrentNode(node *Node) *Model {
	t.currentNode = node
	rows := t.flatten()
	t.selectRow(rows, t.normalize(rows))
	return t
}

// GetPath returns all nodes located on the path from the root to the given
// node, including the root and the node itself. If there is no root node, nil
// is returned. If there are multiple paths to the node, a random one is chosen
// and returned.
func (t *Model) GetPath(node *Node) []*Node {
	if t.root == nil {
		return nil
	}

	var f func(current *Node, path []*Node) []*Node
	f = func(current *Node, path []*Node) []*Node {
		if current == node {
			return path
		}

		for _, child := range current.children {
			newPath := make([]*Node, len(path), len(path)+1)
			copy(newPath, path)
			if p := f(child, append(newPath, child)); p != nil {
				return p
			}
		}

		return nil
	}

	return f(t.root, []*Node{t.root})
}

// SetTopLevel sets the first tree level that is visible with 0 referring to the
// root, 1 to the root's child nodes, and so on. Nodes above the top level are
// not displayed.
func (t *Model) SetTopLevel(topLevel int) *Model {
	t.topLevel = topLevel
	return t
}

// SetCenterCursor controls whether the cursor is kept centered whenever
// possible.
func (t *Model) SetCenterCursor(center bool) *Model {
	t.centerCursor = center
	return t
}

// SetPrefixes defines the strings drawn before the nodes' texts. This is a
// slice of strings where each element corresponds to a node's hierarchy level,
// i.e. 0 for the root, 1 for the root's children, and so on (levels will
// cycle).
//
// For example, to display a hierarchical list with bullet points:
//
//	treeView.SetGraphics(false).
//	  SetPrefixes([]string{"* ", "- ", "x "})
//
// Deeper levels will cycle through the prefixes.
func (t *Model) SetPrefixes(prefixes []string) *Model {
	t.prefixes = prefixes
	return t
}

// Markers returns the marker strings currently used by this tree view.
func (t *Model) Markers() Markers {
	return t.markers
}

// SetMarkers sets the strings drawn before node text depending on node state.
// Expanded is used for nodes with children whose children are visible,
// Collapsed is used for nodes with children whose children are hidden, and
// Leaf is used for nodes without children.
func (t *Model) SetMarkers(markers Markers) *Model {
	t.markers = markers
	return t
}

// SetAlign controls the horizontal alignment of the node texts. If set to true,
// all texts except that of top-level nodes will be placed in the same column.
// If set to false, they will indent with the hierarchy.
func (t *Model) SetAlign(align bool) *Model {
	t.align = align
	return t
}

// SetGraphics sets a flag which determines whether or not line graphics are
// drawn to illustrate the tree's hierarchy.
func (t *Model) SetGraphics(showGraphics bool) *Model {
	t.graphics = showGraphics
	return t
}

// SetGraphicsColor sets the colors of the lines used to draw the tree structure.
func (t *Model) SetGraphicsColor(color tcell.Color) *Model {
	t.graphicsColor = color
	return t
}

// GetScrollOffset returns the number of node rows skipped at the top.
func (t *Model) GetScrollOffset() int {
	return t.offsetY
}

// GetRowCount returns the number of visible nodes, including rows off-screen.
func (t *Model) GetRowCount() int {
	return len(t.flatten())
}

// Move moves the cursor (if a node is currently selected) or scrolls the tree
// view (if there is no cursor), by the given offset (positive values to
// move/scroll down, negative values to move/scroll up). For cursor changes,
// the offset refers to the number selectable, visible nodes. For scrolling, the
// offset refers to the number of visible nodes.
//
// If the offset is 0, nothing happens.
func (t *Model) Move(offset int) *Model {
	if offset == 0 {
		return t
	}
	t.move(t.flatten(), offset)
	return t
}

func (t *Model) flatten() (rows []row) {
	var walk func(*Node, int, int, int)
	walk = func(node *Node, parent, level, parentX int) {
		gx, tx := parentX, parentX+node.indent
		if t.graphics {
			tx++
		}
		if level == t.topLevel || level == 0 || t.align && !t.graphics {
			gx, tx = 0, 0
		}
		if level >= t.topLevel {
			rows = append(rows, row{node: node, parent: parent, level: level, gx: gx, tx: tx})
			parent = len(rows) - 1
		}
		if node.expanded {
			for _, child := range node.children {
				walk(child, parent, level+1, tx)
			}
		}
	}
	if t.root != nil {
		walk(t.root, -1, 0, 0)
	}
	if t.align {
		maxX := 0
		for _, row := range rows {
			maxX = max(maxX, row.tx)
		}
		for i := range rows {
			if rows[i].level > t.topLevel {
				rows[i].tx = maxX
			}
		}
	}
	return
}

func (t *Model) index(rows []row) int {
	for i, row := range rows {
		if row.node == t.currentNode && row.node.selectable {
			return i
		}
	}
	return -1
}

func (t *Model) normalize(rows []row) int {
	if t.currentNode == nil {
		return -1
	}
	if i := t.index(rows); i >= 0 {
		return i
	}
	for i, row := range rows {
		if row.node.selectable {
			t.currentNode = row.node
			return i
		}
	}
	t.currentNode = nil
	return -1
}

func (t *Model) selectRow(rows []row, index int) {
	if index < 0 || index >= len(rows) {
		return
	}
	t.currentNode = rows[index].node
	_, _, _, height := t.InnerRect()
	if height <= 0 {
		return
	}
	if t.centerCursor {
		t.offsetY = min(max(index-height/2, 0), max(len(rows)-height, 0))
	} else if index < t.offsetY {
		t.offsetY = index
	} else if index >= t.offsetY+height {
		t.offsetY = index - height + 1
	}
}

func (t *Model) move(rows []row, delta int) {
	index := t.normalize(rows)
	if index < 0 {
		t.scroll(rows, delta)
		return
	}
	for delta != 0 {
		next, step := index, 1
		if delta < 0 {
			step = -1
		}
		for next += step; next >= 0 && next < len(rows) && !rows[next].node.selectable; next += step {
		}
		if next < 0 || next >= len(rows) {
			break
		}
		index, delta = next, delta-step
	}
	t.selectRow(rows, index)
}

func (t *Model) scroll(rows []row, delta int) {
	_, _, _, height := t.InnerRect()
	t.offsetY = min(max(t.offsetY+delta, 0), max(len(rows)-height, 0))
}

// View draws this model onto the screen.
func (t *Model) View(screen tcell.Screen) {
	t.Box.View(screen)
	rows := t.flatten()
	if len(rows) == 0 {
		return
	}
	_, totalHeight := screen.Size()
	x, y, width, height := t.InnerRect()
	offset := min(max(t.offsetY, 0), max(len(rows)-height, 0))

	// Draw the tree.
	posY := y
	borderSet := t.BorderSet()
	lineStyle := tcell.StyleDefault.Background(t.BackgroundColor()).Foreground(t.graphicsColor)
	for index, current := range rows {
		node := current.node
		// Skip invisible parts.
		if posY >= y+height+1 || posY >= totalHeight {
			break
		}
		if index < offset {
			continue
		}

		// Draw the graphics.
		if t.graphics {
			// Draw ancestor branches.
			for ancestor := current.parent; ancestor >= 0 && rows[ancestor].parent >= 0; ancestor = rows[ancestor].parent {
				a := rows[ancestor]
				if a.gx < width {
					// Draw a branch if this ancestor is not a last child.
					parent := rows[a.parent].node
					if parent.children[len(parent.children)-1] != a.node {
						if posY-1 >= y && a.tx > a.gx {
							tview.PrintJoinedSemigraphics(screen, x+a.gx, posY-1, borderSet.Left, lineStyle)
						}
						if posY < y+height {
							screen.Put(x+a.gx, posY, borderSet.Right, lineStyle)
						}
					}
				}
			}

			if current.tx > current.gx && current.gx < width {
				// BottomLeft for last child; LeftT for non-last siblings.
				connector := borderSet.BottomLeft
				if current.parent >= 0 {
					if siblings := rows[current.parent].node.children; len(siblings) > 0 && siblings[len(siblings)-1] != node {
						connector = borderSet.LeftT
					}
				}

				// Join this node.
				if posY < y+height {
					tview.PrintJoinedSemigraphics(screen, x+current.gx, posY, connector, lineStyle)

					for pos := current.gx + 1; pos < current.tx && pos < width; pos++ {
						screen.Put(x+pos, posY, borderSet.Top, lineStyle)
					}
				}
			}
		}

		// Draw the prefix and the text.
		if current.tx < width && posY < y+height {
			marker := t.markers.Leaf
			if node.expandable || len(node.children) > 0 {
				if node.expanded {
					marker = t.markers.Expanded
				} else {
					marker = t.markers.Collapsed
				}
			}

			// Prefix.
			var prefixWidth int
			prefixStyle := tcell.StyleDefault
			if len(node.line) > 0 {
				prefixStyle = node.line[0].Style
			}
			if len(t.prefixes) > 0 {
				_, _, prefixWidth = tview.PrintStyled(screen, t.prefixes[(current.level-t.topLevel)%len(t.prefixes)], x+current.tx, posY, 0, width-current.tx, tview.AlignmentLeft, prefixStyle, true)
			}

			// Marker.
			markerWidth := 0
			if marker != "" && current.tx+prefixWidth < width {
				_, _, markerWidth = tview.PrintStyled(screen, marker, x+current.tx+prefixWidth, posY, 0, width-current.tx-prefixWidth, tview.AlignmentLeft, prefixStyle, true)
			}

			// Text.
			if current.tx+prefixWidth+markerWidth < width {
				if node == t.currentNode {
					posX := 0
					for _, segment := range node.line {
						if posX >= width-current.tx-prefixWidth-markerWidth {
							break
						}
						style := tview.MergeStyle(segment.Style, node.selectedTextStyle)
						_, _, segmentWidth := tview.PrintStyled(
							screen,
							segment.Text,
							x+current.tx+prefixWidth+markerWidth+posX,
							posY,
							0,
							width-current.tx-prefixWidth-markerWidth-posX,
							tview.AlignmentLeft,
							style,
							false,
						)
						posX += segmentWidth
					}
				} else {
					posX := 0
					for _, segment := range node.line {
						if posX >= width-current.tx-prefixWidth-markerWidth {
							break
						}
						_, _, segmentWidth := tview.PrintStyled(
							screen,
							segment.Text,
							x+current.tx+prefixWidth+markerWidth+posX,
							posY,
							0,
							width-current.tx-prefixWidth-markerWidth-posX,
							tview.AlignmentLeft,
							segment.Style,
							false,
						)
						posX += segmentWidth
					}
				}
			}
		}

		// Advance.
		posY++
	}
}

func (t *Model) selectCurrentNode() tview.Cmd {
	node := t.currentNode
	if node == nil {
		return nil
	}
	return func() tview.Msg { return SelectedMsg{Node: node} }
}

func (t *Model) handleKeyMsg(msg tview.KeyMsg) tview.Cmd {
	rows := t.flatten()
	index := t.normalize(rows)
	switch {
	case keybind.Matches(msg, t.keybinds.Down):
		t.move(rows, 1)
	case keybind.Matches(msg, t.keybinds.Up):
		t.move(rows, -1)
	case keybind.Matches(msg, t.keybinds.Top):
		if index < 0 {
			t.offsetY = 0
		} else {
			for i, row := range rows {
				if row.node.selectable {
					t.selectRow(rows, i)
					break
				}
			}
		}
	case keybind.Matches(msg, t.keybinds.Bottom):
		if index < 0 {
			t.scroll(rows, len(rows))
		} else {
			for i := len(rows) - 1; i >= 0; i-- {
				if rows[i].node.selectable {
					t.selectRow(rows, i)
					break
				}
			}
		}
	case keybind.Matches(msg, t.keybinds.MoveToLastChild):
		child := index
		for i := index + 1; i < len(rows); i++ {
			if rows[i].parent == child && rows[i].node.selectable {
				child = i
			}
		}
		t.selectRow(rows, child)
	case keybind.Matches(msg, t.keybinds.MoveToParent):
		if index >= 0 && rows[index].parent >= 0 && rows[rows[index].parent].node.selectable {
			t.selectRow(rows, rows[index].parent)
		}
	case keybind.Matches(msg, t.keybinds.PageDown):
		_, _, _, height := t.InnerRect()
		t.move(rows, height)
	case keybind.Matches(msg, t.keybinds.PageUp):
		_, _, _, height := t.InnerRect()
		t.move(rows, -height)
	case keybind.Matches(msg, t.keybinds.Select):
		return t.selectCurrentNode()
	}
	return nil
}

func (t *Model) handleMouseMsg(msg tview.MouseMsg) tview.Cmd {
	x, y := msg.Position()
	if !t.InRect(x, y) {
		return nil
	}

	switch msg.Action {
	case tview.MouseLeftDown:
		t.lastMouseY = y
	case tview.MouseMove:
		if msg.Buttons()&tcell.Button1 != 0 && t.lastMouseY != -1 {
			t.scroll(t.flatten(), t.lastMouseY-y)
			t.lastMouseY = y
		}
	case tview.MouseLeftUp:
		t.lastMouseY = -1
	case tview.MouseLeftClick:
		rows := t.flatten()
		_, top, _, height := t.InnerRect()
		index := min(max(t.offsetY, 0), max(len(rows)-height, 0)) + y - top
		if index >= 0 && index < len(rows) {
			node := rows[index].node
			if node.selectable {
				t.selectRow(rows, index)
				return tview.Sequence(tview.SetFocus(t), func() tview.Msg {
					return SelectedMsg{Node: node}
				})
			}
		}
		return tview.SetFocus(t)
	case tview.MouseScrollUp:
		t.scroll(t.flatten(), -1)
	case tview.MouseScrollDown:
		t.scroll(t.flatten(), 1)
	}
	return nil
}

// Update handles input events for this model.
func (t *Model) Update(msg tview.Msg) tview.Cmd {
	switch msg := msg.(type) {
	case tview.KeyMsg:
		return t.handleKeyMsg(msg)
	case tview.MouseMsg:
		return t.handleMouseMsg(msg)
	}
	return nil
}
