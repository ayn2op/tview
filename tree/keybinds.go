package tree

import "github.com/ayn2op/tview/keybind"

type Keybinds struct {
	// Navigation
	Up     keybind.Keybind
	Down   keybind.Keybind
	Top    keybind.Keybind
	Bottom keybind.Keybind

	MoveToParent keybind.Keybind
	Select       keybind.Keybind
}

func DefaultKeybinds() Keybinds {
	return Keybinds{
		Up:     keybind.NewSingleKeybind("up", "up"),
		Down:   keybind.NewSingleKeybind("down", "down"),
		Top:    keybind.NewSingleKeybind("home", "top"),
		Bottom: keybind.NewSingleKeybind("end", "bot"),

		MoveToParent: keybind.NewSingleKeybind("K", "parent"),
		Select:       keybind.NewSingleKeybind("enter", "select"),
	}
}

// defaultKeybinds is built once so that New does not rebuild the keybinds on every View.
var defaultKeybinds = DefaultKeybinds()
