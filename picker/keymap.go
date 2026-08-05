package picker

import (
	"github.com/ayn2op/tview/help"
	"github.com/ayn2op/tview/keybind"
)

var _ help.KeyMap = (*Model)(nil)

func (m *Model) ShortHelp() []keybind.Keybind {
	return []keybind.Keybind{
		m.keybinds.SelectUp,
		m.keybinds.SelectDown,
		m.keybinds.Select,
		m.keybinds.Cancel,
	}
}

func (m *Model) FullHelp() [][]keybind.Keybind {
	return [][]keybind.Keybind{
		{
			m.keybinds.SelectUp,
			m.keybinds.SelectDown,
			m.keybinds.SelectTop,
			m.keybinds.SelectBottom,
		},
		{
			m.keybinds.Select,
			m.keybinds.Cancel,
		},
	}
}
