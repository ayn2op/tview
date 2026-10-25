// Package button provides a declarative button widget.
package button

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/layout"
)

// Widget is a button around its content, such as text. It is a value built in View; the model owns any state it reflects, such as focus, and picks its style from it.
type Widget struct {
	content       tview.Widget
	width, height layout.Length
	style         tview.Style
	keybind       func(tview.KeyMsg) Action
	onClick       tview.Msg
	disabled      bool
	focused       bool
}

var _ tview.Widget = Widget{}

// New returns a button around content, as large as it.
func New(content tview.Widget) Widget {
	return Widget{
		content: content,
		width:   layout.Shrink,
		height:  layout.Shrink,
		keybind: DefaultKeybind,
	}
}

// Width sets the width of the button.
func (w Widget) Width(width layout.Length) Widget {
	w.width = width
	return w
}

// Height sets the height of the button.
func (w Widget) Height(height layout.Length) Widget {
	w.height = height
	return w
}

// Style sets the style of the button.
func (w Widget) Style(style tview.Style) Widget {
	w.style = style
	return w
}

// Keybind sets the function that turns keys into Actions, or ActionNone for keys it does not bind, DefaultKeybind unless set.
func (w Widget) Keybind(f func(tview.KeyMsg) Action) Widget {
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

// Focused sets whether the button has the focus, so ActionPress presses it.
func (w Widget) Focused(focused bool) Widget {
	w.focused = focused
	return w
}

// Size returns the width and height of the button.
func (w Widget) Size() (width, height layout.Length) {
	return w.width, w.height
}

// Layout returns the size of the button within limits.
func (w Widget) Layout(limits layout.Limits) layout.Size {
	return layout.Sized(limits, w.width, w.height, w.content.Layout)
}

// Draw fills area with the button's style and centers the content in it, which takes from that style what it leaves unset.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	screen.FillArea(area.X, area.Y, area.Width, area.Height, ' ', w.style)
	w.content.Draw(screen, w.contentArea(area))
}

// contentArea returns where the content goes in area: in its middle, at the size of its layout.
func (w Widget) contentArea(area tview.Rectangle) tview.Rectangle {
	size := w.content.Layout(layout.Limits{Max: layout.Size{Width: area.Width, Height: area.Height}})
	area.X += (area.Width - size.Width) / 2
	area.Y += (area.Height - size.Height) / 2
	area.Width, area.Height = size.Width, size.Height
	return area
}

// Handle passes msg to the content first, and returns what the content makes of it if it is another message. Otherwise it returns the OnClick message for a left press within area, or for ActionPress while focused, and passes other messages through unchanged.
func (w Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	out := w.content.Handle(msg, w.contentArea(area))
	if w.disabled || w.onClick == nil {
		return out
	}

	// Only input is compared with out, as other messages may be of an uncomparable type.
	switch m := msg.(type) {
	case tview.KeyMsg:
		if out == msg && w.focused && w.keybind(m) == ActionPress {
			return w.onClick
		}
	case tview.MouseMsg:
		if out == msg && m.Action == tview.MouseLeftDown && area.Contains(m.Position()) {
			return w.onClick
		}
	}
	return out
}
