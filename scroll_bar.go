package tview

import "github.com/gdamore/tcell/v3"

// ScrollBarArrows controls which endcaps are rendered.
type ScrollBarArrows uint8

const (
	ScrollBarArrowsNone ScrollBarArrows = iota
	ScrollBarArrowsStart
	ScrollBarArrowsEnd
	ScrollBarArrowsBoth
)

func (a ScrollBarArrows) hasStart() bool {
	return a == ScrollBarArrowsStart || a == ScrollBarArrowsBoth
}

func (a ScrollBarArrows) hasEnd() bool {
	return a == ScrollBarArrowsEnd || a == ScrollBarArrowsBoth
}

// TrackClickBehavior configures behavior when clicking scrollBar track cells
// outside the thumb.
type TrackClickBehavior uint8

const (
	TrackClickBehaviorPage TrackClickBehavior = iota
	TrackClickBehaviorJumpToClick
)

// ScrollLengths bundles content and viewport lengths in logical units.
type ScrollLengths struct {
	ContentLen  int
	ViewportLen int
}

const subcell = 8

// GlyphSet defines vertical track, arrow, and fractional thumb glyphs.
type GlyphSet struct {
	TrackVertical string

	ArrowVerticalStart string
	ArrowVerticalEnd   string

	ThumbVerticalLower [8]string
	ThumbVerticalUpper [8]string
}

// MinimalGlyphSet returns the minimal glyph set (space track, fractional thumbs).
func MinimalGlyphSet() GlyphSet {
	g := LegacyComputingGlyphSet()
	g.TrackVertical = " "
	return g
}

// BoxDrawingGlyphSet returns box-drawing track glyphs with legacy fractional symbols.
func BoxDrawingGlyphSet() GlyphSet {
	return LegacyComputingGlyphSet()
}

// LegacyComputingGlyphSet returns legacy-computing symbols for full 1/8 fractional fidelity.
func LegacyComputingGlyphSet() GlyphSet {
	return GlyphSet{
		TrackVertical: "│",

		ArrowVerticalStart: "▲",
		ArrowVerticalEnd:   "▼",

		ThumbVerticalLower: [8]string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"},
		ThumbVerticalUpper: [8]string{"▔", "🮂", "🮃", "▀", "🮄", "🮅", "🮆", "█"},
	}
}

// UnicodeGlyphSet returns a standard-unicode-only approximation set.
func UnicodeGlyphSet() GlyphSet {
	return GlyphSet{
		TrackVertical: "│",

		ArrowVerticalStart: "▲",
		ArrowVerticalEnd:   "▼",

		ThumbVerticalLower: [8]string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"},
		ThumbVerticalUpper: [8]string{"▔", "▔", "▀", "▀", "▀", "▀", "█", "█"},
	}
}

// ScrollBar renders a vertical customizable scrollBar widget.
type ScrollBar struct {
	*Box

	autoHide    bool
	contentLen  int
	viewportLen int
	offset      int

	trackStyle tcell.Style
	thumbStyle tcell.Style
	arrowStyle tcell.Style

	glyphSet GlyphSet
	arrows   ScrollBarArrows

	trackClickBehavior TrackClickBehavior
	scrollStep         int

	showTrack bool
}

// NewScrollBar returns a new vertical scrollBar.
func NewScrollBar() *ScrollBar {
	return &ScrollBar{
		Box:                NewBox(),
		autoHide:           true,
		trackStyle:         tcell.StyleDefault.Dim(true),
		thumbStyle:         tcell.StyleDefault,
		arrowStyle:         tcell.StyleDefault.Dim(true),
		glyphSet:           MinimalGlyphSet(),
		arrows:             ScrollBarArrowsNone,
		trackClickBehavior: TrackClickBehaviorPage,
		scrollStep:         1,
		showTrack:          true,
	}
}

// NewVerticalScrollBar creates a vertical scrollBar from lengths.
func NewVerticalScrollBar(lengths ScrollLengths) *ScrollBar {
	return NewScrollBar().SetLengths(lengths)
}

// SetLengths sets content and viewport lengths.
func (s *ScrollBar) SetLengths(lengths ScrollLengths) *ScrollBar {
	s.contentLen = max(lengths.ContentLen, 0)
	s.viewportLen = max(lengths.ViewportLen, 0)
	return s
}

