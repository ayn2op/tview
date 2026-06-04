package tree

import (
	"slices"

	"github.com/ayn2op/tview"
	"github.com/gdamore/tcell/v3"
)

// Node represents one node in a tree view.
type Node struct {
	// The reference object.
	reference any

	// This node's child nodes.
	children []*Node

	// The item's text.
	line tview.Line

	// The style of selected text.
	selectedTextStyle tcell.Style

	// Whether or not this node can be selected.
	selectable bool

	// Whether or not this node's children should be displayed.
	expanded bool

	// Whether or not this node can be expanded, even if children are not loaded yet.
	expandable bool

	// The additional horizontal indent of this node's text.
	indent int

	// The hierarchy level (0 for the root, 1 for its children, and so on). This
	// is only up to date immediately after a call to process() (e.g. via
	// View()).
	level int

	// Temporary member variables.
	parent    *Node // The parent node (nil for the root).
	graphicsX int   // The x-coordinate of the left-most graphics rune.
	textX     int   // The x-coordinate of the first rune of the text.
}

// NewNode returns a new tree node.
func NewNode(text string) *Node {
	textStyle := tcell.StyleDefault.Foreground(tview.Styles.PrimaryTextColor).Background(tview.Styles.PrimitiveBackgroundColor)
	return &Node{
		line:              tview.NewLine(tview.NewSegment(text, textStyle)),
		selectedTextStyle: tcell.StyleDefault.Reverse(true),
		indent:            2,
		expanded:          true,
		expandable:        false,
		selectable:        true,
	}
}

// Walk traverses this node's subtree in depth-first, pre-order (NLR) order and
// calls the provided callback function on each traversed node (which includes
// this node) with the traversed node and its parent node (nil for this node).
// The callback returns whether traversal should continue with the traversed
// node's child nodes (true) or not recurse any deeper (false).
func (n *Node) Walk(callback func(node, parent *Node) bool) *Node {
	n.parent = nil
	nodes := []*Node{n}
	for len(nodes) > 0 {
		// Pop the top node and process it.
		node := nodes[len(nodes)-1]
		nodes = nodes[:len(nodes)-1]
		if !callback(node, node.parent) {
			// Don't add any children.
			continue
		}

		// Add children in reverse order.
		for index := len(node.children) - 1; index >= 0; index-- {
			node.children[index].parent = node
			nodes = append(nodes, node.children[index])
		}
	}

	return n
}

// SetReference allows you to store a reference of any type in this node. This
// will allow you to establish a mapping between the Model hierarchy and your
// internal tree structure.
func (n *Node) SetReference(reference any) *Node {
	n.reference = reference

	return n
}

// GetReference returns this node's reference object.
func (n *Node) GetReference() any {
	return n.reference
}

// SetChildren sets this node's child nodes.
func (n *Node) SetChildren(childNodes []*Node) *Node {
	changed := len(n.children) != len(childNodes)
	if !changed {
		for index := range childNodes {
			if n.children[index] != childNodes[index] {
				changed = true
				break
			}
		}
	}
	if changed {
		n.children = childNodes
	}
	return n
}

// SetLine sets the node's styled text line.
func (n *Node) SetLine(line tview.Line) *Node {
	n.line = line.Clone()

	return n
}

// GetLine returns the node's styled text line.
func (n *Node) GetLine() tview.Line {
	return n.line.Clone()
}

// GetChildren returns this node's children.
func (n *Node) GetChildren() []*Node {
	return n.children
}

// ClearChildren removes all child nodes from this node.
func (n *Node) ClearChildren() *Node {
	if len(n.children) > 0 {
		n.children = nil
	}
	return n
}

// AddChild adds a new child node to this node.
func (n *Node) AddChild(node *Node) *Node {
	n.children = append(n.children, node)

	return n
}

// RemoveChild removes a child node from this node. If the child node cannot be
// found, nothing happens.
func (n *Node) RemoveChild(node *Node) *Node {
	for index, child := range n.children {
		if child == node {
			n.children = slices.Delete(n.children, index, index+1)

			break
		}
	}
	return n
}

// SetSelectable sets a flag indicating whether this node can be selected by
// the user.
func (n *Node) SetSelectable(selectable bool) *Node {
	if n.selectable != selectable {
		n.selectable = selectable
	}
	return n
}

// SetExpanded sets whether or not this node's child nodes should be displayed.
func (n *Node) SetExpanded(expanded bool) *Node {
	if n.expanded != expanded {
		n.expanded = expanded
	}
	return n
}

// SetExpandable sets whether this node can be expanded even when there are no
// loaded child nodes yet.
func (n *Node) SetExpandable(expandable bool) *Node {
	if n.expandable != expandable {
		n.expandable = expandable
	}
	return n
}

// IsExpandable returns whether this node can be expanded even when there are
// no loaded child nodes yet.
func (n *Node) IsExpandable() bool {
	return n.expandable
}

// Expand makes the child nodes of this node appear.
func (n *Node) Expand() *Node {
	if !n.expanded {
		n.expanded = true
	}
	return n
}

// Collapse makes the child nodes of this node disappear.
func (n *Node) Collapse() *Node {
	if n.expanded {
		n.expanded = false
	}
	return n
}

// ExpandAll expands this node and all descendent nodes.
func (n *Node) ExpandAll() *Node {
	n.Walk(func(node, parent *Node) bool {
		if !node.expanded {
			node.expanded = true
		}
		return true
	})
	return n
}

// CollapseAll collapses this node and all descendent nodes.
func (n *Node) CollapseAll() *Node {
	n.Walk(func(node, parent *Node) bool {
		if node.expanded {
			node.expanded = false
		}
		return true
	})
	return n
}

// IsExpanded returns whether the child nodes of this node are visible.
func (n *Node) IsExpanded() bool {
	return n.expanded
}

// SetSelectedTextStyle sets the text style for this node when it is selected.
func (n *Node) SetSelectedTextStyle(style tcell.Style) *Node {
	if n.selectedTextStyle != style {
		n.selectedTextStyle = style
	}
	return n
}

// GetSelectedTextStyle returns the text style for this node when it is
// selected.
func (n *Node) GetSelectedTextStyle() tcell.Style {
	return n.selectedTextStyle
}

// SetIndent sets an additional indentation for this node's text. A value of 0
// keeps the text as far left as possible with a minimum of line graphics. Any
// value greater than that moves the text to the right.
func (n *Node) SetIndent(indent int) *Node {
	if n.indent != indent {
		n.indent = indent
	}
	return n
}

// GetLevel returns the node's level within the hierarchy, where 0 corresponds
// to the root node, 1 corresponds to its children, and so on. This is only
// guaranteed to be up to date immediately after the tree that contains this
// node is drawn.
func (n *Node) GetLevel() int {
	return n.level
}
