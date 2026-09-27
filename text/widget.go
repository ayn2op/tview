// Package text draws plain text.
package text

import (
	"github.com/ayn2op/tview"
	"github.com/rivo/uniseg"
)

// Widget draws a single line of text.
type Widget struct {
	content string
}

var _ tview.Element = Widget{}

// New draws content in the style of the cells beneath it, so it takes the colors of the element it is drawn in.
func New(content string) Widget {
	return Widget{content: content}
}

// Size returns the width of the text and a height of one line.
func (w Widget) Size() (width, height tview.Length) {
	return tview.Fixed(uniseg.StringWidth(w.content)), tview.Fixed(1)
}

// Draw draws the text from the top-left corner of area, cut off at its right edge.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	if area.Height <= 0 {
		return
	}
	x, end := area.X, area.X+area.Width
	str, state := w.content, -1
	for len(str) > 0 {
		var cluster string
		var width int
		cluster, str, width, state = uniseg.FirstGraphemeClusterInString(str, state)
		if x+width > end {
			return
		}
		_, style, _ := screen.Get(x, area.Y)
		screen.Put(x, area.Y, cluster, style)
		x += width
	}
}

// Handle passes msg through unchanged.
func (Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	return msg
}
