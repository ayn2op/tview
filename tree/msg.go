package tree

import "github.com/ayn2op/tview"

// ActionMsg makes a focused tree perform its action.
type ActionMsg Action

// Perform is a command that makes a focused tree perform action.
func Perform(action Action) tview.Cmd {
	return func() tview.Msg { return ActionMsg(action) }
}
