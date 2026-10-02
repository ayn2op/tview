// Package textview displays read-only styled text that can be scrolled.
package textview

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/layout"
	"github.com/ayn2op/tview/richtext"
	"github.com/gdamore/tcell/v3"
	"github.com/rivo/uniseg"
)

// Widget draws text and, with a ScrollState, turns scroll input into Changes.
type Widget struct {
	text          richtext.Text
	scrollState   *ScrollState
	width, height tview.Length
	wrap          bool
	wordWrap      bool
	alignment     tview.Alignment
	style         tcell.Style
	keybind       func(tview.KeyMsg) (Action, bool)
	onChange      func(Change) tview.Msg
	focused       bool
}

var _ tview.Element = Widget{}

// New returns a text view of text that wraps on words and fills its parent. It scrolls only once ScrollState and OnChange are set.
func New(text richtext.Text) Widget {
	return Widget{
		text:      text,
		width:     tview.Fill,
		height:    tview.Fill,
		wrap:      true,
		wordWrap:  true,
		alignment: tview.AlignmentLeft,
		keybind:   DefaultKeybind,
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

// Keybind sets the function that turns keys into Actions, DefaultKeybind unless set.
func (w Widget) Keybind(f func(tview.KeyMsg) (Action, bool)) Widget {
	w.keybind = f
	return w
}

// ScrollState sets how far the text view is scrolled.
func (w Widget) ScrollState(scrollState *ScrollState) Widget {
	w.scrollState = scrollState
	return w
}

// OnChange makes the text view scrollable, turning scroll input into the message f returns for the Change, which the model applies with ScrollState.Apply.
func (w Widget) OnChange(f func(Change) tview.Msg) Widget {
	w.onChange = f
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

// Layout returns the size of the text view within limits, where its content is as wide as the longest line and as tall as the lines the text wraps to.
func (w Widget) Layout(limits layout.Limits) tview.Size {
	return layout.Sized(limits, w.width, w.height, func(limits layout.Limits) tview.Size {
		l := w.layout(limits.Bounds().Width)
		return tview.Size{Width: l.longest, Height: len(l.lines)}
	})
}

// layout is the text wrapped to a width.
type textLayout struct {
	lines   richtext.Text
	longest int
}

func (w Widget) layout(width int) textLayout {
	var l textLayout
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
func (w Widget) origin(l textLayout, width int) int {
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
func (w Widget) clamp(l textLayout, area tview.Rectangle, row, column int, followEnd bool) (int, int) {
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
	screen.FillArea(area.X, area.Y, area.Width, area.Height, ' ', w.style)
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

// Handle turns scroll keys (while focused) and mouse scrolling within area into a Change once ScrollState and OnChange are set. Other messages pass through unchanged.
func (w Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	if w.scrollState == nil || w.onChange == nil {
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
		action, ok := w.keybind(m)
		if !ok {
			return msg
		}
		switch action {
		case ActionTop:
			row, column, followEnd = 0, 0, false
		case ActionBottom:
			column, followEnd = 0, true
		case ActionDown:
			row++
		case ActionUp:
			row, followEnd = row-1, false
		case ActionLeft:
			column--
		case ActionRight:
			column++
		case ActionPageDown:
			row += area.Height
		case ActionPageUp:
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
	return w.onChange(Change{row: row, column: column, followEnd: followEnd})
}
