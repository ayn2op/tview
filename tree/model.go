package tree

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/keybind"
	"github.com/gdamore/tcell/v3"
)

// Tree navigation events.
const (
	treeNone int = iota
	treeHome
	treeEnd
	treeMove
	treeParent
	treeChild
	treeScroll // Move without changing the cursor, even when off screen.
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

	// The movement to be performed during the call to View(), one of the
	// constants defined above.
	movement int

	// The number of nodes to move down or up, when movement is treeMove,
	// excluding non-selectable nodes for cursor movement, including them for
	// scrolling.
	step int

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

	// The visible nodes, top-down, as set by process().
	nodes []*Node

	// Temporarily set to true while we know that the tree has not changed and
	// therefore does not need to be reprocessed.
	stableNodes bool

	// Internal mouse track data.
	lastMouseY int

	keybinds Keybinds
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

// SetRoot sets the root node of the tree.
func (t *Model) SetRoot(root *Node) *Model {
	if t.root != root {
		t.root = root
	}
	return t
}

// GetRoot returns the root node of the tree. If no such node was previously
// set, nil is returned.
func (t *Model) GetRoot() *Node {
	return t.root
}

// SetCurrentNode sets the currently selected node. Provide nil to clear all
// cursors. Selected nodes must be visible and selectable, or else the cursor
// will be changed to the top-most selectable and visible node when the tree is
// next drawn.
func (t *Model) SetCurrentNode(node *Node) *Model {
	if t.currentNode != node {
		t.currentNode = node
	}
	return t
}

// GetCurrentNode returns the currently selected node or nil of no node is
// currently selected.
func (t *Model) GetCurrentNode() *Node {
	return t.currentNode
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
	if t.topLevel != topLevel {
		t.topLevel = topLevel
	}
	return t
}

// SetCenterCursor controls whether the cursor is kept centered whenever
// possible.
func (t *Model) SetCenterCursor(center bool) *Model {
	if t.centerCursor != center {
		t.centerCursor = center
	}
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
	changed := len(t.prefixes) != len(prefixes)
	if !changed {
		for index := range prefixes {
			if t.prefixes[index] != prefixes[index] {
				changed = true
				break
			}
		}
	}
	if changed {
		t.prefixes = prefixes
	}
	return t
}

// SetMarkers sets the strings drawn before node text depending on node state.
// Expanded is used for nodes with children whose children are visible,
// Collapsed is used for nodes with children whose children are hidden, and
// Leaf is used for nodes without children.
func (t *Model) SetMarkers(markers Markers) *Model {
	if t.markers != markers {
		t.markers = markers
	}
	return t
}

// GetMarkers returns the marker strings currently used by this tree view.
func (t *Model) GetMarkers() Markers {
	return t.markers
}

// SetAlign controls the horizontal alignment of the node texts. If set to true,
// all texts except that of top-level nodes will be placed in the same column.
// If set to false, they will indent with the hierarchy.
func (t *Model) SetAlign(align bool) *Model {
	if t.align != align {
		t.align = align
	}
	return t
}

// SetGraphics sets a flag which determines whether or not line graphics are
// drawn to illustrate the tree's hierarchy.
func (t *Model) SetGraphics(showGraphics bool) *Model {
	if t.graphics != showGraphics {
		t.graphics = showGraphics
	}
	return t
}

// SetGraphicsColor sets the colors of the lines used to draw the tree structure.
func (t *Model) SetGraphicsColor(color tcell.Color) *Model {
	if t.graphicsColor != color {
		t.graphicsColor = color
	}
	return t
}

// GetScrollOffset returns the number of node rows that were skipped at the top
// of the tree view. Note that when the user navigates the tree view, this value
// is only updated after the tree view has been redrawn.
func (t *Model) GetScrollOffset() int {
	return t.offsetY
}

// GetRowCount returns the number of "visible" nodes. This includes nodes which
// fall outside the tree view's box but notably does not include the children
// of collapsed nodes. Note that this value is only up to date after the tree
// view has been drawn.
func (t *Model) GetRowCount() int {
	return len(t.nodes)
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
	t.movement = treeMove
	t.step = offset
	t.process(false)
	return t
}

// process builds the visible tree, populates the "nodes" slice, and processes
// pending movement actions. Set "drawingAfter" to true if you know that
// [Model.View] will be called immediately after this function (to avoid
// having [Model.View] call it again).
func (t *Model) process(drawingAfter bool) {
	t.stableNodes = drawingAfter
	_, _, _, height := t.InnerRect()

	// Determine visible nodes and their placement.
	t.nodes = nil
	if t.root == nil {
		return
	}
	parentSelectedIndex, selectedIndex, topLevelGraphicsX := -1, -1, -1
	var graphicsOffset, maxTextX int
	if t.graphics {
		graphicsOffset = 1
	}
	t.root.Walk(func(node, parent *Node) bool {
		// Set node attributes.
		node.parent = parent
		if parent == nil {
			node.level = 0
			node.graphicsX = 0
			node.textX = 0
		} else {
			node.level = parent.level + 1
			node.graphicsX = parent.textX
			node.textX = node.graphicsX + graphicsOffset + node.indent
		}
		if !t.graphics && t.align {
			// Without graphics, we align nodes on the first column.
			node.textX = 0
		}
		if node.level == t.topLevel {
			// No graphics for top level nodes.
			node.graphicsX = 0
			node.textX = 0
		}

		// Add the node to the list.
		if node.level >= t.topLevel {
			// This node will be visible.
			if node.textX > maxTextX {
				maxTextX = node.textX
			}
			if node == t.currentNode && node.selectable {
				selectedIndex = len(t.nodes)

				// Also find parent node.
				for index := len(t.nodes) - 1; index >= 0; index-- {
					if t.nodes[index] == parent && t.nodes[index].selectable {
						parentSelectedIndex = index
						break
					}
				}
			}

			// Maybe we want to skip this level.
			if t.topLevel == node.level && (topLevelGraphicsX < 0 || node.graphicsX < topLevelGraphicsX) {
				topLevelGraphicsX = node.graphicsX
			}

			t.nodes = append(t.nodes, node)
		}

		// Recurse if desired.
		return node.expanded
	})

	// Post-process positions.
	for _, node := range t.nodes {
		// If text must align, we correct the positions.
		if t.align && node.level > t.topLevel {
			node.textX = maxTextX
		}

		// If we skipped levels, shift to the left.
		if topLevelGraphicsX > 0 {
			node.graphicsX -= topLevelGraphicsX
			node.textX -= topLevelGraphicsX
		}
	}

	// Process cursor. (Also trigger events if necessary.)
	if selectedIndex >= 0 {
		// Move the cursor.
		switch t.movement {
		case treeMove:
			for t.step < 0 { // Going up.
				index := selectedIndex
				for index > 0 {
					index--
					if t.nodes[index].selectable {
						selectedIndex = index
						break
					}
				}
				t.step++
			}
			for t.step > 0 { // Going down.
				index := selectedIndex
				for index < len(t.nodes)-1 {
					index++
					if t.nodes[index].selectable {
						selectedIndex = index
						break
					}
				}
				t.step--
			}
		case treeParent:
			if parentSelectedIndex >= 0 {
				selectedIndex = parentSelectedIndex
			}
		case treeChild:
			index := selectedIndex
			for index < len(t.nodes)-1 {
				index++
				if t.nodes[index].selectable && t.nodes[index].parent == t.nodes[selectedIndex] {
					selectedIndex = index
				}
			}
		}
		t.currentNode = t.nodes[selectedIndex]

		// Move cursor into viewport.
		if t.movement != treeScroll {
			if t.centerCursor && height > 0 {
				desired := max(selectedIndex-height/2, 0)
				maxOffset := max(len(t.nodes)-height, 0)
				if desired > maxOffset {
					desired = maxOffset
				}
				t.offsetY = desired
			} else {
				if selectedIndex-t.offsetY >= height {
					t.offsetY = selectedIndex - height + 1
				}
				if selectedIndex < t.offsetY {
					t.offsetY = selectedIndex
				}
			}
			if t.movement != treeHome && t.movement != treeEnd {
				// treeScroll, treeHome, and treeEnd are handled by View().
				t.movement = treeNone
				t.step = 0
			}
		}
	} else {
		// If cursor is not visible or selectable, select the first candidate.
		if t.currentNode != nil {
			for index, node := range t.nodes {
				if node.selectable {
					selectedIndex = index
					t.currentNode = node
					break
				}
			}
		}
		if selectedIndex < 0 {
			t.currentNode = nil
		}
	}
}

// View draws this model onto the screen.
func (t *Model) View(screen tcell.Screen) {
	t.Box.View(screen)
	if t.root == nil {
		return
	}
	_, totalHeight := screen.Size()

	if !t.stableNodes {
		t.process(false)
	} else {
		t.stableNodes = false
	}

	// Scroll the tree, t.movement is treeNone after process() when there is a
	// cursor, except for treeScroll, treeHome, and treeEnd.
	x, y, width, height := t.InnerRect()
	switch t.movement {
	case treeMove:
		t.movement = treeNone
		fallthrough
	case treeScroll:
		t.offsetY += t.step
		t.step = 0
	case treeHome:
		t.offsetY = 0
		t.movement = treeNone
	case treeEnd:
		t.offsetY = len(t.nodes)
		t.movement = treeNone
	}

	if t.offsetY > len(t.nodes)-height {
		t.offsetY = len(t.nodes) - height
	}
	if t.offsetY < 0 {
		t.offsetY = 0
	}

	// Draw the tree.
	posY := y
	borderSet := t.GetBorderSet()
	lineStyle := tcell.StyleDefault.Background(t.GetBackgroundColor()).Foreground(t.graphicsColor)
	for index, node := range t.nodes {
		// Skip invisible parts.
		if posY >= y+height+1 || posY >= totalHeight {
			break
		}
		if index < t.offsetY {
			continue
		}

		// Draw the graphics.
		if t.graphics {
			// Draw ancestor branches.
			ancestor := node.parent
			for ancestor != nil && ancestor.parent != nil && ancestor.parent.level >= t.topLevel {
				if ancestor.graphicsX >= width {
					continue
				}

				// Draw a branch if this ancestor is not a last child.
				if ancestor.parent.children[len(ancestor.parent.children)-1] != ancestor {
					if posY-1 >= y && ancestor.textX > ancestor.graphicsX {
						tview.PrintJoinedSemigraphics(screen, x+ancestor.graphicsX, posY-1, borderSet.Left, lineStyle)
					}
					if posY < y+height {
						screen.Put(x+ancestor.graphicsX, posY, borderSet.Right, lineStyle)
					}
				}
				ancestor = ancestor.parent
			}

			if node.textX > node.graphicsX && node.graphicsX < width {
				// BottomLeft for last child; LeftT for non-last siblings.
				connector := borderSet.BottomLeft
				if node.parent != nil {
					if siblings := node.parent.children; len(siblings) > 0 && siblings[len(siblings)-1] != node {
						connector = borderSet.LeftT
					}
				}

				// Join this node.
				if posY < y+height {
					tview.PrintJoinedSemigraphics(screen, x+node.graphicsX, posY, connector, lineStyle)

					for pos := node.graphicsX + 1; pos < node.textX && pos < width; pos++ {
						screen.Put(x+pos, posY, borderSet.Top, lineStyle)
					}
				}
			}
		}

		// Draw the prefix and the text.
		if node.textX < width && posY < y+height {
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
				_, _, prefixWidth = tview.PrintStyled(screen, t.prefixes[(node.level-t.topLevel)%len(t.prefixes)], x+node.textX, posY, 0, width-node.textX, tview.AlignmentLeft, prefixStyle, true)
			}

			// Marker.
			markerWidth := 0
			if marker != "" && node.textX+prefixWidth < width {
				_, _, markerWidth = tview.PrintStyled(screen, marker, x+node.textX+prefixWidth, posY, 0, width-node.textX-prefixWidth, tview.AlignmentLeft, prefixStyle, true)
			}

			// Text.
			if node.textX+prefixWidth+markerWidth < width {
				if node == t.currentNode {
					posX := 0
					for _, segment := range node.line {
						if posX >= width-node.textX-prefixWidth-markerWidth {
							break
						}
						style := mergeStyle(segment.Style, node.selectedTextStyle)
						_, _, segmentWidth := tview.PrintStyled(
							screen,
							segment.Text,
							x+node.textX+prefixWidth+markerWidth+posX,
							posY,
							0,
							width-node.textX-prefixWidth-markerWidth-posX,
							tview.AlignmentLeft,
							style,
							false,
						)
						posX += segmentWidth
					}
				} else {
					posX := 0
					for _, segment := range node.line {
						if posX >= width-node.textX-prefixWidth-markerWidth {
							break
						}
						_, _, segmentWidth := tview.PrintStyled(
							screen,
							segment.Text,
							x+node.textX+prefixWidth+markerWidth+posX,
							posY,
							0,
							width-node.textX-prefixWidth-markerWidth-posX,
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

func mergeStyle(base, overlay tcell.Style) tcell.Style {
	if fg := overlay.GetForeground(); fg != tcell.ColorDefault {
		base = base.Foreground(fg)
	}
	if bg := overlay.GetBackground(); bg != tcell.ColorDefault {
		base = base.Background(bg)
	}

	if overlay.HasBold() {
		base = base.Bold(true)
	}
	if overlay.HasBlink() {
		base = base.Blink(true)
	}
	if overlay.HasDim() {
		base = base.Dim(true)
	}
	if overlay.HasItalic() {
		base = base.Italic(true)
	}
	if overlay.HasReverse() {
		base = base.Reverse(true)
	}
	if overlay.HasStrikeThrough() {
		base = base.StrikeThrough(true)
	}
	if overlay.HasUnderline() {
		base = base.Underline(true)
	}

	return base
}

func (t *Model) selectCurrentNode() tview.Cmd {
	node := t.currentNode
	if node == nil {
		return nil
	}
	selectedNode := node
	return func() tview.Msg {
		return SelectedMsg{Node: selectedNode}
	}
}

func (t *Model) handleKeyMsg(msg tview.KeyMsg) tview.Cmd {
	// Because the tree is flattened into a list only at drawing time, we also
	// postpone the (cursor) movement to drawing time.
	var selectCmd tview.Cmd
	switch {
	case keybind.Matches(msg, t.keybinds.Down):
		t.movement = treeMove
		t.step = 1
	case keybind.Matches(msg, t.keybinds.Up):
		t.movement = treeMove
		t.step = -1
	case keybind.Matches(msg, t.keybinds.Top):
		t.movement = treeHome
	case keybind.Matches(msg, t.keybinds.Bottom):
		t.movement = treeEnd
	case keybind.Matches(msg, t.keybinds.MoveToLastChild):
		t.movement = treeChild
	case keybind.Matches(msg, t.keybinds.MoveToParent):
		t.movement = treeParent
	case keybind.Matches(msg, t.keybinds.PageDown):
		_, _, _, height := t.InnerRect()
		t.movement = treeMove
		t.step = height
	case keybind.Matches(msg, t.keybinds.PageUp):
		_, _, _, height := t.InnerRect()
		t.movement = treeMove
		t.step = -height
	case keybind.Matches(msg, t.keybinds.Select):
		selectCmd = t.selectCurrentNode()
	}

	t.process(true)
	return selectCmd
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
			t.movement = treeScroll
			t.step = t.lastMouseY - y
			t.lastMouseY = y
		}
	case tview.MouseLeftUp:
		t.lastMouseY = -1
	case tview.MouseLeftClick:
		_, rectY, _, _ := t.InnerRect()
		y += t.offsetY - rectY
		if t.lastMouseY != -1 {
			y += t.lastMouseY - y
			t.lastMouseY = -1
			t.movement = treeNone
		}
		if y >= 0 && y < len(t.nodes) {
			node := t.nodes[y]
			if node.selectable {
				t.currentNode = node
				return tview.Sequence(tview.SetFocus(t), func() tview.Msg {
					return SelectedMsg{Node: node}
				})
			}
		}
		return tview.SetFocus(t)
	case tview.MouseScrollUp:
		t.movement = treeScroll
		t.step = -1
	case tview.MouseScrollDown:
		t.movement = treeScroll
		t.step = 1
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
