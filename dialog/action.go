package dialog

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/button"
	"github.com/ayn2op/tview/keybind"
)

// Action is what a key makes a dialog do.
type Action int

const (
	ActionNone Action = iota
	ActionPrevious
	ActionNext
	ActionPress
	ActionCancel
)

// DefaultKeybind is the default keybind.
func DefaultKeybind(key tview.KeyMsg) Action {
	switch keybind.String(key) {
	case "shift+tab", "up", "left":
		return ActionPrevious
	case "tab", "down", "right":
		return ActionNext
	case "enter":
		return ActionPress
	case "esc":
		return ActionCancel
	}
	return ActionNone
}

// noKeys binds no keys, since the dialog presses its buttons itself.
func noKeys(tview.KeyMsg) button.Action {
	return button.ActionNone
}
