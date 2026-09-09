package tview

import (
	"github.com/gdamore/tcell/v3"
	"github.com/rivo/uniseg"
)

type ButtonSelectedMsg struct {
	Label string
}

// Button is labeled box that triggers an action when selected.
type Button struct {
	*Box
	// If set to true, the button cannot be activated.
	disabled bool
	// The text to be displayed inside the button.
	text string
	// The button's style (when deactivated).
	style tcell.Style
	// The button's style (when disabled).
	disabledStyle tcell.Style
}

// NewButton returns a new input field.
func NewButton(label string) *Button {
	box := NewBox()
	box.SetRect(0, 0, uniseg.StringWidth(label)+4, 1)
	return &Button{
		Box:           box,
		text:          label,
		style:         tcell.StyleDefault.Background(Styles.ContrastBackgroundColor).Foreground(Styles.PrimaryTextColor),
		disabledStyle: tcell.StyleDefault.Background(Styles.ContrastBackgroundColor).Foreground(Styles.ContrastSecondaryTextColor),
	}
}

// Label returns the button text.
func (b *Button) Label() string {
	return b.text
}

// SetLabel sets the button text.
func (b *Button) SetLabel(label string) *Button {
	b.text = label
	return b
}

// SetLabelColor sets the color of the button text.
func (b *Button) SetLabelColor(color tcell.Color) *Button {
	style := b.style.Foreground(color)
	b.style = style
	return b
}

// SetStyle sets the style of the button used when it is not focused.
func (b *Button) SetStyle(style tcell.Style) *Button {
	b.style = style
	return b
}

// SetDisabledStyle sets the style of the button used when it is disabled.
func (b *Button) SetDisabledStyle(style tcell.Style) *Button {
	b.disabledStyle = style
	return b
}

// Disabled returns whether or not the button is disabled.
func (b *Button) Disabled() bool {
	return b.disabled
}

// SetDisabled sets whether or not the button is disabled. Disabled buttons
// cannot be activated.
//
// If the button is part of a form, you should set focus to the form itself
// after calling this function to set focus to the next non-disabled form item.
func (b *Button) SetDisabled(disabled bool) *Button {
	b.disabled = disabled
	return b
}

// View draws this model onto the screen.
func (b *Button) View(screen Screen) {
	// Draw the box.
	style := b.style
	if b.disabled {
		style = b.disabledStyle
	}
	backgroundColor := style.GetBackground()
	b.SetBackgroundColor(backgroundColor)
	b.Box.View(screen)

	// Draw label.
	x, y, width, height := b.InnerRect()
	if width > 0 && height > 0 {
		y = y + height/2
		PrintStyled(screen, b.text, x, y, 0, width, AlignmentCenter, style, true)
	}
}

// Update handles input events for this model.
func (b *Button) Update(msg Msg) Cmd {
	if b.disabled {
		return b.Box.Update(msg)
	}

	switch msg := msg.(type) {
	case KeyMsg:
		// Process key event.
		switch key := msg.Key(); key {
		case tcell.KeyEnter:
			label := b.Label()
			return func() Msg { return ButtonSelectedMsg{Label: label} }
		}
		return nil
	case MouseMsg:
		if !b.InRect(msg.Position()) {
			return nil
		}

		switch msg.Action {
		case MouseLeftDown:
			return nil
		case MouseLeftClick:
			label := b.Label()
			return func() Msg { return ButtonSelectedMsg{Label: label} }
		}
	}
	return b.Box.Update(msg)
}
