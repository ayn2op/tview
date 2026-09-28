// Package dialog shows a message with buttons, centered over the rest of the screen.
package dialog

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/box"
	"github.com/ayn2op/tview/button"
	"github.com/ayn2op/tview/center"
	"github.com/ayn2op/tview/column"
	"github.com/ayn2op/tview/richtext"
	"github.com/ayn2op/tview/row"
	"github.com/gdamore/tcell/v3"
	"github.com/rivo/uniseg"
)

// Widget is a message with buttons. ActionNext and ActionPrevious move the focus between the buttons, ActionPress or a click presses one, and ActionCancel cancels.
type Widget struct {
	text                  string
	buttons               []string
	focus                 int
	background, textColor tcell.Color
	buttonStyle           tcell.Style
	activatedStyle        tcell.Style
	keybind               func(tview.KeyMsg) (Action, bool)
	onFocus               func(index int) tview.Msg
	onDone                func(index int, label string) tview.Msg
}

var _ tview.Element = Widget{}

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

// Keybind sets the function that turns keys into Actions, DefaultKeybind unless set.
func (w Widget) Keybind(f func(tview.KeyMsg) (Action, bool)) Widget {
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
func (w Widget) ButtonStyle(style tcell.Style) Widget {
	w.buttonStyle = style
	return w
}

// ActivatedStyle sets the style of the focused button.
func (w Widget) ActivatedStyle(style tcell.Style) Widget {
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

// layout returns the dialog's box and the area it takes within area.
func (w Widget) layout(area tview.Rectangle) (tview.Element, tview.Rectangle) {
	maxContentWidth := max(area.Width-4, 1)
	buttonsWidth := 0
	for _, label := range w.buttons {
		buttonsWidth += uniseg.StringWidth(label) + 6
	}
	contentWidth := min(max(min(80, maxContentWidth), buttonsWidth-2), maxContentWidth)
	lines := richtext.WordWrap(w.text, contentWidth)
	lines = lines[:min(len(lines), max(area.Height-6, 0))]

	buttons := row.New().Width(tview.Shrink).Height(tview.Shrink).Spacing(2)
	for i, label := range w.buttons {
		b := button.New().
			Label(label).
			Style(w.buttonStyle).
			FocusedStyle(w.activatedStyle).
			Width(tview.Fixed(uniseg.StringWidth(label) + 4)).
			Height(tview.Fixed(1)).
			Keybind(noKeys).
			Focused(i == w.focus)
		if w.onDone != nil {
			b = b.OnClick(w.onDone(i, label))
		}
		buttons = buttons.Push(b)
	}
	content := column.New(
		column.New(text{lines: lines, style: tcell.StyleDefault.Background(w.background).Foreground(w.textColor)}).Height(tview.Fixed(len(lines))),
		column.New().Height(tview.Fixed(1)),
		center.New(buttons),
	)
	dialog := box.New(content).Borders(tview.BordersAll).Background(w.background).Padding(1, 1, 1, 1)

	width, height := min(contentWidth+4, area.Width), min(len(lines)+6, area.Height)
	return dialog, tview.Rectangle{X: area.X + (area.Width-width)/2, Y: area.Y + (area.Height-height)/2, Width: width, Height: height}
}

// Draw draws the dialog in the middle of area.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	if area.Width <= 0 || area.Height <= 0 {
		return
	}
	dialog, rect := w.layout(area)
	dialog.Draw(screen, rect)
}

// Handle turns ActionNext and ActionPrevious into the OnFocus message, ActionCancel into the OnDone message for canceling, and ActionPress or a click on a button into its OnDone message. Other messages pass through unchanged.
func (w Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	if key, ok := msg.(tview.KeyMsg); ok && len(w.buttons) > 0 {
		action, _ := w.keybind(key)
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
	dialog, rect := w.layout(area)
	return dialog.Handle(msg, rect)
}

// text draws lines centered in the width of its area.
type text struct {
	lines []string
	style tcell.Style
}

func (t text) Draw(screen tview.Screen, area tview.Rectangle) {
	for i, line := range t.lines {
		if i < area.Height {
			tview.Print(screen, line, area.X, area.Y+i, area.Width, tview.AlignmentCenter, t.style)
		}
	}
}

func (text) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg { return msg }
