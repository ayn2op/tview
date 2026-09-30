package list

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/keybind"
)

// Action is what a key makes a list do. Apps that handle keys themselves send it with Perform.
type Action int

const (
	ActionSelectUp Action = iota
	ActionSelectDown
	ActionSelectTop
	ActionSelectBottom
	ActionScrollUp
	ActionScrollDown
	ActionScrollTop
	ActionScrollBottom
)

// DefaultKeybind binds up, down, home, and end to moving the selection, and pgup, pgdn, ctrl+home, and ctrl+end to scrolling.
func DefaultKeybind(key tview.KeyMsg) (Action, bool) {
	switch keybind.String(key) {
	case "up":
		return ActionSelectUp, true
	case "down":
		return ActionSelectDown, true
	case "home":
		return ActionSelectTop, true
	case "end":
		return ActionSelectBottom, true
	case "pgup":
		return ActionScrollUp, true
	case "pgdn":
		return ActionScrollDown, true
	case "ctrl+home":
		return ActionScrollTop, true
	case "ctrl+end":
		return ActionScrollBottom, true
	}
	return 0, false
}
