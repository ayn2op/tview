package cmd

// Batch combines multiple commands into a single command.
type Batch []Cmd

func (Batch) isCmd() {}
