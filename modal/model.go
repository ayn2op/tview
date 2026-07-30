package modal

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/frame"
	"github.com/gdamore/tcell/v3"
	"github.com/rivo/uniseg"
)

type DoneMsg struct {
	ButtonIndex int
	ButtonLabel string
}

type Model struct {
	*tview.Box
	form      *tview.Form
	frame     *frame.Model
	text      string
	textColor tcell.Color
	buttons   []string
}

func NewModel() *Model {
	form := tview.NewForm().SetButtonsAlignment(tview.AlignmentCenter)
	form.SetBackgroundColor(tview.Styles.ContrastBackgroundColor).SetBorderPadding(0, 0, 0, 0)
	frame := frame.NewModel(form).SetBorders(0, 0, 1, 0, 0, 0)
	frame.SetBackgroundColor(tview.Styles.ContrastBackgroundColor).SetBorderPadding(1, 1, 1, 1)
	return &Model{
		Box:       tview.NewBox().SetBorders(tview.BordersAll).SetBackgroundColor(tview.Styles.ContrastBackgroundColor),
		form:      form,
		frame:     frame,
		textColor: tview.Styles.PrimaryTextColor,
	}
}

func (m *Model) SetBackgroundColor(color tcell.Color) *Model {
	m.Box.SetBackgroundColor(color)
	m.form.SetBackgroundColor(color)
	m.frame.SetBackgroundColor(color)
	return m
}

func (m *Model) SetTextColor(color tcell.Color) *Model {
	m.textColor = color
	return m
}

func (m *Model) SetButtonStyle(style tcell.Style) *Model {
	m.form.SetButtonStyle(style)
	return m
}

func (m *Model) SetButtonActivatedStyle(style tcell.Style) *Model {
	m.form.SetButtonActivatedStyle(style)
	return m
}

func (m *Model) SetText(text string) *Model {
	m.text = text
	return m
}

func (m *Model) AddButtons(labels []string) *Model {
	m.buttons = append(m.buttons, labels...)
	for _, label := range labels {
		m.form.AddButton(label)
	}
	return m
}

func (m *Model) ClearButtons() *Model {
	m.buttons = nil
	m.form.ClearButtons()
	return m
}

func (m *Model) SetFocus(index int) *Model {
	m.form.SetFocus(index)
	return m
}

func (m *Model) Focus(delegate func(tview.Model)) {
	delegate(m.form)
}

func (m *Model) HasFocus() bool {
	return m.form.HasFocus()
}

func (m *Model) View(screen tcell.Screen) {
	x, y, availableWidth, availableHeight := m.Rect()
	if availableWidth <= 0 || availableHeight <= 0 {
		return
	}

	maxContentWidth := max(availableWidth-4, 1)
	contentWidth := min(80, maxContentWidth)
	buttonsWidth := 0
	for _, label := range m.buttons {
		buttonsWidth += uniseg.StringWidth(label) + 6
	}
	if buttonsWidth > 0 {
		buttonsWidth -= 2
	}
	contentWidth = min(max(contentWidth, buttonsWidth), maxContentWidth)

	lines := tview.WordWrap(m.text, contentWidth)
	lines = lines[:min(len(lines), max(availableHeight-6, 0))]
	m.frame.Clear()
	for _, line := range lines {
		m.frame.AddText(line, true, tview.AlignmentCenter, m.textColor)
	}

	width := min(contentWidth+4, availableWidth)
	height := min(len(lines)+6, availableHeight)
	m.SetRect(x+(availableWidth-width)/2, y+(availableHeight-height)/2, width, height)
	m.Box.View(screen)
	x, y, width, height = m.InnerRect()
	m.frame.SetRect(x, y, width, height)
	m.frame.View(screen)
}

func (m *Model) Update(msg tview.Msg) tview.Cmd {
	switch msg := msg.(type) {
	case tview.FormSubmitMsg:
		return func() tview.Msg { return DoneMsg(msg) }
	case tview.FormCancelMsg:
		return func() tview.Msg { return DoneMsg{ButtonIndex: -1} }
	case tview.ButtonExitMsg:
		return m.form.Update(msg)
	case tview.MouseMsg:
		if m.form.InRect(msg.Position()) {
			return m.form.Update(msg)
		}
		if msg.Action == tview.MouseLeftDown && m.InRect(msg.Position()) {
			return tview.SetFocus(m)
		}
	case tview.KeyMsg:
		switch msg.Key() {
		case tcell.KeyDown, tcell.KeyRight:
			msg = tcell.NewEventKey(tcell.KeyTab, "", tcell.ModNone)
		case tcell.KeyUp, tcell.KeyLeft:
			msg = tcell.NewEventKey(tcell.KeyBacktab, "", tcell.ModNone)
		}
		return m.frame.Update(msg)
	case tview.PasteMsg:
		return m.frame.Update(msg)
	}
	return nil
}
