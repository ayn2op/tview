package list

import "github.com/ayn2op/tview"

// ActionMsg makes a focused list perform its action.
type ActionMsg Action

// Perform is a command that makes a focused list perform action.
func Perform(action Action) tview.Cmd {
	return func() tview.Msg { return ActionMsg(action) }
}
