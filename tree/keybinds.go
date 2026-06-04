package tree

import "github.com/ayn2op/tview/keybind"

type Keybinds struct {
	// Navigation
	Up     keybind.Keybind
	Down   keybind.Keybind
	Top    keybind.Keybind
	Bottom keybind.Keybind

	MoveToParent    keybind.Keybind
	MoveToLastChild keybind.Keybind

	PageUp   keybind.Keybind
	PageDown keybind.Keybind

	Select keybind.Keybind
}

func DefaultKeybinds() Keybinds {
	return Keybinds{
		Up:     keybind.NewSingleKeybind("up", "up"),
		Down:   keybind.NewSingleKeybind("down", "down"),
		Top:    keybind.NewSingleKeybind("home", "top"),
		Bottom: keybind.NewSingleKeybind("end", "bot"),

		MoveToParent:    keybind.NewSingleKeybind("K", "parent"),
		MoveToLastChild: keybind.NewSingleKeybind("J", "last child"),

		PageUp:   keybind.NewSingleKeybind("pgup", "page up"),
		PageDown: keybind.NewSingleKeybind("pgdn", "page down"),

		Select: keybind.NewSingleKeybind("enter", "select"),
	}
}

func (t *Model) Keybinds() Keybinds {
	return t.keybinds
}

func (t *Model) SetKeybinds(keybinds Keybinds) *Model {
	t.keybinds = keybinds
	return t
}
