// Package scrollbar draws a vertical or horizontal scroll bar.
package scrollbar

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/layout"
	"github.com/gdamore/tcell/v3"
)

// SymbolSet is the characters a scroll bar is drawn with. An empty Track, Begin, or End is not drawn.
type SymbolSet struct {
	Track, Thumb string
	Begin, End   string
}

// SymbolSetVertical returns a solid thumb on a single-line track between arrows, for a vertical scroll bar.
func SymbolSetVertical() SymbolSet {
	return SymbolSet{Track: "│", Thumb: "█", Begin: "↑", End: "↓"}
}

// SymbolSetDoubleVertical returns a solid thumb on a double-line track between triangles, for a vertical scroll bar.
func SymbolSetDoubleVertical() SymbolSet {
	return SymbolSet{Track: "║", Thumb: "█", Begin: "▲", End: "▼"}
}

// SymbolSetHorizontal returns a solid thumb on a single-line track between arrows, for a horizontal scroll bar.
func SymbolSetHorizontal() SymbolSet {
	return SymbolSet{Track: "─", Thumb: "█", Begin: "←", End: "→"}
}

// SymbolSetDoubleHorizontal returns a solid thumb on a double-line track between triangles, for a horizontal scroll bar.
func SymbolSetDoubleHorizontal() SymbolSet {
	return SymbolSet{Track: "═", Thumb: "█", Begin: "◄", End: "►"}
}

// Widget draws a scroll bar for content of a length scrolled by an offset in a viewport. It is hidden when everything fits.
type Widget struct {
	content, viewport, offset int
	horizontal                bool
	symbols                   SymbolSet
	thumbStyle, trackStyle    tview.Style
	beginStyle, endStyle      tview.Style
}

var _ tview.Widget = Widget{}

// New returns a vertical scroll bar drawn with SymbolSetDoubleVertical.
func New() Widget {
	return Widget{symbols: SymbolSetDoubleVertical()}
}

// Horizontal sets whether the scroll bar runs along the first row of its area and not down its first column. It keeps its symbols, so set a horizontal set too.
func (w Widget) Horizontal(horizontal bool) Widget {
	w.horizontal = horizontal
	return w
}

// Lengths sets the length of the content and of the part of it that is visible.
func (w Widget) Lengths(content, viewport int) Widget {
	w.content, w.viewport = max(content, 0), max(viewport, 0)
	return w
}

// Offset sets how far the content is scrolled.
func (w Widget) Offset(offset int) Widget {
	w.offset = max(offset, 0)
	return w
}

// SymbolSet sets the characters the scroll bar is drawn with.
func (w Widget) SymbolSet(symbols SymbolSet) Widget {
	w.symbols = symbols
	return w
}

// ThumbSymbol sets the character of the thumb.
func (w Widget) ThumbSymbol(symbol string) Widget {
	w.symbols.Thumb = symbol
	return w
}

// TrackSymbol sets the character of the track, or none if it is empty.
func (w Widget) TrackSymbol(symbol string) Widget {
	w.symbols.Track = symbol
	return w
}

// BeginSymbol sets the character at the top end, or none if it is empty.
func (w Widget) BeginSymbol(symbol string) Widget {
	w.symbols.Begin = symbol
	return w
}

// EndSymbol sets the character at the bottom end, or none if it is empty.
func (w Widget) EndSymbol(symbol string) Widget {
	w.symbols.End = symbol
	return w
}

// Style sets the style of every part of the scroll bar.
func (w Widget) Style(style tview.Style) Widget {
	w.thumbStyle, w.trackStyle, w.beginStyle, w.endStyle = style, style, style, style
	return w
}

// ThumbStyle sets the style of the thumb.
func (w Widget) ThumbStyle(style tview.Style) Widget {
	w.thumbStyle = style
	return w
}

// TrackStyle sets the style of the track.
func (w Widget) TrackStyle(style tview.Style) Widget {
	w.trackStyle = style
	return w
}

// BeginStyle sets the style of the character at the top end.
func (w Widget) BeginStyle(style tview.Style) Widget {
	w.beginStyle = style
	return w
}

// EndStyle sets the style of the character at the bottom end.
func (w Widget) EndStyle(style tview.Style) Widget {
	w.endStyle = style
	return w
}

// HasBegin reports whether the top cell is the begin symbol.
func (w Widget) HasBegin() bool {
	return w.symbols.Begin != ""
}

// TrackCells returns the number of cells of a scroll bar length cells long that are not the begin or end symbol.
func (w Widget) TrackCells(length int) int {
	if w.symbols.Begin != "" {
		length--
	}
	if w.symbols.End != "" {
		length--
	}
	return max(length, 0)
}

// Thumb returns where the thumb starts and how long it is, in cells of the track of a scroll bar length cells long.
func (w Widget) Thumb(length int) (start, size int) {
	track := w.TrackCells(length)
	content := max(w.content, 1)
	viewport := min(max(w.viewport, 1), content)
	maxOffset := content - viewport
	if track == 0 || maxOffset == 0 {
		return 0, track
	}
	// Rounding to the nearest cell keeps the thumb at each end for as long.
	size = min(max(divide(track*viewport, content), 1), track)
	return divide((track-size)*min(w.offset, maxOffset), maxOffset), size
}

// divide returns numerator over denominator rounded to the nearest integer.
func divide(numerator, denominator int) int {
	return (numerator + denominator/2) / denominator
}

// Size returns Fill, as the scroll bar takes its whole area.
func (Widget) Size() (width, height layout.Length) {
	return layout.Fill, layout.Fill
}

// Layout returns the size of limits, as the scroll bar takes its whole area.
func (Widget) Layout(limits layout.Limits) layout.Size {
	return layout.Atomic(limits, layout.Fill, layout.Fill)
}

// Draw clears area and draws the scroll bar down its first column, or along its first row if it is horizontal, unless all the content is visible.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	screen.FillArea(area.X, area.Y, area.Width, area.Height, ' ', tcell.StyleDefault)
	length := area.Height
	if w.horizontal {
		length = area.Width
	}
	if area.Width <= 0 || area.Height <= 0 || w.content <= w.viewport || w.TrackCells(length) == 0 {
		return
	}

	// put draws symbol in cell i of the scroll bar.
	put := func(i int, symbol string, style tview.Style) {
		if w.horizontal {
			screen.Put(area.X+i, area.Y, symbol, style)
		} else {
			screen.Put(area.X, area.Y+i, symbol, style)
		}
	}
	i := 0
	if w.symbols.Begin != "" {
		put(i, w.symbols.Begin, w.beginStyle)
		i++
	}
	thumbStart, thumbSize := w.Thumb(length)
	for cell := range w.TrackCells(length) {
		switch {
		case cell >= thumbStart && cell < thumbStart+thumbSize:
			put(i, w.symbols.Thumb, w.thumbStyle)
		case w.symbols.Track != "":
			put(i, w.symbols.Track, w.trackStyle)
		}
		i++
	}
	if w.symbols.End != "" {
		put(i, w.symbols.End, w.endStyle)
	}
}

// Handle passes msg through unchanged; the list drawing the scroll bar handles clicks on it.
func (Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg { return msg }
