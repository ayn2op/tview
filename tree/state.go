package tree

// SelectionState is a tree's current node. The model owns it and applies the Changes a tree produces with Apply.
type SelectionState struct {
	current *Node
}

// CurrentNode returns the selected node, or nil for none.
func (s *SelectionState) CurrentNode() *Node {
	return s.current
}

// SetCurrentNode selects node, or none if it is nil.
func (s *SelectionState) SetCurrentNode(node *Node) {
	s.current = node
}

// Change is an update to SelectionState produced by a tree.
type Change struct {
	current *Node
}

func (s *SelectionState) Apply(change Change) {
	s.current = change.current
}
