// Package dialog shows a message with buttons, centered over the rest of the screen.
package dialog

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/box"
	"github.com/ayn2op/tview/button"
	"github.com/ayn2op/tview/center"
	"github.com/ayn2op/tview/column"
	"github.com/ayn2op/tview/layout"
	"github.com/ayn2op/tview/row"
	"github.com/ayn2op/tview/text"
	"github.com/gdamore/tcell/v3"
	"github.com/rockorager/go-uucode"
)

// Widget is a message with buttons. ActionNext and ActionPrevious move the focus between the buttons, ActionPress or a click presses one, and ActionCancel cancels.
type Widget struct {
	text                  string
	buttons               []string
	focus                 int
	background, textColor tcell.Color
	buttonStyle           tview.Style
	activatedStyle        tview.Style
	keybind               func(tview.KeyMsg) Action
	onFocus               func(index int) tview.Msg
	onDone                func(index int, label string) tview.Msg
}

var _ tview.Widget = Widget{}

// New returns an empty dialog.
func New() Widget {
	return Widget{
		activatedStyle: tcell.StyleDefault.Reverse(true),
		keybind:        DefaultKeybind,
	}
}

// Text sets the message, which is wrapped to fit.
func (w Widget) Text(text string) Widget {
	w.text = text
	return w
}

// Buttons sets the labels of the buttons below the message.
func (w Widget) Buttons(labels ...string) Widget {
	w.buttons = labels
	return w
}

// Focus sets the index of the focused button.
func (w Widget) Focus(index int) Widget {
	w.focus = index
	return w
}

// Keybind sets the function that turns keys into Actions, or ActionNone for keys it does not bind, DefaultKeybind unless set.
func (w Widget) Keybind(f func(tview.KeyMsg) Action) Widget {
	w.keybind = f
	return w
}

// Background sets the background color.
func (w Widget) Background(color tcell.Color) Widget {
	w.background = color
	return w
}

// TextColor sets the color of the message.
func (w Widget) TextColor(color tcell.Color) Widget {
	w.textColor = color
	return w
}

// ButtonStyle sets the style of the buttons that are not focused.
func (w Widget) ButtonStyle(style tview.Style) Widget {
	w.buttonStyle = style
	return w
}

// ActivatedStyle sets the style of the focused button.
func (w Widget) ActivatedStyle(style tview.Style) Widget {
	w.activatedStyle = style
	return w
}

// OnFocus sets the message f returns for the button the focus moves to.
func (w Widget) OnFocus(f func(index int) tview.Msg) Widget {
	w.onFocus = f
	return w
}

// OnDone sets the message f returns for the pressed button, or for index -1 and an empty label when the dialog is canceled.
func (w Widget) OnDone(f func(index int, label string) tview.Msg) Widget {
	w.onDone = f
	return w
}

// contentWidth is how wide the message is at most, unless the buttons are wider.
const contentWidth = 80

// view returns the dialog in the middle of its area: a box as tall as the message and the buttons below it.
func (w Widget) view() tview.Widget {
	buttons := row.New().Width(layout.Shrink).Height(layout.Shrink).Spacing(2)
	for i, label := range w.buttons {
		b := button.New().
			Label(label).
			Style(w.buttonStyle).
			FocusedStyle(w.activatedStyle).
			Width(layout.Fixed(uucode.StringWidth(label) + 4)).
			Keybind(noKeys).
			Focused(i == w.focus)
		if w.onDone != nil {
			b = b.OnClick(w.onDone(i, label))
		}
		buttons = buttons.Push(b)
	}

	style := tcell.StyleDefault.Background(w.background).Foreground(w.textColor)
	message := text.New(w.text).
		Height(layout.Fill).
		AlignX(tview.HorizontalAlignmentCenter).
		Wrapping(text.WrappingWordOrGlyph).
		Style(style)
	// The message fills what the buttons leave, so they stay in view when it does not fit.
	content := column.New(message, buttons).Spacing(1).AlignX(tview.HorizontalAlignmentCenter).Height(layout.Shrink)
	width := max(contentWidth, buttons.Layout(layout.Limits{Infinite: layout.Axes{Width: true, Height: true}}).Width)
	return center.New(box.New(content).
		Borders(tview.BordersAll).
		Background(w.background).
		Padding(1, 1, 1, 1).
		Width(layout.Fixed(width + 4)).
		Height(layout.Shrink))
}

// Size returns Fill, as the dialog takes its whole area.
func (Widget) Size() (width, height layout.Length) {
	return layout.Fill, layout.Fill
}

// Layout returns the size of limits, as the dialog takes its whole area.
func (Widget) Layout(limits layout.Limits) layout.Size {
	return layout.Atomic(limits, layout.Fill, layout.Fill)
}

// Draw draws the dialog in the middle of area.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	w.view().Draw(screen, area)
}

// Handle turns ActionNext and ActionPrevious into the OnFocus message, ActionCancel into the OnDone message for canceling, and ActionPress or a click on a button into its OnDone message. Other messages pass through unchanged.
func (w Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	if key, ok := msg.(tview.KeyMsg); ok && len(w.buttons) > 0 {
		action := w.keybind(key)
		switch {
		case action == ActionNext && w.onFocus != nil:
			return w.onFocus((w.focus + 1) % len(w.buttons))
		case action == ActionPrevious && w.onFocus != nil:
			return w.onFocus((w.focus + len(w.buttons) - 1) % len(w.buttons))
		case action == ActionPress && w.onDone != nil && w.focus >= 0 && w.focus < len(w.buttons):
			return w.onDone(w.focus, w.buttons[w.focus])
		case action == ActionCancel && w.onDone != nil:
			return w.onDone(-1, "")
		}
	}
	return w.view().Handle(msg, area)
}
