package tabs

import (
	"github.com/ayn2op/tview/help"
	"github.com/ayn2op/tview/keybind"
)

// Keybinds are the keys that switch tabs.
type Keybinds struct {
	Previous keybind.Keybind
	Next     keybind.Keybind
}

var _ help.KeyMap = Keybinds{}

func DefaultKeybinds() Keybinds {
	return Keybinds{
		Previous: keybind.NewSingleKeybind("ctrl+h", "prev tab"),
		Next:     keybind.NewSingleKeybind("ctrl+l", "next tab"),
	}
}

// defaultKeybinds is built once so that New does not rebuild the keybinds on every View.
var defaultKeybinds = DefaultKeybinds()

func (k Keybinds) ShortHelp() []keybind.Keybind {
	return []keybind.Keybind{k.Previous, k.Next}
}

func (k Keybinds) FullHelp() [][]keybind.Keybind {
	return [][]keybind.Keybind{{k.Previous, k.Next}}
}
