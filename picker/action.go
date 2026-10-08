package picker

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/keybind"
)

// Action is what a key makes a picker do. Keys that move the list are bound with ListKeybind.
type Action int

const (
	ActionNone Action = iota
	ActionSelect
	ActionCancel
)

// DefaultKeybind is the default keybind.
func DefaultKeybind(key tview.KeyMsg) Action {
	switch keybind.String(key) {
	case "enter":
		return ActionSelect
	case "esc":
		return ActionCancel
	}
	return ActionNone
}
