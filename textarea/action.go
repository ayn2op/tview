package textarea

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/keybind"
)

// Action is what a key makes a focused text area do. Typing a character inserts it.
type Action int

const (
	ActionNone Action = iota
	ActionLeft
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
func DefaultKeybind(key tview.KeyMsg) Action {
	switch keybind.String(key) {
	case "left":
		return ActionLeft
	case "right":
		return ActionRight
	case "up":
		return ActionUp
	case "down":
		return ActionDown
	case "home":
		return ActionHome
	case "end":
		return ActionEnd
	case "backspace":
		return ActionBackspace
	case "delete":
		return ActionDelete
	case "enter":
		return ActionNewline
	}
	return ActionNone
}
