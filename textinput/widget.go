// Package textinput provides a single-line text input.
package textinput

import (
	"strings"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/grapheme"
	"github.com/ayn2op/tview/keybind"
	"github.com/gdamore/tcell/v3"
	"github.com/rivo/uniseg"
)

// Widget draws EditState and turns typing into Actions.
type Widget struct {
	editState *EditState
	width     tview.Length
	style     tcell.Style
	mask      string
	keybinds  Keybinds
	onAction  func(Action) tview.Msg
	onSubmit  tview.Msg
	focused   bool
}

var _ tview.Element = Widget{}

// New returns a one-line text input showing editState that fills its parent's width.
func New(editState *EditState) Widget {
	return Widget{
		editState: editState,
		width:     tview.Fill,
		keybinds:  defaultKeybinds,
	}
}

// Width sets the width of the text input.
func (w Widget) Width(width tview.Length) Widget {
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

// Keybinds sets the keys the text input edits with.
func (w Widget) Keybinds(keybinds Keybinds) Widget {
	w.keybinds = keybinds
	return w
}

// OnAction turns typing into the message f returns for the Action, which the model applies with EditState.Perform.
func (w Widget) OnAction(f func(Action) tview.Msg) Widget {
	w.onAction = f
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
func (w Widget) Size() (width, height tview.Length) {
	return w.width, tview.Fixed(1)
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

// Handle turns editing keys and paste into an Action while focused, and the Submit keybind into the OnSubmit message. Other messages pass through unchanged.
func (w Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	if !w.focused || w.onAction == nil {
		return msg
	}
	value, cursor := w.editState.value, w.editState.cursor
	switch m := msg.(type) {
	case tview.KeyMsg:
		if w.onSubmit != nil && keybind.Matches(m, w.keybinds.Submit) {
			return w.onSubmit
		}
		var ok bool
		if value, cursor, ok = w.edit(m, value, cursor); !ok {
			return msg
		}
	case tview.PasteMsg:
		text := strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ").Replace(string(m))
		value, cursor = value[:cursor]+text+value[cursor:], cursor+len(text)
	default:
		return msg
	}
	return w.onAction(Action{value: value, cursor: cursor, offset: w.scroll(value, cursor, w.editState.offset, area.Width)})
}

// edit applies key to value and cursor, and reports whether it is an editing key.
func (w Widget) edit(key tview.KeyMsg, value string, cursor int) (string, int, bool) {
	k := w.keybinds
	switch {
	case keybind.Matches(key, k.Backspace):
		start := grapheme.Previous(value, cursor)
		return value[:start] + value[cursor:], start, true
	case keybind.Matches(key, k.Delete):
		return value[:cursor] + value[grapheme.Next(value, cursor):], cursor, true
	case keybind.Matches(key, k.Left):
		return value, grapheme.Previous(value, cursor), true
	case keybind.Matches(key, k.Right):
		return value, grapheme.Next(value, cursor), true
	case keybind.Matches(key, k.Home):
		return value, 0, true
	case keybind.Matches(key, k.End):
		return value, len(value), true
	case key.Key() == tcell.KeyRune && key.Modifiers()&^tcell.ModShift == 0:
		return value[:cursor] + key.Str() + value[cursor:], cursor + len(key.Str()), true
	}
	return value, cursor, false
}
