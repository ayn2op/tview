// Package textview displays read-only styled text.
package textview

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/layout"
	"github.com/ayn2op/tview/richtext"
	"github.com/rivo/uniseg"
)

// Widget draws text.
type Widget struct {
	text          richtext.Text
	width, height layout.Length
	wrap          bool
	wordWrap      bool
	alignment     tview.Alignment
	style         tview.Style
}

var _ tview.Widget = Widget{}

// New returns a text view of text that wraps on words.
func New(text richtext.Text) Widget {
	return Widget{
		text:      text,
		width:     layout.Fill,
		height:    layout.Fill,
		wrap:      true,
		wordWrap:  true,
		alignment: tview.AlignmentLeft,
	}
}

// Width sets the width of the text view.
func (w Widget) Width(width layout.Length) Widget {
	w.width = width
	return w
}

// Height sets the height of the text view.
func (w Widget) Height(height layout.Length) Widget {
	w.height = height
	return w
}

// Wrap sets whether lines longer than the width continue on the next line.
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
func (w Widget) Style(style tview.Style) Widget {
	w.style = style
	return w
}

// Size returns the width and height of the text view.
func (w Widget) Size() (width, height layout.Length) {
	return w.width, w.height
}

// Layout returns the size of the text view within limits, where its content is as wide as the longest line and as tall as the lines the text wraps to.
func (w Widget) Layout(limits layout.Limits) layout.Size {
	return layout.Sized(limits, w.width, w.height, func(limits layout.Limits) layout.Size {
		l := w.layout(limits.Bounds().Width)
		return layout.Size{Width: l.longest, Height: len(l.lines)}
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

// origin returns where, in a canvas as wide as the longest line or area, the view starts.
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

// Draw fills area with the style and draws the lines.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	screen.FillArea(area.X, area.Y, area.Width, area.Height, ' ', w.style)
	if area.Width <= 0 || area.Height <= 0 {
		return
	}

	l := w.layout(area.Width)
	canvas := max(l.longest, area.Width)
	viewStart := w.origin(l, area.Width)
	// Lines off the screen, as in a viewport, are skipped.
	_, bottom := screen.Size()
	for i := max(-area.Y, 0); i < min(area.Height, bottom-area.Y, len(l.lines)); i++ {
		line := l.lines[i]
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

// Handle passes msg through unchanged.
func (Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg { return msg }
