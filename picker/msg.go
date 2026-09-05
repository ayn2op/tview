package picker

import (
	"github.com/ayn2op/tview"
)

type SelectedMsg struct {
	Item
}

func (m *Model) selectItem() tview.Cmd {
	index := m.list.Cursor()
	if index >= 0 && index < len(m.filtered) {
		item := m.filtered[index]
		return func() tview.Msg {
			return SelectedMsg{Item: item}
		}
	}
	return nil
}

type CancelMsg struct{}

func cancel() tview.Cmd {
	return func() tview.Msg {
		return CancelMsg{}
	}
}
