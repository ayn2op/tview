package textarea

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/keybind"
)

// Action is what a key makes a focused text area do. Typing a character inserts it.
type Action int

const (
	ActionLeft Action = iota
	ActionRight
	ActionUp
	ActionDown
	ActionHome
	ActionEnd
	ActionBackspace
	ActionDelete
	ActionNewline
)

// DefaultKeybind binds the keys of the same names, and enter to a newline.
func DefaultKeybind(key tview.KeyMsg) (Action, bool) {
	switch keybind.String(key) {
	case "left":
		return ActionLeft, true
	case "right":
		return ActionRight, true
	case "up":
		return ActionUp, true
	case "down":
		return ActionDown, true
	case "home":
		return ActionHome, true
	case "end":
		return ActionEnd, true
	case "backspace":
		return ActionBackspace, true
	case "delete":
		return ActionDelete, true
	case "enter":
		return ActionNewline, true
	}
	return 0, false
}
