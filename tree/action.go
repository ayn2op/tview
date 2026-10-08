package tree

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/keybind"
)

// Action is what a key makes a tree do. Apps that handle keys themselves send it with Perform.
type Action int

const (
	ActionNone Action = iota

	ActionUp
	ActionDown
	ActionTop
	ActionBottom
	ActionMoveToParent
	ActionSelect
)

// DefaultKeybind is the default keybind.
func DefaultKeybind(key tview.KeyMsg) Action {
	switch keybind.String(key) {
	case "up":
		return ActionUp
	case "down":
		return ActionDown
	case "home":
		return ActionTop
	case "end":
		return ActionBottom
	case "K":
		return ActionMoveToParent
	case "enter":
		return ActionSelect
	}
	return ActionNone
}
