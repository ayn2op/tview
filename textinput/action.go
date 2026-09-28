package textinput

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/keybind"
)

// Action is what a key makes a focused text input do. Typing a character inserts it.
type Action int

const (
	ActionLeft Action = iota
	ActionRight
	ActionHome
	ActionEnd
	ActionBackspace
	ActionDelete
	ActionSubmit
)

// DefaultKeybind binds the keys of the same names, and enter to submitting.
func DefaultKeybind(key tview.KeyMsg) (Action, bool) {
	switch keybind.String(key) {
	case "left":
		return ActionLeft, true
	case "right":
		return ActionRight, true
	case "home":
		return ActionHome, true
	case "end":
		return ActionEnd, true
	case "backspace":
		return ActionBackspace, true
	case "delete":
		return ActionDelete, true
	case "enter":
		return ActionSubmit, true
	}
	return 0, false
}
