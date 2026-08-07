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
	type entry struct{ node, parent *Node }
	stack := []entry{{node: n}}
	for len(stack) > 0 {
		last := len(stack) - 1
		current := stack[last]
		stack = stack[:last]
		if !callback(current.node, current.parent) {
			continue
		}
		for i := len(current.node.children) - 1; i >= 0; i-- {
			stack = append(stack, entry{current.node.children[i], current.node})
		}
	}
	return n
}

// Reference returns this node's reference object.
func (n *Node) Reference() any {
	return n.reference
}

// SetReference allows you to store a reference of any type in this node. This
// will allow you to establish a mapping between the Model hierarchy and your
// internal tree structure.
func (n *Node) SetReference(reference any) *Node {
	n.reference = reference

	return n
}

// Children returns this node's children.
func (n *Node) Children() []*Node {
	return n.children
}

// SetChildren sets this node's child nodes.
func (n *Node) SetChildren(childNodes []*Node) *Node {
	n.children = childNodes
	return n
}

// Line returns the node's styled text line.
func (n *Node) Line() tview.Line {
	return n.line.Clone()
}

// SetLine sets the node's styled text line.
func (n *Node) SetLine(line tview.Line) *Node {
	n.line = line.Clone()

	return n
}

// ClearChildren removes all child nodes from this node.
func (n *Node) ClearChildren() *Node {
	n.children = nil
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
	n.selectable = selectable
	return n
}

// Expanded returns whether the child nodes of this node are visible.
func (n *Node) Expanded() bool {
	return n.expanded
}

// SetExpanded sets whether or not this node's child nodes should be displayed.
func (n *Node) SetExpanded(expanded bool) *Node {
	n.expanded = expanded
	return n
}

// Expandable returns whether this node can be expanded even when there are
// no loaded child nodes yet.
func (n *Node) Expandable() bool {
	return n.expandable
}

// SetExpandable sets whether this node can be expanded even when there are no
// loaded child nodes yet.
func (n *Node) SetExpandable(expandable bool) *Node {
	n.expandable = expandable
	return n
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

// SelectedTextStyle returns the text style for this node when it is
// selected.
func (n *Node) SelectedTextStyle() tcell.Style {
	return n.selectedTextStyle
}

// SetSelectedTextStyle sets the text style for this node when it is selected.
func (n *Node) SetSelectedTextStyle(style tcell.Style) *Node {
	n.selectedTextStyle = style
	return n
}

// SetIndent sets an additional indentation for this node's text. A value of 0
// keeps the text as far left as possible with a minimum of line graphics. Any
// value greater than that moves the text to the right.
func (n *Node) SetIndent(indent int) *Node {
	n.indent = indent
	return n
}
