package tree

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/keybind"
)

// Action is what a key makes a tree do. Apps that handle keys themselves send it with Perform.
type Action int

const (
	ActionUp Action = iota
	ActionDown
	ActionTop
	ActionBottom
	ActionMoveToParent
	ActionSelect
)

// DefaultKeybind binds up, down, home, end, K, and enter.
func DefaultKeybind(key tview.KeyMsg) (Action, bool) {
	switch keybind.String(key) {
	case "up":
		return ActionUp, true
	case "down":
		return ActionDown, true
	case "home":
		return ActionTop, true
	case "end":
		return ActionBottom, true
	case "K":
		return ActionMoveToParent, true
	case "enter":
		return ActionSelect, true
	}
	return 0, false
}

// ActionMsg makes a focused tree perform its action.
type ActionMsg Action

// Perform is a command that makes a focused tree perform action.
func Perform(action Action) tview.Cmd {
	return func() tview.Msg { return ActionMsg(action) }
}
