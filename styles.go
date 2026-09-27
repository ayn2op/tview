package tview

import "github.com/gdamore/tcell/v3"

type Style = tcell.Style

// MergeStyle layers b on top of a and returns the result, merging every style component. Colors (foreground, background, underline) set on b — i.e. not [tcell.ColorDefault] — override a's; otherwise a's are kept. The underline style and hyperlink set on b likewise win over a's. Boolean attributes (bold, dim, italic, blink, reverse, strikethrough) are the union of both.
func MergeStyle(a, b tcell.Style) tcell.Style {
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
