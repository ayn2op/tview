package tabs

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/help"
	"github.com/ayn2op/tview/keybind"
	"github.com/gdamore/tcell/v3"
)

type Tab interface {
	tview.Primitive
	Label() string
}

type Model struct {
	*tview.Box
	Keybinds Keybinds

	tabs   []Tab
	active int

	labelAlignment   tview.Alignment
	labelStyle       tcell.Style
	activeLabelStyle tcell.Style
}

func NewModel(tabs []Tab) *Model {
	return &Model{
		Box:      tview.NewBox(),
		Keybinds: DefaultKeybinds(),

		tabs: tabs,

		labelAlignment:   tview.AlignmentCenter,
		labelStyle:       tcell.StyleDefault,
		activeLabelStyle: tcell.StyleDefault.Reverse(true),
	}
}

var _ tview.Primitive = (*Model)(nil)
var _ help.KeyMap = (*Model)(nil)

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

func (m *Model) Focus(delegate func(p tview.Primitive)) {
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

func (m *Model) HandleEvent(event tcell.Event) tview.Command {
	if len(m.tabs) == 0 {
		return nil
	}

	switch event := event.(type) {
	case *tview.InitEvent:
		return m.activateTab()
	case *tview.KeyEvent:
		switch {
		case keybind.Matches(event, m.Keybinds.Previous):
			if !m.canPrevious() {
				return m.tabs[m.active].HandleEvent(event)
			}
			m.Previous()
			return m.activateTab()
		case keybind.Matches(event, m.Keybinds.Next):
			if !m.canNext() {
				return m.tabs[m.active].HandleEvent(event)
			}
			m.Next()
			return m.activateTab()
		}
	}
	return m.tabs[m.active].HandleEvent(event)
}

func (m *Model) Draw(screen tcell.Screen) {
	m.DrawForSubclass(screen, m)

	if len(m.tabs) == 0 {
		return
	}

	x, y, width, height := m.GetInnerRect()
	tmpX := x
	var content tview.Primitive
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

func (m *Model) activateTab() tview.Command {
	return tview.BatchCommand{
		m.tabs[m.active].HandleEvent(tview.NewInitEvent()),
		tview.SetFocus(m),
	}
}

func (m *Model) ShortHelp() []keybind.Keybind {
	if len(m.tabs) == 0 {
		return nil
	}

	var short []keybind.Keybind
	if m.canPrevious() {
		short = append(short, m.Keybinds.Previous)
	}
	if m.canNext() {
		short = append(short, m.Keybinds.Next)
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
		nav = append(nav, m.Keybinds.Previous)
	}
	if m.canNext() {
		nav = append(nav, m.Keybinds.Next)
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
