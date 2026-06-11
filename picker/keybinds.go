package picker

import (
	"github.com/ayn2op/tview/keybind"
	"github.com/ayn2op/tview/list"
)

type Keybinds struct {
	list.Keybinds
	Cancel keybind.Keybind
	Select keybind.Keybind
}

func DefaultKeybinds() Keybinds {
	return Keybinds{
		Keybinds: list.DefaultKeybinds(),
		Cancel:   keybind.NewSingleKeybind("esc", "cancel"),
		Select:   keybind.NewSingleKeybind("enter", "select"),
	}
}

func (m *Model) Keybinds() Keybinds {
	return m.keybinds
}

func (m *Model) SetKeybinds(keybinds Keybinds) *Model {
	m.list.SetKeybinds(keybinds.Keybinds)
	m.keybinds = keybinds
	return m
}
