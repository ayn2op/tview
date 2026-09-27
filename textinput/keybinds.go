package textinput

import "github.com/ayn2op/tview/keybind"

// Keybinds are the keys a text input edits with. Typing a character inserts it.
type Keybinds struct {
	Left  keybind.Keybind
	Right keybind.Keybind
	Home  keybind.Keybind
	End   keybind.Keybind

	Backspace keybind.Keybind
	Delete    keybind.Keybind

	Submit keybind.Keybind
}

func DefaultKeybinds() Keybinds {
	return Keybinds{
		Left:  keybind.NewSingleKeybind("left", "left"),
		Right: keybind.NewSingleKeybind("right", "right"),
		Home:  keybind.NewSingleKeybind("home", "start"),
		End:   keybind.NewSingleKeybind("end", "end"),

		Backspace: keybind.NewSingleKeybind("backspace", "delete before"),
		Delete:    keybind.NewSingleKeybind("delete", "delete after"),

		Submit: keybind.NewSingleKeybind("enter", "submit"),
	}
}

// defaultKeybinds is built once so that New does not rebuild the keybinds on every View.
var defaultKeybinds = DefaultKeybinds()
