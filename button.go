package tview

import (
	"github.com/gdamore/tcell/v3"
)

type ButtonSelectedEvent struct {
	tcell.EventTime
	Label string
}

func newButtonSelectedEvent(label string) *ButtonSelectedEvent {
	event := &ButtonSelectedEvent{Label: label}
	event.SetEventNow()
	return event
}

type ButtonExitEvent struct {
	tcell.EventTime
	tcell.Key
}

func newButtonExitEvent(key tcell.Key) *ButtonExitEvent {
	event := &ButtonExitEvent{Key: key}
	event.SetEventNow()
	return event
}

// Button is labeled box that triggers an action when selected.
//
// See https://github.com/ayn2op/tview/wiki/Button for an example.
type Button struct {
	*Box
	// If set to true, the button cannot be activated.
	disabled bool
	// The text to be displayed inside the button.
	text string
	// The button's style (when deactivated).
	style tcell.Style
	// The button's style (when activated).
	activatedStyle tcell.Style
	// The button's style (when disabled).
	disabledStyle tcell.Style
}

// NewButton returns a new input field.
func NewButton(label string) *Button {
	box := NewBox()
	box.SetRect(0, 0, TaggedStringWidth(label)+4, 1)
	return &Button{
		Box:            box,
		text:           label,
		style:          tcell.StyleDefault.Background(Styles.ContrastBackgroundColor).Foreground(Styles.PrimaryTextColor),
		activatedStyle: tcell.StyleDefault.Background(Styles.PrimaryTextColor).Foreground(Styles.InverseTextColor),
		disabledStyle:  tcell.StyleDefault.Background(Styles.ContrastBackgroundColor).Foreground(Styles.ContrastSecondaryTextColor),
	}
}

// SetLabel sets the button text.
func (b *Button) SetLabel(label string) *Button {
	if b.text != label {
		b.text = label
	}
	return b
}

// GetLabel returns the button text.
func (b *Button) GetLabel() string {
	return b.text
}

// SetLabelColor sets the color of the button text.
func (b *Button) SetLabelColor(color tcell.Color) *Button {
	style := b.style.Foreground(color)
	if b.style != style {
		b.style = style
	}
	return b
}

// SetStyle sets the style of the button used when it is not focused.
func (b *Button) SetStyle(style tcell.Style) *Button {
	if b.style != style {
		b.style = style
	}
	return b
}

// SetLabelColorActivated sets the color of the button text when the button is
// in focus.
func (b *Button) SetLabelColorActivated(color tcell.Color) *Button {
	style := b.activatedStyle.Foreground(color)
	if b.activatedStyle != style {
		b.activatedStyle = style
	}
	return b
}

// SetBackgroundColorActivated sets the background color of the button text when
// the button is in focus.
func (b *Button) SetBackgroundColorActivated(color tcell.Color) *Button {
	style := b.activatedStyle.Background(color)
	if b.activatedStyle != style {
		b.activatedStyle = style
	}
	return b
}

// SetActivatedStyle sets the style of the button used when it is focused.
func (b *Button) SetActivatedStyle(style tcell.Style) *Button {
	if b.activatedStyle != style {
		b.activatedStyle = style
	}
	return b
}

// SetDisabledStyle sets the style of the button used when it is disabled.
func (b *Button) SetDisabledStyle(style tcell.Style) *Button {
	if b.disabledStyle != style {
		b.disabledStyle = style
	}
	return b
}

// SetDisabled sets whether or not the button is disabled. Disabled buttons
// cannot be activated.
//
// If the button is part of a form, you should set focus to the form itself
// after calling this function to set focus to the next non-disabled form item.
func (b *Button) SetDisabled(disabled bool) *Button {
	if b.disabled != disabled {
		b.disabled = disabled
	}
	return b
}

// GetDisabled returns whether or not the button is disabled.
func (b *Button) GetDisabled() bool {
	return b.disabled
}

// Draw draws this primitive onto the screen.
func (b *Button) Draw(screen tcell.Screen) {
	// Draw the box.
	style := b.style
	if b.disabled {
		style = b.disabledStyle
	}
	if b.HasFocus() && !b.disabled {
		style = b.activatedStyle
	}
	backgroundColor := style.GetBackground()
	b.SetBackgroundColor(backgroundColor)
	b.DrawForSubclass(screen, b)

	// Draw label.
	x, y, width, height := b.GetInnerRect()
	if width > 0 && height > 0 {
		y = y + height/2
		printWithStyle(screen, b.text, x, y, 0, width, AlignmentCenter, style, true)
	}
}

// HandleEvent handles input events for this primitive.
func (b *Button) HandleEvent(event tcell.Event) Command {
	if b.disabled {
		return nil
	}

	switch event := event.(type) {
	case *KeyEvent:
		// Process key event.
		switch key := event.Key(); key {
		case tcell.KeyEnter: // Selected.
			label := b.GetLabel()
			return EventCommand(func() tcell.Event {
				return newButtonSelectedEvent(label)
			})
		case tcell.KeyBacktab, tcell.KeyTab, tcell.KeyEscape: // Leave. No action.
			exitKey := key
			return EventCommand(func() tcell.Event {
				return newButtonExitEvent(exitKey)
			})
		}
		return RedrawCommand{}
	case *MouseEvent:
		if !b.InRect(event.Position()) {
			return nil
		}

		// Process mouse event.
		switch event.Action {
		case MouseLeftDown:
			return SetFocus(b)
		case MouseLeftClick:
			label := b.GetLabel()
			return EventCommand(func() tcell.Event {
				return newButtonSelectedEvent(label)
			})
		}
	}
	return nil
}
