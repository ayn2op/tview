package tview

import (
	"github.com/gdamore/tcell/v3"
)

// Modal is a centered message window used to inform the user or prompt them
// for an immediate decision. It needs to have at least one button (added via
// [Modal.AddButtons]) or it will never disappear.
//
// See https://github.com/ayn2op/tview/wiki/Modal for an example.
type Modal struct {
	*Box

	// The frame embedded in the modal.
	frame *Frame

	// The form embedded in the modal's frame.
	form *Form

	// The message text (original, not word-wrapped).
	text string

	// The text color.
	textColor tcell.Color
}

type ModalDoneEvent struct {
	tcell.EventTime
	ButtonIndex int
	ButtonLabel string
}

func newModalDoneEvent(buttonIndex int, buttonLabel string) *ModalDoneEvent {
	return &ModalDoneEvent{
		ButtonIndex: buttonIndex,
		ButtonLabel: buttonLabel,
	}
}

// NewModal returns a new modal message window.
func NewModal() *Modal {
	m := &Modal{
		Box:       NewBox().SetBorders(BordersAll).SetBackgroundColor(Styles.ContrastBackgroundColor),
		textColor: Styles.PrimaryTextColor,
	}
	m.form = NewForm().SetButtonsAlignment(AlignmentCenter)
	m.form.SetBackgroundColor(Styles.ContrastBackgroundColor).SetBorderPadding(0, 0, 0, 0)
	m.frame = NewFrame(m.form).SetBorders(0, 0, 1, 0, 0, 0)
	m.frame.SetBackgroundColor(Styles.ContrastBackgroundColor).
		SetBorderPadding(1, 1, 1, 1)
	return m
}

// SetBackgroundColor sets the color of the modal frame background.
func (m *Modal) SetBackgroundColor(color tcell.Color) *Modal {
	if m.form.GetBackgroundColor() != color || m.frame.GetBackgroundColor() != color {
		m.form.SetBackgroundColor(color)
		m.frame.SetBackgroundColor(color)
	}
	return m
}

// SetTextColor sets the color of the message text.
func (m *Modal) SetTextColor(color tcell.Color) *Modal {
	if m.textColor != color {
		m.textColor = color
	}
	return m
}

// SetButtonStyle sets the style of the buttons when they are not focused.
func (m *Modal) SetButtonStyle(style tcell.Style) *Modal {
	m.form.SetButtonStyle(style)
	return m
}

// SetButtonActivatedStyle sets the style of the buttons when they are focused.
func (m *Modal) SetButtonActivatedStyle(style tcell.Style) *Modal {
	m.form.SetButtonActivatedStyle(style)
	return m
}

// SetText sets the message text of the window. The text may contain line
// breaks. Note that words are wrapped, too, based on the final size of the
// window.
func (m *Modal) SetText(text string) *Modal {
	if m.text != text {
		m.text = text
	}
	return m
}

// AddButtons adds buttons to the window.
func (m *Modal) AddButtons(labels []string) *Modal {
	if len(labels) == 0 {
		return m
	}
	for _, label := range labels {
		m.form.AddButton(label)
	}
	return m
}

// ClearButtons removes all buttons from the window.
func (m *Modal) ClearButtons() *Modal {
	m.form.ClearButtons()
	return m
}

// SetFocus shifts the focus to the button with the given index.
func (m *Modal) SetFocus(index int) *Modal {
	m.form.SetFocus(index)
	return m
}

// Focus is called when this primitive receives focus.
func (m *Modal) Focus(delegate func(p Primitive)) {
	delegate(m.form)
}

// HasFocus returns whether or not this primitive has focus.
func (m *Modal) HasFocus() bool {
	return m.form.HasFocus()
}

// Draw draws this primitive onto the screen.
func (m *Modal) Draw(screen tcell.Screen) {
	// Calculate the width of this modal.
	buttonsWidth := 0
	for _, button := range m.form.buttons {
		buttonsWidth += TaggedStringWidth(button.text) + 4 + 2
	}
	buttonsWidth -= 2
	screenWidth, screenHeight := screen.Size()
	width := max(screenWidth/3, buttonsWidth)
	// width is now without the box border.

	// Reset the text and find out how wide it is.
	m.frame.Clear()
	lines := WordWrap(m.text, width)
	for _, line := range lines {
		m.frame.AddText(line, true, AlignmentCenter, m.textColor)
	}

	// Set the modal's position and size.
	height := len(lines) + 6
	width += 4
	x := (screenWidth - width) / 2
	y := (screenHeight - height) / 2
	m.SetRect(x, y, width, height)

	// Draw the frame.
	m.DrawForSubclass(screen, m)
	x, y, width, height = m.InnerRect()
	m.frame.SetRect(x, y, width, height)
	m.frame.Draw(screen)
}

// HandleEvent handles input events for this primitive.
func (m *Modal) HandleEvent(event tcell.Event) Command {
	switch event := event.(type) {
	case *FormSubmitEvent:
		buttonIndex := event.ButtonIndex
		buttonLabel := event.ButtonLabel
		return func() tcell.Event {
			return newModalDoneEvent(buttonIndex, buttonLabel)
		}
	case *FormCancelEvent:
		return func() tcell.Event {
			return newModalDoneEvent(-1, "")
		}
	case *ButtonExitEvent:
		return m.form.HandleEvent(event)
	case *MouseEvent:
		// Pass mouse events on to the form.
		cmd := m.form.HandleEvent(event)
		if cmd == nil && event.Action == MouseLeftDown && m.InRect(event.Position()) {
			cmd = SetFocus(m)
		}
		return cmd
	case *KeyEvent:
		// Keep arrow-key navigation between modal buttons.
		switch event.Key() {
		case tcell.KeyDown, tcell.KeyRight:
			event = tcell.NewEventKey(tcell.KeyTab, "", tcell.ModNone)
		case tcell.KeyUp, tcell.KeyLeft:
			event = tcell.NewEventKey(tcell.KeyBacktab, "", tcell.ModNone)
		}
		// Forward the key event to the frame so the focused form button receives Tab/Backtab and Form.finished can move focus to the next/previous button.
		if m.frame.HasFocus() {
			return m.frame.HandleEvent(event)
		}
	case *PasteEvent:
		if m.frame.HasFocus() {
			return m.frame.HandleEvent(event)
		}
	}
	return nil
}
