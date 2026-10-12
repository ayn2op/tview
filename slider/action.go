package slider

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/keybind"
)

// Action is what a key makes a focused slider do.
type Action int

const (
	ActionNone Action = iota
	ActionDecrease
	ActionIncrease
)

// DefaultKeybind is the default keybind.
func DefaultKeybind(key tview.KeyMsg) Action {
	switch keybind.String(key) {
	case "left", "shift+left":
		return ActionDecrease
	case "right", "shift+right":
		return ActionIncrease
	}

	return ActionNone
}
