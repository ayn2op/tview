package textview

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/keybind"
)

// Action is what a key makes a focused text view do.
type Action int

const (
	ActionUp Action = iota
	ActionDown
	ActionLeft
	ActionRight
	ActionTop
	ActionBottom
	ActionPageUp
	ActionPageDown
)

// DefaultKeybind binds the arrows and hjkl, home and g, end and G, and pgup, ctrl+b, pgdn, and ctrl+f.
func DefaultKeybind(key tview.KeyMsg) (Action, bool) {
	switch keybind.String(key) {
	case "up", "k":
		return ActionUp, true
	case "down", "j":
		return ActionDown, true
	case "left", "h":
		return ActionLeft, true
	case "right", "l":
		return ActionRight, true
	case "home", "g":
		return ActionTop, true
	case "end", "G":
		return ActionBottom, true
	case "pgup", "ctrl+b":
		return ActionPageUp, true
	case "pgdn", "ctrl+f":
		return ActionPageDown, true
	}
	return 0, false
}