// SetOffset sets the logical offset.
func (s *ScrollBar) SetOffset(offset int) *ScrollBar {
	s.offset = max(offset, 0)
	return s
}

// SetGlyphSet applies a glyph set.
func (s *ScrollBar) SetGlyphSet(g GlyphSet) *ScrollBar {
	s.glyphSet = g
	return s
}

// SetArrows sets which arrow endcaps are rendered.
func (s *ScrollBar) SetArrows(arrows ScrollBarArrows) *ScrollBar {
	if s.arrows != arrows {
		s.arrows = arrows
	}
	return s
}

// SetTrackClickBehavior sets behavior used for track clicks.
func (s *ScrollBar) SetTrackClickBehavior(behavior TrackClickBehavior) *ScrollBar {
	if s.trackClickBehavior != behavior {
		s.trackClickBehavior = behavior
	}
	return s
}

// SetScrollStep sets scroll step used by wheel and arrow interactions.
func (s *ScrollBar) SetScrollStep(step int) *ScrollBar {
	if step < 1 {
		step = 1
	}
	if s.scrollStep != step {
		s.scrollStep = step
	}
	return s
}

// ScrollStep returns scroll step used by wheel and arrow interactions.
func (s *ScrollBar) ScrollStep() int {
	return s.scrollStep
}

// TrackClickBehavior returns behavior used for track clicks.
func (s *ScrollBar) TrackClickBehavior() TrackClickBehavior {
	return s.trackClickBehavior
}

// HasStartArrow reports whether a start arrow is rendered.
func (s *ScrollBar) HasStartArrow() bool {
	return s.arrows.hasStart()
}

// HasEndArrow reports whether an end arrow is rendered.
func (s *ScrollBar) HasEndArrow() bool {
	return s.arrows.hasEnd()
}

// TrackLengthExcludingArrowHeads returns track length excluding arrow cells.
func (s *ScrollBar) TrackLengthExcludingArrowHeads(length int) int {
	return s.trackLengthExcludingArrowHeads(length)
}

// SetAutoHide controls whether the scrollBar is hidden when there is nothing to scroll.
func (s *ScrollBar) SetAutoHide(autoHide bool) *ScrollBar {
	if s.autoHide != autoHide {
		s.autoHide = autoHide
	}
	return s
}

// SetThumbGlyph sets all thumb glyphs to a single symbol.
func (s *ScrollBar) SetThumbGlyph(glyph string) *ScrollBar {
	for i := range len(s.glyphSet.ThumbVerticalLower) {
		s.glyphSet.ThumbVerticalLower[i] = glyph
		s.glyphSet.ThumbVerticalUpper[i] = glyph
	}
	return s
}

// SetThumbStyle sets the thumb style.
func (s *ScrollBar) SetThumbStyle(style tcell.Style) *ScrollBar {
	if s.thumbStyle != style {
		s.thumbStyle = style
	}
	return s
}

// SetTrackGlyph sets the track symbol and visibility.
func (s *ScrollBar) SetTrackGlyph(glyph string, visible bool) *ScrollBar {
	s.glyphSet.TrackVertical = glyph
	s.showTrack = visible
	return s
}

// SetTrackStyle sets the track style.
func (s *ScrollBar) SetTrackStyle(style tcell.Style) *ScrollBar {
	if s.trackStyle != style {
		s.trackStyle = style
	}
	return s
}

// SetArrowStyle sets the arrow endcap style.
func (s *ScrollBar) SetArrowStyle(style tcell.Style) *ScrollBar {
	if s.arrowStyle != style {
		s.arrowStyle = style
	}
	return s
}

func (s *ScrollBar) trackLengthExcludingArrowHeads(length int) int {
	if length <= 0 {
		return 0
	}
	arrows := 0
	if s.arrows.hasStart() {
		arrows++
	}
	if s.arrows.hasEnd() {
		arrows++
	}
	return max(length-arrows, 0)
}

func (s *ScrollBar) viewportLength(length int) int {
	if s.viewportLen > 0 {
		return s.viewportLen
	}
	return max(length, 0)
}

type scrollMetrics struct {
	trackCells int
	trackLen   int
	thumbLen   int
	thumbStart int
}

