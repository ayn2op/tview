package picker

import (
	"github.com/ayn2op/tview/help"
	"github.com/ayn2op/tview/keybind"
	"github.com/ayn2op/tview/list"
)

type Keybinds struct {
	list.Keybinds
	Cancel keybind.Keybind
	Select keybind.Keybind
}

var _ help.KeyMap = Keybinds{}

func DefaultKeybinds() Keybinds {
	return Keybinds{
		Keybinds: list.DefaultKeybinds(),
		Cancel:   keybind.NewSingleKeybind("esc", "cancel"),
		Select:   keybind.NewSingleKeybind("enter", "select"),
	}
}

// defaultKeybinds is built once so that New does not rebuild the keybinds on every View.
var defaultKeybinds = DefaultKeybinds()

func (k Keybinds) ShortHelp() []keybind.Keybind {
	return []keybind.Keybind{k.SelectUp, k.SelectDown, k.Select, k.Cancel}
}

func (k Keybinds) FullHelp() [][]keybind.Keybind {
	return [][]keybind.Keybind{
		{k.SelectUp, k.SelectDown, k.SelectTop, k.SelectBottom},
		{k.Select, k.Cancel},
	}
}

// moves returns the keybinds that move the list.
func (k Keybinds) moves() []keybind.Keybind {
	return []keybind.Keybind{k.SelectUp, k.SelectDown, k.SelectTop, k.SelectBottom, k.ScrollUp, k.ScrollDown, k.ScrollTop, k.ScrollBottom}
}
