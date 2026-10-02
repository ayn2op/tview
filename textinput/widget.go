// Package textinput provides a single-line text input.
package textinput

import (
	"github.com/ayn2op/tview/layout"
	"strings"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/grapheme"
	"github.com/gdamore/tcell/v3"
	"github.com/rivo/uniseg"
)

// Widget draws EditState and turns typing into Changes.
type Widget struct {
	editState *EditState
	width     layout.Length
	style     tcell.Style
	mask      string
	keybind   func(tview.KeyMsg) (Action, bool)
	onChange  func(Change) tview.Msg
	onSubmit  tview.Msg
	focused   bool
}

var _ tview.Element = Widget{}

// New returns a one-line text input showing editState that fills its parent's width.
func New(editState *EditState) Widget {
	return Widget{
		editState: editState,
		width:     layout.Fill,
		keybind:   DefaultKeybind,
	}
}

// Width sets the width of the text input.
func (w Widget) Width(width layout.Length) Widget {
	w.width = width
	return w
}

// Style sets the style of the text input.
func (w Widget) Style(style tcell.Style) Widget {
	w.style = style
	return w
}

// Mask sets the text drawn for each character of the value instead, such as "*" for passwords. An empty mask shows the value.
func (w Widget) Mask(mask string) Widget {
	w.mask = mask
	return w
}

// Keybind sets the function that turns keys into Actions, DefaultKeybind unless set.
func (w Widget) Keybind(f func(tview.KeyMsg) (Action, bool)) Widget {
	w.keybind = f
	return w
}

// OnChange turns typing into the message f returns for the Change, which the model applies with EditState.Apply.
func (w Widget) OnChange(f func(Change) tview.Msg) Widget {
	w.onChange = f
	return w
}

// OnSubmit sets the message returned for Enter.
func (w Widget) OnSubmit(msg tview.Msg) Widget {
	w.onSubmit = msg
	return w
}

// Focused sets whether the text input receives keys and paste and shows the cursor.
func (w Widget) Focused(focused bool) Widget {
	w.focused = focused
	return w
}

// Size returns the width and a height of one line.
func (w Widget) Size() (width, height layout.Length) {
	return w.width, layout.Fixed(1)
}

// Layout returns the size of the text input within limits.
func (w Widget) Layout(limits layout.Limits) layout.Size {
	width, height := w.Size()
	return layout.Atomic(limits, width, height)
}

// shown returns text as it is drawn, masked if a mask is set.
func (w Widget) shown(text string) string {
	if w.mask == "" {
		return text
	}
	return strings.Repeat(w.mask, uniseg.GraphemeClusterCount(text))
}

// skip returns text without the clusters in its first n cells.
func skip(text string, n int) string {
	state := -1
	for n > 0 && text != "" {
		var w int
		_, text, w, state = uniseg.FirstGraphemeClusterInString(text, state)
		n -= w
	}
	return text
}

// scroll returns the scroll offset that keeps the cursor visible in width cells.
func (w Widget) scroll(value string, cursor, offset, width int) int {
	column := uniseg.StringWidth(w.shown(value[:cursor]))
	return max(min(offset, column), column-width+1, 0)
}

// Draw fills the first row of area with the style and draws the value, showing the cursor while focused.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	if area.Width <= 0 || area.Height <= 0 {
		return
	}
	for x := area.X; x < area.X+area.Width; x++ {
		screen.Put(x, area.Y, " ", w.style)
	}
	c := w.editState
	offset := w.scroll(c.value, c.cursor, c.offset, area.Width)
	tview.Print(screen, skip(w.shown(c.value), offset), area.X, area.Y, area.Width, tview.AlignmentLeft, w.style)
	if w.focused {
		screen.ShowCursor(area.X+uniseg.StringWidth(w.shown(c.value[:c.cursor]))-offset, area.Y)
	}
}

// Handle turns editing keys and paste into a Change while focused, and ActionSubmit into the OnSubmit message. Other messages pass through unchanged.
func (w Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	if !w.focused || w.onChange == nil {
		return msg
	}
	value, cursor := w.editState.value, w.editState.cursor
	switch m := msg.(type) {
	case tview.KeyMsg:
		action, ok := w.keybind(m)
		switch {
		case ok && action == ActionSubmit:
			if w.onSubmit == nil {
				return msg
			}
			return w.onSubmit
		case ok:
			value, cursor = edit(action, value, cursor)
		case m.Key() == tcell.KeyRune && m.Modifiers()&^tcell.ModShift == 0:
			value, cursor = value[:cursor]+m.Str()+value[cursor:], cursor+len(m.Str())
		default:
			return msg
		}
	case tview.PasteMsg:
		text := strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ").Replace(string(m))
		value, cursor = value[:cursor]+text+value[cursor:], cursor+len(text)
	default:
		return msg
	}
	return w.onChange(Change{value: value, cursor: cursor, offset: w.scroll(value, cursor, w.editState.offset, area.Width)})
}

// edit applies action to value and cursor.
func edit(action Action, value string, cursor int) (string, int) {
	switch action {
	case ActionBackspace:
		start := grapheme.Previous(value, cursor)
		return value[:start] + value[cursor:], start
	case ActionDelete:
		return value[:cursor] + value[grapheme.Next(value, cursor):], cursor
	case ActionLeft:
		return value, grapheme.Previous(value, cursor)
	case ActionRight:
		return value, grapheme.Next(value, cursor)
	case ActionHome:
		return value, 0
	case ActionEnd:
		return value, len(value)
	}
	return value, cursor
}
