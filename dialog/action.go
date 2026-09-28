package dialog

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/button"
	"github.com/ayn2op/tview/keybind"
)

// Action is what a key makes a dialog do.
type Action int

const (
	ActionNext Action = iota
	ActionPrevious
	ActionPress
	ActionCancel
)

// DefaultKeybind binds tab, down, and right to the next button, shift+tab, up, and left to the previous one, enter to pressing it, and esc to canceling.
func DefaultKeybind(key tview.KeyMsg) (Action, bool) {
	switch keybind.String(key) {
	case "tab", "down", "right":
		return ActionNext, true
	case "shift+tab", "up", "left":
		return ActionPrevious, true
	case "enter":
		return ActionPress, true
	case "esc":
		return ActionCancel, true
	}
	return 0, false
}

// noKeys binds no keys, since the dialog presses its buttons itself.
func noKeys(tview.KeyMsg) (button.Action, bool) {
	return 0, false
}
