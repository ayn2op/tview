package button

import "github.com/ayn2op/tview/keybind"

// Keybinds are the keys a focused button responds to.
type Keybinds struct {
	Press keybind.Keybind
}

func DefaultKeybinds() Keybinds {
	return Keybinds{
		Press: keybind.NewSingleKeybind("enter", "press"),
	}
}

// defaultKeybinds is built once so that New does not rebuild the keybinds on every View.
var defaultKeybinds = DefaultKeybinds()
