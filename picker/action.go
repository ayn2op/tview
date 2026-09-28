package picker

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/keybind"
)

// Action is what a key makes a picker do. Keys that move the list are bound with ListKeybind.
type Action int

const (
	ActionSelect Action = iota
	ActionCancel
)

// DefaultKeybind binds enter and esc.
func DefaultKeybind(key tview.KeyMsg) (Action, bool) {
	switch keybind.String(key) {
	case "enter":
		return ActionSelect, true
	case "esc":
		return ActionCancel, true
	}
	return 0, false
}
