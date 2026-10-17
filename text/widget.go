// Package text draws text, plain or in spans of different styles.
package text

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/layout"
	"github.com/rockorager/go-uucode"
)

// Widget draws text that wraps to its width.
type Widget struct {
	spans         []Span
	width, height layout.Length
	alignX        tview.Alignment
	wrapping      Wrapping
	style         tview.Style
}

var _ tview.Widget = Widget{}

// New returns text of content, as wide and tall as its lines, that wraps between words.
func New(content string) Widget {
	return Rich(Span{Text: content})
}

// Rich returns text of spans, each in its own style, as wide and tall as its lines, that wraps between words.
func Rich(spans ...Span) Widget {
	return Widget{spans: spans, width: layout.Shrink, height: layout.Shrink, wrapping: WrappingWord}
}

// Width sets the width of the text.
func (w Widget) Width(width layout.Length) Widget {
	w.width = width
	return w
}

// Height sets the height of the text.
func (w Widget) Height(height layout.Length) Widget {
	w.height = height
	return w
}

// AlignX sets the horizontal alignment of each line.
func (w Widget) AlignX(alignment tview.Alignment) Widget {
	w.alignX = alignment
	return w
}

// Wrapping sets how lines wider than the text are broken.
func (w Widget) Wrapping(wrapping Wrapping) Widget {
	w.wrapping = wrapping
	return w
}

// Style sets the style beneath the styles of the spans. Unset colors and attributes are those of the cells the text is drawn on.
func (w Widget) Style(style tview.Style) Widget {
	w.style = style
	return w
}

// Size returns the width and height of the text.
func (w Widget) Size() (width, height layout.Length) {
	return w.width, w.height
}

// Layout returns the size of the text within limits, where its content is as wide as the longest line and as tall as the lines it wraps to.
func (w Widget) Layout(limits layout.Limits) layout.Size {
	return layout.Sized(limits, w.width, w.height, func(limits layout.Limits) layout.Size {
		lines := Lines(w.spans, limits.Bounds().Width, w.wrapping)
		return layout.Size{Width: longest(lines), Height: len(lines)}
	})
}

// Draw draws the lines in area.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	if area.Width <= 0 || area.Height <= 0 {
		return
	}

	lines := Lines(w.spans, area.Width, w.wrapping)
	// Lines off the screen, as in a viewport, are skipped.
	_, bottom := screen.Size()
	for i := max(-area.Y, 0); i < min(area.Height, bottom-area.Y, len(lines)); i++ {
		line := lines[i]
		// A line wider than area is cut off on the side it is not aligned to, or on both if it is centered.
		start := 0
		switch w.alignX {
		case tview.AlignmentCenter:
			start = (area.Width - Width(line)) / 2
		case tview.AlignmentRight:
			start = area.Width - Width(line)
		}
		w.drawLine(screen, line, area.X+start, area.Y+i, area)
	}
}

// drawLine draws line starting at x, keeping only the cells within area.
func (w Widget) drawLine(screen tview.Screen, line []Span, x, y int, area tview.Rectangle) {
	for _, span := range line {
		style := tview.MergeStyle(w.style, span.Style)
		// The cells beneath a span mostly share a style, so it is merged once for each run of them.
		var beneath, merged tview.Style
		known := false
		clusters := uucode.NewGraphemeWidthIterator(span.Text)
		for c, ok := clusters.Next(); ok; c, ok = clusters.Next() {
			cluster, width := span.Text[c.Start:c.End], max(c.Width, 1)
			// What is past the right edge is not drawn.
			if x >= area.X+area.Width {
				return
			}
			if x >= area.X && x+width <= area.X+area.Width {
				if _, under, _ := screen.Get(x, y); !known || under != beneath {
					beneath, merged, known = under, tview.MergeStyle(under, style), true
				}
				screen.Put(x, y, cluster, merged)
			}
			x += width
		}
	}
}

// Handle passes msg through unchanged.
func (Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg { return msg }

// longest returns the width of the widest of lines.
func longest(lines [][]Span) int {
	widest := 0
	for _, line := range lines {
		widest = max(widest, Width(line))
	}
	return widest
}
