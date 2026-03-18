package tabs

import (
	"github.com/ayn2op/tview/help"
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

func (m *Model) Keybinds() Keybinds {
	return m.keybinds
}

func (m *Model) SetKeybinds(keybinds Keybinds) *Model {
	m.keybinds = keybinds
	return m
}

var _ help.KeyMap = (*Model)(nil)

func (m *Model) ShortHelp() []keybind.Keybind {
	if len(m.tabs) == 0 {
		return nil
	}

	var short []keybind.Keybind
	if m.canPrevious() {
		short = append(short, m.keybinds.Previous)
	}
	if m.canNext() {
		short = append(short, m.keybinds.Next)
	}
	if activeKeyMap, ok := m.tabs[m.active].(help.KeyMap); ok {
		short = append(short, activeKeyMap.ShortHelp()...)
	}
	return short
}

func (m *Model) FullHelp() [][]keybind.Keybind {
	if len(m.tabs) == 0 {
		return nil
	}

	var nav []keybind.Keybind
	if m.canPrevious() {
		nav = append(nav, m.keybinds.Previous)
	}
	if m.canNext() {
		nav = append(nav, m.keybinds.Next)
	}
	var full [][]keybind.Keybind
	if len(nav) > 0 {
		full = append(full, nav)
	}
	if activeKeyMap, ok := m.tabs[m.active].(help.KeyMap); ok {
		full = append(full, activeKeyMap.FullHelp()...)
	}
	return full
}
