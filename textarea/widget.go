// Package textarea provides a multi-line text area that wraps on words.
package textarea

import (
	"strings"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/grapheme"
	"github.com/ayn2op/tview/keybind"
	"github.com/ayn2op/tview/richtext"
	"github.com/gdamore/tcell/v3"
	"github.com/rivo/uniseg"
)

// Widget draws EditState and turns typing into Actions.
type Widget struct {
	editState     *EditState
	width, height tview.Length
	placeholder   string
	style         tcell.Style
	keybinds      Keybinds
	onAction      func(Action) tview.Msg
	focused       bool
}

var _ tview.Element = Widget{}

// New returns a text area showing editState that fills its parent.
func New(editState *EditState) Widget {
	return Widget{
		editState: editState,
		width:     tview.Fill,
		height:    tview.Fill,
		keybinds:  defaultKeybinds,
	}
}

// Width sets the width of the text area.
func (w Widget) Width(width tview.Length) Widget {
	w.width = width
	return w
}

// Height sets the height of the text area.
func (w Widget) Height(height tview.Length) Widget {
	w.height = height
	return w
}

// Placeholder sets the text shown dimmed while the value is empty.
func (w Widget) Placeholder(placeholder string) Widget {
	w.placeholder = placeholder
	return w
}

// Style sets the style of the text area.
func (w Widget) Style(style tcell.Style) Widget {
	w.style = style
	return w
}

// Keybinds sets the keys the text area edits with.
func (w Widget) Keybinds(keybinds Keybinds) Widget {
	w.keybinds = keybinds
	return w
}

// OnAction turns typing into the message f returns for the Action, which the model applies with EditState.Perform.
func (w Widget) OnAction(f func(Action) tview.Msg) Widget {
	w.onAction = f
	return w
}

// Focused sets whether the text area receives keys and paste and shows the cursor.
func (w Widget) Focused(focused bool) Widget {
	w.focused = focused
	return w
}

// Size returns the width and height of the text area.
func (w Widget) Size() (width, height tview.Length) {
	return w.width, w.height
}

// line is the byte range of one wrapped line of the value.
type line struct {
	start, end int
}

// wrap splits value into lines that fit width, breaking between words where possible.
func wrap(value string, width int) []line {
	width = max(width, 1)
	var lines []line
	start := 0
	for _, logical := range strings.Split(value, "\n") {
		pos := start
		for _, piece := range richtext.WordWrap(logical, width) {
			lines = append(lines, line{start: pos, end: pos + len(piece)})
			pos += len(piece)
		}
		start += len(logical) + 1
	}
	return lines
}

// locate returns the wrapped line holding cursor and the cursor's column in it.
func locate(value string, lines []line, cursor int) (row, column int) {
	for i, l := range lines {
		if l.start <= cursor {
			row = i
		}
	}
	return row, uniseg.StringWidth(value[lines[row].start:cursor])
}

// at returns the offset in l closest to column.
func at(value string, l line, column int) int {
	offset := l.start
	for offset < l.end {
		end := grapheme.Next(value, offset)
		if column -= uniseg.StringWidth(value[offset:end]); column < 0 {
			break
		}
		offset = end
	}
	return offset
}

// scroll returns the first visible line that keeps the cursor row visible in height lines.
func scroll(first, cursorRow, height int) int {
	return max(min(first, cursorRow), cursorRow-height+1, 0)
}

// Draw fills area with the style and draws the visible lines, or the placeholder while the value is empty, showing the cursor while focused.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	for y := area.Y; y < area.Y+area.Height; y++ {
		for x := area.X; x < area.X+area.Width; x++ {
			screen.Put(x, y, " ", w.style)
		}
	}
	if area.Width <= 0 || area.Height <= 0 {
		return
	}
	c := w.editState
	if c.value == "" && w.placeholder != "" {
		tview.Print(screen, w.placeholder, area.X, area.Y, area.Width, tview.AlignmentLeft, w.style.Dim(true))
	}
	lines := wrap(c.value, area.Width)
	cursorRow, column := locate(c.value, lines, c.cursor)
	first := scroll(c.row, cursorRow, area.Height)
	for i := 0; i < area.Height && first+i < len(lines); i++ {
		l := lines[first+i]
		tview.Print(screen, c.value[l.start:l.end], area.X, area.Y+i, area.Width, tview.AlignmentLeft, w.style)
	}
	if w.focused {
		screen.ShowCursor(area.X+column, area.Y+cursorRow-first)
	}
}

// Handle turns editing keys and paste into an Action while focused. Other messages pass through unchanged.
func (w Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	if !w.focused || w.onAction == nil {
		return msg
	}
	value, cursor := w.editState.value, w.editState.cursor
	switch m := msg.(type) {
	case tview.KeyMsg:
		var ok bool
		if value, cursor, ok = w.edit(m, value, cursor, wrap(value, area.Width)); !ok {
			return msg
		}
	case tview.PasteMsg:
		text := strings.ReplaceAll(string(m), "\r\n", "\n")
		value, cursor = value[:cursor]+text+value[cursor:], cursor+len(text)
	default:
		return msg
	}
	cursorRow, _ := locate(value, wrap(value, area.Width), cursor)
	return w.onAction(Action{value: value, cursor: cursor, row: scroll(w.editState.row, cursorRow, area.Height)})
}

// edit applies key to value and cursor, with lines being value wrapped, and reports whether it is an editing key.
func (w Widget) edit(key tview.KeyMsg, value string, cursor int, lines []line) (string, int, bool) {
	k := w.keybinds
	row, column := locate(value, lines, cursor)
	switch {
	case keybind.Matches(key, k.Newline):
		return value[:cursor] + "\n" + value[cursor:], cursor + 1, true
	case keybind.Matches(key, k.Backspace):
		start := grapheme.Previous(value, cursor)
		return value[:start] + value[cursor:], start, true
	case keybind.Matches(key, k.Delete):
		return value[:cursor] + value[grapheme.Next(value, cursor):], cursor, true
	case keybind.Matches(key, k.Left):
		return value, grapheme.Previous(value, cursor), true
	case keybind.Matches(key, k.Right):
		return value, grapheme.Next(value, cursor), true
	case keybind.Matches(key, k.Up):
		if row > 0 {
			cursor = at(value, lines[row-1], column)
		}
		return value, cursor, true
	case keybind.Matches(key, k.Down):
		if row < len(lines)-1 {
			cursor = at(value, lines[row+1], column)
		}
		return value, cursor, true
	case keybind.Matches(key, k.Home):
		return value, lines[row].start, true
	case keybind.Matches(key, k.End):
		return value, lines[row].end, true
	case key.Key() == tcell.KeyRune && key.Modifiers()&^tcell.ModShift == 0:
		return value[:cursor] + key.Str() + value[cursor:], cursor + len(key.Str()), true
	}
	return value, cursor, false
}
