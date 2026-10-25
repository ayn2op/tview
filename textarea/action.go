package textarea

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/keybind"
)

// Action is what a key makes a focused text area do. Typing a character inserts it.
type Action int

const (
	ActionNone Action = iota
	ActionLeft
	ActionRight
	ActionUp
	ActionDown
	ActionHome
	ActionEnd
	ActionBackspace
	ActionDelete
	ActionNewline
	// ActionWordLeft and ActionWordRight move over a word, a run of letters and digits, and what is before it.
	ActionWordLeft
	ActionWordRight
	// ActionDeleteWordBackward and ActionDeleteWordForward delete what ActionWordLeft and ActionWordRight move over.
	ActionDeleteWordBackward
	ActionDeleteWordForward
	// ActionDeleteToSpaceBackward deletes back to the whitespace before the cursor, past the whitespace right before it.
	ActionDeleteToSpaceBackward
	// ActionDeleteToLineStart and ActionDeleteToLineEnd delete to the start and end of the line, or the newline there if the cursor is at it.
	ActionDeleteToLineStart
	ActionDeleteToLineEnd
)

// DefaultKeybind is the default keybind. It leaves letters with ctrl or alt, such as readline's, for the app to bind.
func DefaultKeybind(key tview.KeyMsg) Action {
	switch keybind.String(key) {
	case "left":
		return ActionLeft
	case "right":
		return ActionRight
	case "up":
		return ActionUp
	case "down":
		return ActionDown
	case "home":
		return ActionHome
	case "end":
		return ActionEnd
	case "backspace":
		return ActionBackspace
	case "delete":
		return ActionDelete
	case "enter":
		return ActionNewline
	case "ctrl+left", "alt+left":
		return ActionWordLeft
	case "ctrl+right", "alt+right":
		return ActionWordRight
	case "alt+backspace":
		return ActionDeleteWordBackward
	case "ctrl+delete", "alt+delete":
		return ActionDeleteWordForward
	}
	return ActionNone
}
