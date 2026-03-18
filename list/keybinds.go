package list

import "github.com/ayn2op/tview/keybind"

type Keybinds struct {
	SelectUp     keybind.Keybind
	SelectDown   keybind.Keybind
	SelectTop    keybind.Keybind
	SelectBottom keybind.Keybind

	ScrollUp     keybind.Keybind
	ScrollDown   keybind.Keybind
	ScrollTop    keybind.Keybind
	ScrollBottom keybind.Keybind
}

func DefaultKeybinds() Keybinds {
	return Keybinds{
		SelectUp:     keybind.NewSingleKeybind("up", "up"),
		SelectDown:   keybind.NewSingleKeybind("down", "down"),
		SelectTop:    keybind.NewSingleKeybind("home", "top"),
		SelectBottom: keybind.NewSingleKeybind("end", "bottom"),

		ScrollUp:     keybind.NewSingleKeybind("pgup", "scroll up"),
		ScrollDown:   keybind.NewSingleKeybind("pgdn", "scroll down"),
		ScrollTop:    keybind.NewSingleKeybind("ctrl+home", "scroll top"),
		ScrollBottom: keybind.NewSingleKeybind("ctrl+end", "scroll bottom"),
	}
}

func (l *Model) Keybinds() Keybinds {
	return l.keybinds
}

func (l *Model) SetKeybinds(keybinds Keybinds) *Model {
	l.keybinds = keybinds
	return l
}