// metrics computes scrollBar geometry in subcell units.
func (s *ScrollBar) metrics(length int) scrollMetrics {
	trackCells := s.trackLengthExcludingArrowHeads(length)
	return computeScrollMetrics(trackCells, s.contentLen, s.viewportLength(length), s.offset)
}

func computeScrollMetrics(trackCells int, contentLen int, viewportLen int, offset int) scrollMetrics {
	trackLen := trackCells * subcell
	if trackLen == 0 {
		return scrollMetrics{}
	}

	contentLen = max(contentLen, 1)
	viewportLen = min(max(viewportLen, 1), contentLen)
	maxOffset := max(contentLen-viewportLen, 0)
	offset = min(max(offset, 0), maxOffset)

	if maxOffset == 0 {
		return scrollMetrics{trackCells: trackCells, trackLen: trackLen, thumbLen: trackLen, thumbStart: 0}
	}

	// Use subcell math so the thumb can move in 1/8-cell steps while staying proportional to viewport/content size.
	thumbLen := min(max((trackLen*viewportLen)/contentLen, subcell), trackLen)
	thumbTravel := max(trackLen-thumbLen, 0)
	thumbStart := (thumbTravel * offset) / maxOffset
	return scrollMetrics{trackCells: trackCells, trackLen: trackLen, thumbLen: thumbLen, thumbStart: thumbStart}
}

func (s *ScrollBar) shouldDraw(length int, m scrollMetrics) bool {
	if length <= 0 || m.trackLen == 0 || s.contentLen <= 0 {
		return false
	}
	if s.autoHide {
		contentLen := max(s.contentLen, 1)
		viewportLen := min(max(s.viewportLength(length), 1), contentLen)
		if contentLen <= viewportLen {
			return false
		}
	}
	return true
}

func cellFill(m scrollMetrics, cellIndex int) (start int, fillLen int) {
	if m.thumbLen == 0 {
		return 0, 0
	}
	cellStart := cellIndex * subcell
	cellEnd := cellStart + subcell
	thumbEnd := m.thumbStart + m.thumbLen
	start = max(m.thumbStart, cellStart)
	end := min(thumbEnd, cellEnd)
	if end <= start {
		return 0, 0
	}
	// Convert absolute subcell coverage into cell-local [start,len] used by fractional glyph selection.
	fillLen = min(end-start, subcell)
	start = min(max(start-cellStart, 0), subcell)
	return start, fillLen
}

func (s *ScrollBar) glyphForVertical(start, fillLen int) (string, tcell.Style) {
	if fillLen <= 0 {
		if !s.showTrack {
			return " ", s.trackStyle
		}
		return s.glyphSet.TrackVertical, s.trackStyle
	}
	if fillLen >= subcell {
		return s.glyphSet.ThumbVerticalLower[7], s.thumbStyle
	}
	ix := fillLen - 1
	if start == 0 {
		return s.glyphSet.ThumbVerticalUpper[ix], s.thumbStyle
	}
	return s.glyphSet.ThumbVerticalLower[ix], s.thumbStyle
}

func (s *ScrollBar) put(screen tcell.Screen, x, y, index int, glyph string, style tcell.Style) {
	screen.Put(x, y+index, glyph, style)
}

// Draw draws the scrollBar.
func (s *ScrollBar) Draw(screen tcell.Screen) {
	s.DrawForSubclass(screen, s)

	x, y, _, height := s.InnerRect()
	if height <= 0 {
		return
	}
	length := height
	m := s.metrics(length)
	if !s.shouldDraw(length, m) {
		return
	}

	idx := 0
	if s.arrows.hasStart() {
		s.put(screen, x, y, idx, s.glyphSet.ArrowVerticalStart, s.arrowStyle)
		idx++
	}

	for cell := range m.trackCells {
		start, fillLen := cellFill(m, cell)
		glyph, style := s.glyphForVertical(start, fillLen)
		s.put(screen, x, y, idx, glyph, style)
		idx++
	}

	if s.arrows.hasEnd() {
		s.put(screen, x, y, idx, s.glyphSet.ArrowVerticalEnd, s.arrowStyle)
	}
}

var _ Model = &ScrollBar{}
