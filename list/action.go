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

	ActionScrollUp
	ActionScrollDown
	ActionScrollTop
	ActionScrollBottom
)

// DefaultKeybind binds up, down, home, and end to moving the selection, and pgup, pgdn, ctrl+home, and ctrl+end to scrolling.
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
	case "pgup":
		return ActionScrollUp
	case "pgdn":
		return ActionScrollDown
	case "ctrl+home":
		return ActionScrollTop
	case "ctrl+end":
		return ActionScrollBottom
	}
	return ActionNone
}
