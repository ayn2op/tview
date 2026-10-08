package tabs

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/keybind"
)

// Action is what a key makes tabs do.
type Action int

const (
	ActionNone Action = iota
	ActionPrevious
	ActionNext
)

// DefaultKeybind is the default keybind.
func DefaultKeybind(key tview.KeyMsg) Action {
	switch keybind.String(key) {
	case "ctrl+h":
		return ActionPrevious
	case "ctrl+l":
		return ActionNext
	}
	return ActionNone
}
