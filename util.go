package tview

import (
	"github.com/gdamore/tcell/v3"
	"github.com/rivo/uniseg"
)

type Alignment int

const (
	AlignmentLeft Alignment = iota
	AlignmentCenter
	AlignmentRight
)

// Print draws text on row y from x, at most width cells wide and aligned within them, and returns the width it drew. Text too wide for a right or center alignment loses its start, or both ends.
func Print(screen Screen, text string, x, y, width int, alignment Alignment, style tcell.Style) int {
	if width <= 0 {
		return 0
	}
	textWidth := uniseg.StringWidth(text)
	cut := 0
	switch alignment {
	case AlignmentRight:
		cut = textWidth - width
	case AlignmentCenter:
		cut = (textWidth - width) / 2
	}
	state := -1
	for cut > 0 && text != "" {
		var w int
		_, text, w, state = uniseg.FirstGraphemeClusterInString(text, state)
		cut -= w
		textWidth -= w
	}
	if textWidth < width {
		switch alignment {
		case AlignmentRight:
			x += width - textWidth
		case AlignmentCenter:
			x += width/2 - textWidth/2
		}
	}

	drawn := 0
	for text != "" {
		var cluster string
		var w int
		cluster, text, w, state = uniseg.FirstGraphemeClusterInString(text, state)
		if drawn+w > width {
			break
		}
		// Fill the cells a wide cluster covers, then put the cluster in the first.
		for i := w - 1; i > 0; i-- {
			screen.Put(x+drawn+i, y, " ", style)
		}
		if w > 0 {
			screen.Put(x+drawn, y, cluster, style)
		}
		drawn += w
	}
	return drawn
}
