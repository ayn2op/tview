package button

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/keybind"
)

// Action is what a key makes a focused button do.
type Action int

const (
	ActionPress Action = iota
)

// DefaultKeybind binds enter.
func DefaultKeybind(key tview.KeyMsg) (Action, bool) {
	return ActionPress, keybind.String(key) == "enter"
}
