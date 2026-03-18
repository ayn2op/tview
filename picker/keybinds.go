package picker

import (
	"github.com/ayn2op/tview/keybind"
	"github.com/ayn2op/tview/list"
)

type Keybinds struct {
	list.Keybinds

	Cancel keybind.Keybind
	Top    keybind.Keybind
	Bottom keybind.Keybind
	Select keybind.Keybind
}

func DefaultKeybinds() Keybinds {
	return Keybinds{
		Keybinds: list.DefaultKeybinds(),
		Cancel:   keybind.NewSingleKeybind("esc", "cancel"),
		Top:      keybind.NewSingleKeybind("home", "top"),
		Bottom:   keybind.NewSingleKeybind("end", "bottom"),
		Select:   keybind.NewSingleKeybind("enter", "select"),
	}
}

func (m *Model) Keybinds() Keybinds {
	return m.keybinds
}

func (m *Model) SetKeybinds(keybinds Keybinds) *Model {
	m.keybinds = keybinds

	listKeybinds := m.list.Keybinds()
	listKeybinds.SelectUp = keybinds.SelectUp
	listKeybinds.SelectDown = keybinds.SelectDown
	m.list.SetKeybinds(listKeybinds)

	return m
}
