// Package scrollbar draws a vertical scroll bar.
package scrollbar

import (
	"github.com/ayn2op/tview/layout"
	"math/bits"

	"github.com/ayn2op/tview"
	"github.com/gdamore/tcell/v3"
)

// Arrows is which ends of the scroll bar have an arrow.
type Arrows uint8

const (
	ArrowsStart Arrows = 1 << iota
	ArrowsEnd

	ArrowsNone Arrows = 0
	ArrowsBoth        = ArrowsStart | ArrowsEnd
)

func (a Arrows) start() bool { return a&ArrowsStart != 0 }
func (a Arrows) end() bool   { return a&ArrowsEnd != 0 }

// GlyphSet is the characters a scroll bar is drawn with: the track, the thumb, and the arrows.
type GlyphSet struct {
	TrackVertical string
	ThumbVertical string

	ArrowVerticalStart string
	ArrowVerticalEnd   string
}

// MinimalGlyphSet returns a solid thumb on an empty track.
func MinimalGlyphSet() GlyphSet {
	g := BoxDrawingGlyphSet()
	g.TrackVertical = " "
	return g
}

// BoxDrawingGlyphSet returns a solid thumb on a box-drawing track.
func BoxDrawingGlyphSet() GlyphSet {
	return GlyphSet{
		TrackVertical:      "│",
		ThumbVertical:      "█",
		ArrowVerticalStart: "▲",
		ArrowVerticalEnd:   "▼",
	}
}

// Widget draws a vertical scroll bar for content of a length scrolled by an offset in a viewport. It is hidden when everything fits.
type Widget struct {
	content, viewport, offset int
	trackStyle, thumbStyle    tcell.Style
	arrowStyle                tcell.Style
	glyphs                    GlyphSet
	arrows                    Arrows
}

var _ tview.Element = Widget{}

// New returns a scroll bar with a minimal glyph set and no arrows.
func New() Widget {
	return Widget{
		trackStyle: tcell.StyleDefault.Dim(true),
		arrowStyle: tcell.StyleDefault.Dim(true),
		glyphs:     MinimalGlyphSet(),
	}
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

// GlyphSet sets the characters the scroll bar is drawn with.
func (w Widget) GlyphSet(glyphs GlyphSet) Widget {
	w.glyphs = glyphs
	return w
}

// Arrows sets which ends have an arrow.
func (w Widget) Arrows(arrows Arrows) Widget {
	w.arrows = arrows
	return w
}

// TrackStyle sets the style of the track.
func (w Widget) TrackStyle(style tcell.Style) Widget {
	w.trackStyle = style
	return w
}

// ThumbStyle sets the style of the thumb.
func (w Widget) ThumbStyle(style tcell.Style) Widget {
	w.thumbStyle = style
	return w
}

// HasStartArrow reports whether the top cell is an arrow.
func (w Widget) HasStartArrow() bool {
	return w.arrows.start()
}

// TrackCells returns the number of cells of a scroll bar length cells long that are not arrows.
func (w Widget) TrackCells(length int) int {
	return max(length-bits.OnesCount8(uint8(w.arrows)), 0)
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
	size = min(max(track*viewport/content, 1), track)
	return (track - size) * min(w.offset, maxOffset) / maxOffset, size
}

// Size returns Fill, as the scroll bar takes its whole area.
func (Widget) Size() (width, height layout.Length) {
	return layout.Fill, layout.Fill
}

// Layout returns the size of limits, as the scroll bar takes its whole area.
func (Widget) Layout(limits layout.Limits) layout.Size {
	return layout.Atomic(limits, layout.Fill, layout.Fill)
}

// Draw clears area and draws the scroll bar down its first column, unless all the content is visible.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	screen.FillArea(area.X, area.Y, area.Width, area.Height, ' ', tcell.StyleDefault)
	if area.Width <= 0 || w.content <= w.viewport || w.TrackCells(area.Height) == 0 {
		return
	}

	x, y := area.X, area.Y
	if w.arrows.start() {
		screen.Put(x, y, w.glyphs.ArrowVerticalStart, w.arrowStyle)
		y++
	}
	thumbStart, thumbSize := w.Thumb(area.Height)
	for cell := range w.TrackCells(area.Height) {
		glyph, style := w.glyphs.TrackVertical, w.trackStyle
		if cell >= thumbStart && cell < thumbStart+thumbSize {
			glyph, style = w.glyphs.ThumbVertical, w.thumbStyle
		}
		screen.Put(x, y, glyph, style)
		y++
	}
	if w.arrows.end() {
		screen.Put(x, y, w.glyphs.ArrowVerticalEnd, w.arrowStyle)
	}
}

// Handle passes msg through unchanged; the list drawing the scroll bar handles clicks on it.
func (Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg { return msg }
