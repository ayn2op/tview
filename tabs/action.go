package tabs

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/keybind"
)

// Action is what a key makes tabs do.
type Action int

const (
	ActionPrevious Action = iota
	ActionNext
)

// DefaultKeybind binds ctrl+h and ctrl+l.
func DefaultKeybind(key tview.KeyMsg) (Action, bool) {
	switch keybind.String(key) {
	case "ctrl+h":
		return ActionPrevious, true
	case "ctrl+l":
		return ActionNext, true
	}
	return 0, false
}
