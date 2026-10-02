// Package textarea provides a multi-line text area that wraps on words.
package textarea

import (
	"github.com/ayn2op/tview/layout"
	"strings"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/grapheme"
	"github.com/ayn2op/tview/richtext"
	"github.com/gdamore/tcell/v3"
	"github.com/rivo/uniseg"
)

// Widget draws EditState and turns typing into Changes.
type Widget struct {
	editState     *EditState
	width, height layout.Length
	placeholder   string
	style         tcell.Style
	keybind       func(tview.KeyMsg) (Action, bool)
	onChange      func(Change) tview.Msg
	focused       bool
}

var _ tview.Element = Widget{}

// New returns a text area showing editState that fills its parent.
func New(editState *EditState) Widget {
	return Widget{
		editState: editState,
		width:     layout.Fill,
		height:    layout.Fill,
		keybind:   DefaultKeybind,
	}
}

// Width sets the width of the text area.
func (w Widget) Width(width layout.Length) Widget {
	w.width = width
	return w
}

// Height sets the height of the text area.
func (w Widget) Height(height layout.Length) Widget {
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

// Focused sets whether the text area receives keys and paste and shows the cursor.
func (w Widget) Focused(focused bool) Widget {
	w.focused = focused
	return w
}

// Size returns the width and height of the text area.
func (w Widget) Size() (width, height layout.Length) {
	return w.width, w.height
}

// Layout returns the size of the text area within limits.
func (w Widget) Layout(limits layout.Limits) layout.Size {
	width, height := w.Size()
	return layout.Atomic(limits, width, height)
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
	screen.FillArea(area.X, area.Y, area.Width, area.Height, ' ', w.style)
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

// Handle turns editing keys and paste into a Change while focused. Other messages pass through unchanged.
func (w Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	if !w.focused || w.onChange == nil {
		return msg
	}
	value, cursor := w.editState.value, w.editState.cursor
	switch m := msg.(type) {
	case tview.KeyMsg:
		action, ok := w.keybind(m)
		switch {
		case ok:
			value, cursor = edit(action, value, cursor, wrap(value, area.Width))
		case m.Key() == tcell.KeyRune && m.Modifiers()&^tcell.ModShift == 0:
			value, cursor = value[:cursor]+m.Str()+value[cursor:], cursor+len(m.Str())
		default:
			return msg
		}
	case tview.PasteMsg:
		text := strings.ReplaceAll(string(m), "\r\n", "\n")
		value, cursor = value[:cursor]+text+value[cursor:], cursor+len(text)
	default:
		return msg
	}
	cursorRow, _ := locate(value, wrap(value, area.Width), cursor)
	return w.onChange(Change{value: value, cursor: cursor, row: scroll(w.editState.row, cursorRow, area.Height)})
}

// edit applies action to value and cursor, with lines being value wrapped.
func edit(action Action, value string, cursor int, lines []line) (string, int) {
	row, column := locate(value, lines, cursor)
	switch action {
	case ActionNewline:
		return value[:cursor] + "\n" + value[cursor:], cursor + 1
	case ActionBackspace:
		start := grapheme.Previous(value, cursor)
		return value[:start] + value[cursor:], start
	case ActionDelete:
		return value[:cursor] + value[grapheme.Next(value, cursor):], cursor
	case ActionLeft:
		return value, grapheme.Previous(value, cursor)
	case ActionRight:
		return value, grapheme.Next(value, cursor)
	case ActionUp:
		if row > 0 {
			cursor = at(value, lines[row-1], column)
		}
	case ActionDown:
		if row < len(lines)-1 {
			cursor = at(value, lines[row+1], column)
		}
	case ActionHome:
		return value, lines[row].start
	case ActionEnd:
		return value, lines[row].end
	}
	return value, cursor
}
