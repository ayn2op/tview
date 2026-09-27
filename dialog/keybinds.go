package dialog

import (
	"github.com/ayn2op/tview/button"
	"github.com/ayn2op/tview/keybind"
)

// Keybinds are the keys a dialog moves the focus between its buttons and cancels with. The embedded button keybinds press the focused button.
type Keybinds struct {
	button.Keybinds
	Next     keybind.Keybind
	Previous keybind.Keybind
	Cancel   keybind.Keybind
}

func DefaultKeybinds() Keybinds {
	return Keybinds{
		Keybinds: button.DefaultKeybinds(),
		Next:     keybind.New("tab", "down", "right").WithHelp("tab", "next"),
		Previous: keybind.New("shift+tab", "up", "left").WithHelp("shift+tab", "previous"),
		Cancel:   keybind.NewSingleKeybind("esc", "cancel"),
	}
}

// defaultKeybinds is built once so that New does not rebuild the keybinds on every View.
var defaultKeybinds = DefaultKeybinds()
