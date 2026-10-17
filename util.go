package tview

import (
	"github.com/gdamore/tcell/v3"
	"github.com/rockorager/go-uucode"
)

type Style = tcell.Style

// HorizontalAlignment is where something goes horizontally.
type HorizontalAlignment int

const (
	HorizontalAlignmentLeft HorizontalAlignment = iota
	HorizontalAlignmentCenter
	HorizontalAlignmentRight
)

// VerticalAlignment is where something goes vertically.
type VerticalAlignment int

const (
	VerticalAlignmentTop VerticalAlignment = iota
	VerticalAlignmentCenter
	VerticalAlignmentBottom
)

// Print draws text on row y from x, at most width cells wide and aligned within them, and returns the width it drew. Text too wide for a right or center alignment loses its start, or both ends.
func Print(screen Screen, text string, x, y, width int, align HorizontalAlignment, style Style) int {
	if width <= 0 {
		return 0
	}
	textWidth := uucode.StringWidth(text)
	cut := 0
	switch align {
	case HorizontalAlignmentRight:
		cut = textWidth - width
	case HorizontalAlignmentCenter:
		cut = (textWidth - width) / 2
	}
	clusters := uucode.NewGraphemeWidthIterator(text)
	for cut > 0 {
		cluster, ok := clusters.Next()
		if !ok {
			break
		}
		cut -= cluster.Width
		textWidth -= cluster.Width
	}
	if textWidth < width {
		switch align {
		case HorizontalAlignmentRight:
			x += width - textWidth
		case HorizontalAlignmentCenter:
			x += width/2 - textWidth/2
		}
	}

	drawn := 0
	for c, ok := clusters.Next(); ok; c, ok = clusters.Next() {
		cluster, w := text[c.Start:c.End], c.Width
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

// MergeStyle layers b on top of a and returns the result, merging every style component. Colors (foreground, background, underline) set on b — i.e. not [tcell.ColorDefault] — override a's; otherwise a's are kept. The underline style and hyperlink set on b likewise win over a's. Boolean attributes (bold, dim, italic, blink, reverse, strikethrough) are the union of both.
func MergeStyle(a, b Style) Style {
	fg := b.GetForeground()
	if fg == tcell.ColorDefault {
		fg = a.GetForeground()
	}
	bg := b.GetBackground()
	if bg == tcell.ColorDefault {
		bg = a.GetBackground()
	}

	// Underline carries an on/off+style and a separate color. A non-None style on b (which is also set when b enables a plain underline) wins; otherwise a's is kept. The same fallback applies to the underline color.
	ulStyle := b.GetUnderlineStyle()
	if ulStyle == tcell.UnderlineStyleNone {
		ulStyle = a.GetUnderlineStyle()
	}
	ulColor := b.GetUnderlineColor()
	if ulColor == tcell.ColorDefault {
		ulColor = a.GetUnderlineColor()
	}

	style := a.
		Foreground(fg).
		Background(bg).
		Bold(a.HasBold() || b.HasBold()).
		Dim(a.HasDim() || b.HasDim()).
		Italic(a.HasItalic() || b.HasItalic()).
		Blink(a.HasBlink() || b.HasBlink()).
		Reverse(a.HasReverse() || b.HasReverse()).
		StrikeThrough(a.HasStrikeThrough() || b.HasStrikeThrough()).
		Underline(ulStyle, ulColor)

	// Hyperlink: b's wins when set, otherwise a's (already carried by style) is kept.
	if id, url := b.GetUrl(); id != "" || url != "" {
		style = style.Url(url).UrlId(id)
	}
	return style
}
