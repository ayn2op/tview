package list

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/keybind"
)

// Action is what a key makes a list do. Apps that handle keys themselves send it with Perform.
type Action int

const (
	ActionNone Action = iota
	ActionSelectUp
	ActionSelectDown
	ActionSelectTop
	ActionSelectBottom
)

// DefaultKeybind is the default keybind.
func DefaultKeybind(key tview.KeyMsg) Action {
	switch keybind.String(key) {
	case "up":
		return ActionSelectUp
	case "down":
		return ActionSelectDown
	case "home":
		return ActionSelectTop
	case "end":
		return ActionSelectBottom
	}
	return ActionNone
}
