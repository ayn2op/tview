package textarea

import "github.com/ayn2op/tview/keybind"

// Keybinds are the keys a text area edits with. Typing a character inserts it.
type Keybinds struct {
	Left  keybind.Keybind
	Right keybind.Keybind
	Up    keybind.Keybind
	Down  keybind.Keybind
	Home  keybind.Keybind
	End   keybind.Keybind

	Backspace keybind.Keybind
	Delete    keybind.Keybind
	Newline   keybind.Keybind
}

func DefaultKeybinds() Keybinds {
	return Keybinds{
		Left:  keybind.NewSingleKeybind("left", "left"),
		Right: keybind.NewSingleKeybind("right", "right"),
		Up:    keybind.NewSingleKeybind("up", "up"),
		Down:  keybind.NewSingleKeybind("down", "down"),
		Home:  keybind.NewSingleKeybind("home", "line start"),
		End:   keybind.NewSingleKeybind("end", "line end"),

		Backspace: keybind.NewSingleKeybind("backspace", "delete before"),
		Delete:    keybind.NewSingleKeybind("delete", "delete after"),
		Newline:   keybind.NewSingleKeybind("enter", "newline"),
	}
}

// defaultKeybinds is built once so that New does not rebuild the keybinds on every View.
var defaultKeybinds = DefaultKeybinds()
