package tabs

import (
	"github.com/ayn2op/tview/keybind"
)

type Keybinds struct {
	Previous keybind.Keybind
	Next     keybind.Keybind
}

func DefaultKeybinds() Keybinds {
	return Keybinds{
		Previous: keybind.NewSingleKeybind("ctrl+h", "prev tab"),
		Next:     keybind.NewSingleKeybind("ctrl+l", "next tab"),
	}
}
