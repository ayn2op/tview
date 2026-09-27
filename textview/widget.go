// Package textview displays read-only styled text that can be scrolled.
package textview

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/keybind"
	"github.com/ayn2op/tview/richtext"
	"github.com/gdamore/tcell/v3"
	"github.com/rivo/uniseg"
)

// Widget draws text and, with a ScrollState, turns scroll input into Actions.
type Widget struct {
	text          richtext.Text
	scrollState   *ScrollState
	width, height tview.Length
	wrap          bool
	wordWrap      bool
	alignment     tview.Alignment
	style         tcell.Style
	keybinds      Keybinds
	onAction      func(Action) tview.Msg
	focused       bool
}

var _ tview.Element = Widget{}

// New returns a text view of text that wraps on words and fills its parent. It scrolls only once ScrollState and OnAction are set.
func New(text richtext.Text) Widget {
	return Widget{
		text:      text,
		width:     tview.Fill,
		height:    tview.Fill,
		wrap:      true,
		wordWrap:  true,
		alignment: tview.AlignmentLeft,
		keybinds:  defaultKeybinds,
	}
}

// Width sets the width of the text view.
func (w Widget) Width(width tview.Length) Widget {
	w.width = width
	return w
}

// Height sets the height of the text view.
func (w Widget) Height(height tview.Length) Widget {
	w.height = height
	return w
}

// Wrap sets whether lines longer than the width continue on the next line. Otherwise they are cut off and can be scrolled horizontally.
func (w Widget) Wrap(wrap bool) Widget {
	w.wrap = wrap
	return w
}

// WordWrap sets whether wrapped lines break between words.
func (w Widget) WordWrap(wordWrap bool) Widget {
	w.wordWrap = wordWrap
	return w
}

// Alignment sets the horizontal alignment of each line.
func (w Widget) Alignment(alignment tview.Alignment) Widget {
	w.alignment = alignment
	return w
}

// Style sets the style of the background and the text beneath the text's own styles.
func (w Widget) Style(style tcell.Style) Widget {
	w.style = style
	return w
}

// Keybinds sets the keys the text view scrolls with.
func (w Widget) Keybinds(keybinds Keybinds) Widget {
	w.keybinds = keybinds
	return w
}

// ScrollState sets how far the text view is scrolled.
func (w Widget) ScrollState(scrollState *ScrollState) Widget {
	w.scrollState = scrollState
	return w
}

// OnAction makes the text view scrollable, turning scroll input into the message f returns for the Action, which the model applies with ScrollState.Perform.
func (w Widget) OnAction(f func(Action) tview.Msg) Widget {
	w.onAction = f
	return w
}

// Focused sets whether the text view scrolls with keys.
func (w Widget) Focused(focused bool) Widget {
	w.focused = focused
	return w
}

// Size returns the width and height of the text view.
func (w Widget) Size() (width, height tview.Length) {
	return w.width, w.height
}

// Rows returns the number of lines the text takes at width.
func (w Widget) Rows(width int) int {
	return len(w.layout(width).lines)
}

// layout is the text wrapped to a width.
type layout struct {
	lines   richtext.Text
	longest int
}

func (w Widget) layout(width int) layout {
	var l layout
	for _, line := range w.text {
		wrapped := richtext.Text{line}
		switch {
		case w.wrap && w.wordWrap:
			wrapped = richtext.WrapWords(line, width)
		case w.wrap:
			wrapped = richtext.Wrap(line, width)
		}
		for _, line := range wrapped {
			l.lines = append(l.lines, line)
			l.longest = max(l.longest, line.Width())
		}
	}
	return l
}

// origin returns where, in a canvas as wide as the longest line or area, the view starts for a column offset of 0.
func (w Widget) origin(l layout, width int) int {
	switch w.alignment {
	case tview.AlignmentCenter:
		return max(l.longest-width, 0) / 2
	case tview.AlignmentRight:
		return max(l.longest-width, 0)
	default:
		return 0
	}
}

