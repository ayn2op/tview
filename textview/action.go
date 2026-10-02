package textview

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/keybind"
)

// Action is what a key makes a focused text view do.
type Action int

const (
	ActionNone Action = iota
	ActionUp
	ActionDown
	ActionLeft
	ActionRight
	ActionTop
	ActionBottom
	ActionPageUp
	ActionPageDown
)

// DefaultKeybind binds the arrows and hjkl, home and g, end and G, and pgup, ctrl+b, pgdn, and ctrl+f.
func DefaultKeybind(key tview.KeyMsg) Action {
	switch keybind.String(key) {
	case "up", "k":
		return ActionUp
	case "down", "j":
		return ActionDown
	case "left", "h":
		return ActionLeft
	case "right", "l":
		return ActionRight
	case "home", "g":
		return ActionTop
	case "end", "G":
		return ActionBottom
	case "pgup", "ctrl+b":
		return ActionPageUp
	case "pgdn", "ctrl+f":
		return ActionPageDown
	}
	return ActionNone
}
