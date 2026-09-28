// Package button provides a declarative button element.
package button

import (
	"github.com/ayn2op/tview"
	"github.com/gdamore/tcell/v3"
)

// Widget is a labeled button. It is a value built in View; the model owns any state it reflects, such as focus.
type Widget struct {
	label         string
	width, height tview.Length
	style         tcell.Style
	focusedStyle  tcell.Style
	disabledStyle tcell.Style
	keybind       func(tview.KeyMsg) (Action, bool)
	onClick       tview.Msg
	disabled      bool
	focused       bool
}

var _ tview.Element = Widget{}

// New returns an empty button that fills its parent.
func New() Widget {
	return Widget{
		width:         tview.Fill,
		height:        tview.Fill,
		keybind:       DefaultKeybind,
		focusedStyle:  tcell.StyleDefault.Reverse(true),
		disabledStyle: tcell.StyleDefault.Dim(true),
	}
}

// Label sets the text centered on the button.
func (w Widget) Label(label string) Widget {
	w.label = label
	return w
}

// Width sets the width of the button.
func (w Widget) Width(width tview.Length) Widget {
	w.width = width
	return w
}

// Height sets the height of the button.
func (w Widget) Height(height tview.Length) Widget {
	w.height = height
	return w
}

// Style sets the style of the button.
func (w Widget) Style(style tcell.Style) Widget {
	w.style = style
	return w
}

// FocusedStyle sets the style of the button when it is focused.
func (w Widget) FocusedStyle(style tcell.Style) Widget {
	w.focusedStyle = style
	return w
}

// DisabledStyle sets the style of the button when it is disabled.
func (w Widget) DisabledStyle(style tcell.Style) Widget {
	w.disabledStyle = style
	return w
}

// Keybind sets the function that turns keys into Actions, DefaultKeybind unless set.
func (w Widget) Keybind(f func(tview.KeyMsg) (Action, bool)) Widget {
	w.keybind = f
	return w
}

// OnClick sets the message the button returns when it is pressed.
func (w Widget) OnClick(msg tview.Msg) Widget {
	w.onClick = msg
	return w
}

// Disabled sets whether the button ignores input.
func (w Widget) Disabled(disabled bool) Widget {
	w.disabled = disabled
	return w
}

// Focused sets whether the button has the focus, so ActionPress presses it and it is drawn with the focused style.
func (w Widget) Focused(focused bool) Widget {
	w.focused = focused
	return w
}

// Size returns the width and height of the button.
func (w Widget) Size() (width, height tview.Length) {
	return w.width, w.height
}

// Draw fills area with the style for the button's state and centers the label in it.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	style := w.style
	switch {
	case w.disabled:
		style = w.disabledStyle
	case w.focused:
		style = w.focusedStyle
	}
	for y := area.Y; y < area.Y+area.Height; y++ {
		for x := area.X; x < area.X+area.Width; x++ {
			screen.Put(x, y, " ", style)
		}
	}
	if area.Width > 0 && area.Height > 0 {
		tview.Print(screen, w.label, area.X, area.Y+area.Height/2, area.Width, tview.AlignmentCenter, style)
	}
}

// Handle returns the OnClick message for a left click within area, or for ActionPress while focused. Other messages pass through unchanged.
func (w Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	if w.disabled || w.onClick == nil {
		return msg
	}
	switch m := msg.(type) {
	case tview.KeyMsg:
		if action, ok := w.keybind(m); w.focused && ok && action == ActionPress {
			return w.onClick
		}
	case tview.MouseMsg:
		if m.Action == tview.MouseLeftClick && area.Contains(m.Position()) {
			return w.onClick
		}
	}
	return msg
}
