package textview

import "github.com/ayn2op/tview/keybind"

// Keybinds are the keys a text view scrolls with.
type Keybinds struct {
	Up    keybind.Keybind
	Down  keybind.Keybind
	Left  keybind.Keybind
	Right keybind.Keybind

	Top      keybind.Keybind
	Bottom   keybind.Keybind
	PageUp   keybind.Keybind
	PageDown keybind.Keybind
}

func DefaultKeybinds() Keybinds {
	return Keybinds{
		Up:    keybind.New("up", "k").WithHelp("↑/k", "up"),
		Down:  keybind.New("down", "j").WithHelp("↓/j", "down"),
		Left:  keybind.New("left", "h").WithHelp("←/h", "left"),
		Right: keybind.New("right", "l").WithHelp("→/l", "right"),

		Top:      keybind.New("home", "g").WithHelp("home/g", "top"),
		Bottom:   keybind.New("end", "G").WithHelp("end/G", "bottom"),
		PageUp:   keybind.New("pgup", "ctrl+b").WithHelp("pgup/ctrl+b", "page up"),
		PageDown: keybind.New("pgdn", "ctrl+f").WithHelp("pgdn/ctrl+f", "page down"),
	}
}

// defaultKeybinds is built once so that New does not rebuild the keybinds on every View.
var defaultKeybinds = DefaultKeybinds()
