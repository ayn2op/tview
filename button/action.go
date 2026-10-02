package button

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/keybind"
)

// Action is what a key makes a focused button do.
type Action int

const (
	ActionNone Action = iota
	ActionPress
)

// DefaultKeybind binds enter.
func DefaultKeybind(key tview.KeyMsg) Action {
	if keybind.String(key) == "enter" {
		return ActionPress
	}
	return ActionNone
}