// clamp returns the scroll position limited to what can be shown in area, with the last lines shown if followEnd is set.
func (w Widget) clamp(l layout, area tview.Rectangle, row, column int, followEnd bool) (int, int) {
	if followEnd {
		row = len(l.lines) - area.Height
	}
	row = min(max(row, 0), max(len(l.lines)-area.Height, 0))
	origin := w.origin(l, area.Width)
	column = min(max(origin+column, 0), max(l.longest-area.Width, 0)) - origin
	return row, column
}

// Draw fills area with the style and draws the visible lines.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	for y := area.Y; y < area.Y+area.Height; y++ {
		for x := area.X; x < area.X+area.Width; x++ {
			screen.Put(x, y, " ", w.style)
		}
	}
	if area.Width <= 0 || area.Height <= 0 {
		return
	}

	l := w.layout(area.Width)
	scroll := w.scroll()
	row, column := w.clamp(l, area, scroll.row, scroll.column, scroll.followEnd)
	canvas := max(l.longest, area.Width)
	viewStart := w.origin(l, area.Width) + column
	for i := 0; i < area.Height && row+i < len(l.lines); i++ {
		line := l.lines[row+i]
		start := 0
		switch w.alignment {
		case tview.AlignmentCenter:
			start = (canvas - line.Width()) / 2
		case tview.AlignmentRight:
			start = canvas - line.Width()
		}
		w.drawLine(screen, line, area.X+start-viewStart, area.Y+i, area)
	}
}

// drawLine draws line starting at x, keeping only the cells within area.
func (w Widget) drawLine(screen tview.Screen, line richtext.Line, x, y int, area tview.Rectangle) {
	for _, segment := range line {
		style := tview.MergeStyle(w.style, segment.Style)
		str, state := segment.Text, -1
		for len(str) > 0 {
			var cluster string
			var width int
			cluster, str, width, state = uniseg.FirstGraphemeClusterInString(str, state)
			if cluster == "\t" {
				cluster = " "
			}
			width = max(width, 1)
			if x >= area.X && x+width <= area.X+area.Width {
				screen.Put(x, y, cluster, style)
			}
			x += width
		}
	}
}

// scroll returns the scroll state, or the top if none is set.
func (w Widget) scroll() ScrollState {
	if w.scrollState == nil {
		return ScrollState{}
	}
	return *w.scrollState
}

// Handle turns scroll keys (while focused) and mouse scrolling within area into an Action once ScrollState and OnAction are set. Other messages pass through unchanged.
func (w Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	if w.scrollState == nil || w.onAction == nil {
		return msg
	}
	l := w.layout(area.Width)
	followEnd := w.scrollState.followEnd
	row, column := w.clamp(l, area, w.scrollState.row, w.scrollState.column, followEnd)
	switch m := msg.(type) {
	case tview.KeyMsg:
		if !w.focused {
			return msg
		}
		k := w.keybinds
		switch {
		case keybind.Matches(m, k.Top):
			row, column, followEnd = 0, 0, false
		case keybind.Matches(m, k.Bottom):
			column, followEnd = 0, true
		case keybind.Matches(m, k.Down):
			row++
		case keybind.Matches(m, k.Up):
			row, followEnd = row-1, false
		case keybind.Matches(m, k.Left):
			column--
		case keybind.Matches(m, k.Right):
			column++
		case keybind.Matches(m, k.PageDown):
			row += area.Height
		case keybind.Matches(m, k.PageUp):
			row, followEnd = row-area.Height, false
		default:
			return msg
		}
	case tview.MouseMsg:
		if !area.Contains(m.Position()) {
			return msg
		}
		switch m.Action {
		case tview.MouseScrollUp:
			row, followEnd = row-1, false
		case tview.MouseScrollDown:
			row++
		case tview.MouseScrollLeft:
			column -= area.Width / 2
		case tview.MouseScrollRight:
			column += area.Width / 2
		default:
			return msg
		}
	default:
		return msg
	}
	// Store the position clamped to what is visible, so scrolling back does not first work through overshoot.
	row, column = w.clamp(l, area, row, column, false)
	return w.onAction(Action{row: row, column: column, followEnd: followEnd})
}
