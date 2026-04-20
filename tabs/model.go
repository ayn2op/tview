package tabs

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/keybind"
	"github.com/gdamore/tcell/v3"
)

type Tab interface {
	tview.Model
	Label() string
}

type Model struct {
	*tview.Box
	keybinds Keybinds

	tabs   []Tab
	active int

	labelAlignment   tview.Alignment
	labelStyle       tcell.Style
	activeLabelStyle tcell.Style
}

func NewModel(tabs []Tab) *Model {
	return &Model{
		Box:      tview.NewBox(),
		keybinds: DefaultKeybinds(),

		tabs: tabs,

		labelAlignment:   tview.AlignmentCenter,
		labelStyle:       tcell.StyleDefault,
		activeLabelStyle: tcell.StyleDefault.Reverse(true),
	}
}

func (m *Model) canPrevious() bool {
	return len(m.tabs) > 0 && m.active > 0
}

func (m *Model) canNext() bool {
	return len(m.tabs) > 0 && m.active < len(m.tabs)-1
}

func (m *Model) Previous() {
	m.active = max(m.active-1, 0)
}

func (m *Model) Next() {
	m.active = min(m.active+1, len(m.tabs)-1)
}

func (m *Model) Focus(delegate func(m tview.Model)) {
	if len(m.tabs) == 0 {
		m.Box.Focus(delegate)
		return
	}
	delegate(m.tabs[m.active]) // delegate to active tab
}

func (m *Model) HasFocus() bool {
	if len(m.tabs) > 0 && m.tabs[m.active].HasFocus() {
		return true
	}
	return m.Box.HasFocus()
}

func (m *Model) Blur() {
	if len(m.tabs) > 0 {
		m.tabs[m.active].Blur()
	}
	m.Box.Blur()
}

func (m *Model) Update(msg tview.Msg) tview.Cmd {
	if len(m.tabs) == 0 {
		return nil
	}

	switch msg := msg.(type) {
	case tview.InitMsg:
		return m.activateTab()
	case tview.KeyMsg:
		switch {
		case keybind.Matches(msg, m.keybinds.Previous):
			if !m.canPrevious() {
				return m.tabs[m.active].Update(msg)
			}
			m.Previous()
			return m.activateTab()
		case keybind.Matches(msg, m.keybinds.Next):
			if !m.canNext() {
				return m.tabs[m.active].Update(msg)
			}
			m.Next()
			return m.activateTab()
		}
	case tview.MouseMsg:
		x, y := msg.Position()
		if !m.InRect(x, y) {
			return nil
		}

		if msg.Action == tview.MouseLeftDown {
			return tview.SetFocus(m)
		}

		if tab, ok := m.tabAt(x, y); ok {
			switch msg.Action {
			case tview.MouseLeftClick:
				if tab == m.active {
					return tview.SetFocus(m)
				}
				m.active = tab
				return m.activateTab()
			case tview.MouseScrollUp, tview.MouseScrollLeft:
				if !m.canPrevious() {
					return nil
				}
				m.Previous()
				return m.activateTab()
			case tview.MouseScrollDown, tview.MouseScrollRight:
				if !m.canNext() {
					return nil
				}
				m.Next()
				return m.activateTab()
			}
		}
	}
	return m.tabs[m.active].Update(msg)
}

func (m *Model) Draw(screen tcell.Screen) {
	m.DrawForSubclass(screen, m)

	if len(m.tabs) == 0 {
		return
	}

	x, y, width, height := m.InnerRect()
	tmpX := x
	var content tview.Model
	for i, tab := range m.tabs {
		labelStyle := m.labelStyle
		if i == m.active {
			content = tab
			labelStyle = m.activeLabelStyle
		}

		label := tab.Label()
		tview.PrintWithStyle(screen, label, tmpX, y, width, m.labelAlignment, labelStyle)
		tmpX += len(label) + 1
	}

	if content != nil {
		y++
		height--

		content.SetRect(x, y, width, height)
		content.Draw(screen)
	}
}

func (m *Model) activateTab() tview.Cmd {
	return tview.Batch(
		m.tabs[m.active].Update(tview.InitMsg{}),
		tview.SetFocus(m),
	)
}

func (m *Model) tabAt(x, y int) (int, bool) {
	innerX, innerY, width, _ := m.InnerRect()
	if y != innerY {
		return 0, false
	}

	tmpX := innerX
	for i, tab := range m.tabs {
		labelWidth := len(tab.Label())
		labelX := tmpX
		switch m.labelAlignment {
		case tview.AlignmentCenter:
			labelX = tmpX + width/2 - labelWidth/2
		case tview.AlignmentRight:
			labelX = tmpX + width - labelWidth
		}
		if x >= labelX && x < labelX+labelWidth {
			return i, true
		}
		tmpX += labelWidth + 1
	}

	return 0, false
}
