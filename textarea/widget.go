// Package textarea provides a multi-line text area that wraps on words.
package textarea

import (
	"strings"

	"github.com/ayn2op/tview/layout"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/grapheme"
	"github.com/ayn2op/tview/richtext"
	"github.com/gdamore/tcell/v3"
	"github.com/rivo/uniseg"
)

// Widget draws a text with a cursor in it and turns typing into the text and cursor it leaves.
type Widget struct {
	text string
	// cursor is the byte offset of the cursor in text, always at a grapheme boundary.
	cursor int

	width, height layout.Length
	placeholder   string
	style         tview.Style

	keybind func(tview.KeyMsg) Action
	focused bool

	onInput func(text string, cursor int) tview.Msg
}

var _ tview.Widget = Widget{}

// New returns a text area showing text, with the cursor at the byte offset cursor in it, that fills its parent. The model owns both.
func New(text string, cursor int) Widget {
	return Widget{
		text:    text,
		cursor:  min(max(cursor, 0), len(text)),
		width:   layout.Fill,
		height:  layout.Fill,
		keybind: DefaultKeybind,
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
func (w Widget) Style(style tview.Style) Widget {
	w.style = style
	return w
}

// Keybind sets the function that turns keys into Actions, or ActionNone for keys it does not bind, DefaultKeybind unless set.
func (w Widget) Keybind(f func(tview.KeyMsg) Action) Widget {
	w.keybind = f
	return w
}

// OnInput turns typing and moving the cursor into the message f returns for the text and cursor they leave.
func (w Widget) OnInput(f func(text string, cursor int) tview.Msg) Widget {
	w.onInput = f
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

// Layout returns the size of the text area within limits, as tall as its wrapped lines where it shrinks.
func (w Widget) Layout(limits layout.Limits) layout.Size {
	return layout.Sized(limits, w.width, w.height, func(limits layout.Limits) layout.Size {
		return layout.Size{Height: max(len(wrap(w.text, limits.Max.Width)), 1)}
	})
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

// Draw fills area with the style and draws the visible lines, or the placeholder while the value is empty, showing the cursor while focused.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	screen.FillArea(area.X, area.Y, area.Width, area.Height, ' ', w.style)
	if area.Width <= 0 || area.Height <= 0 {
		return
	}
	if w.text == "" && w.placeholder != "" {
		tview.Print(screen, w.placeholder, area.X, area.Y, area.Width, tview.AlignmentLeft, w.style.Dim(true))
	}
	lines := wrap(w.text, area.Width)
	cursorRow, column := locate(w.text, lines, w.cursor)
	// The first visible line is the one that keeps the cursor row visible.
	first := max(cursorRow-area.Height+1, 0)
	for i := 0; i < area.Height && first+i < len(lines); i++ {
		l := lines[first+i]
		tview.Print(screen, w.text[l.start:l.end], area.X, area.Y+i, area.Width, tview.AlignmentLeft, w.style)
	}
	if w.focused {
		screen.ShowCursor(area.X+column, area.Y+cursorRow-first)
	}
}

// Handle turns editing keys and paste into the OnInput message while focused. Other messages pass through unchanged.
func (w Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	if !w.focused || w.onInput == nil {
		return msg
	}
	value, cursor := w.text, w.cursor
	switch m := msg.(type) {
	case tview.KeyMsg:
		action := w.keybind(m)
		switch {
		case action != ActionNone:
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
	return w.onInput(value, cursor)
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
