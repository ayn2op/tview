package tree

// SelectionState is a tree's current node and how far it is scrolled. The model owns it and applies the Changes a tree produces with Apply.
type SelectionState struct {
	current *Node
	// offset is the number of rows scrolled off the top.
	offset int
	// center asks the next view to scroll the current node to the middle.
	center bool
	// dragging is set while a drag to scroll is in progress, which last was at row dragY.
	dragging bool
	dragY    int
}

// CurrentNode returns the selected node, or nil for none.
func (s *SelectionState) CurrentNode() *Node {
	return s.current
}

// SetCurrentNode selects node, or none if it is nil, and scrolls it to the middle of the view.
func (s *SelectionState) SetCurrentNode(node *Node) {
	s.current, s.center = node, true
}

// Change is an update to SelectionState produced by a tree, such as moving the cursor or scrolling.
type Change struct {
	current       *Node
	offset, dragY int
	dragging      bool
}

func (s *SelectionState) Apply(change Change) {
	s.current, s.offset, s.dragging, s.dragY, s.center = change.current, change.offset, change.dragging, change.dragY, false
}
