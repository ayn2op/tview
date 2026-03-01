package tview

import (
	"strings"

	"github.com/gdamore/tcell/v3"
	"github.com/rivo/uniseg"
)

// Segment is a styled piece of text.
type Segment struct {
	Text  string
	Style tcell.Style
}

// NewSegment returns a styled segment.
func NewSegment(text string, style tcell.Style) Segment {
	return Segment{Text: text, Style: style}
}

// Line is a list of styled segments with indent used on softwrapping.
type Line struct {
	Segments []Segment
	Indent   []Segment
}

// Clone returns a copy of this line with an independent backing array.
func (l Line) Clone() Line {
	out := Line{Segments: make([]Segment, len(l.Segments)), Indent: l.Indent}
	copy(out.Segments, l.Segments)
	return out
}

// NewLine returns a line from the provided segments, skipping empty segments.
func NewLine(segments ...Segment) Line {
	line := Line{Segments: make([]Segment, 0, len(segments))}
	for _, segment := range segments {
		if segment.Text == "" {
			continue
		}
		line.Segments = append(line.Segments, segment)
	}
	return line
}

// WithIndent sets the indent segment that will be inserted on every
// wrapped logical line
func (l Line) WithIndent(indent []Segment) Line {
	l.Indent = indent
	return l
}

// LineBuilder incrementally builds styled lines from text writes.
type LineBuilder struct {
	lines   []Line
	current Line
}

// NewLineBuilder returns a new line builder.
func NewLineBuilder() *LineBuilder {
	return &LineBuilder{}
}

// Write appends text with style and splits on newline boundaries.
func (b *LineBuilder) Write(text string, style tcell.Style) {
	if text == "" {
		return
	}
	for len(text) > 0 {
		nl := strings.IndexByte(text, '\n')
		if nl < 0 {
			b.writeSegment(text, style)
			return
		}
		if nl > 0 {
			b.writeSegment(text[:nl], style)
		}
		b.NewLine()
		text = text[nl+1:]
	}
}

// WriteSegments is just like Write but takes multiple arguments.
func (b *LineBuilder) WriteSegments(segments []Segment) {
	for _, seg := range segments {
		b.Write(seg.Text, seg.Style)
	}
}

func (b *LineBuilder) writeSegment(text string, style tcell.Style) {
	if text == "" {
		return
	}
	if n := len(b.current.Segments); n > 0 && b.current.Segments[n-1].Style == style {
		b.current.Segments[n-1].Text += text
		return
	}
	b.current.Segments = append(b.current.Segments, Segment{Text: text, Style: style})
}

// AppendLines appends fully built lines into the builder.
func (b *LineBuilder) AppendLines(lines []Line) {
	if len(lines) == 0 {
		return
	}
	for _, segment := range lines[0].Segments {
		b.writeSegment(segment.Text, segment.Style)
	}
	if len(lines) == 1 {
		return
	}
	b.NewLine()
	b.lines = append(b.lines, lines[1:]...)
}

// AppendLine appends a fully built line into the builder.
// Needs to be called after a .NewLine() or called on an empty builder.
func (b *LineBuilder) AppendLine(line Line) {
	b.lines = append(b.lines, line)
}

// NewLine flushes the current line into the builder output.
func (b *LineBuilder) NewLine() {
	line := Line{Segments: make([]Segment, len(b.current.Segments)), Indent: b.current.Indent}
	copy(line.Segments, b.current.Segments)
	b.lines = append(b.lines, line)
	b.current.Segments = nil
}

// SetIndent sets default new indentation segment
func (b *LineBuilder) SetIndent(indent []Segment) {
	b.current.Indent = indent
}

// HasCurrentLine returns true when unflushed segments exist.
func (b *LineBuilder) HasCurrentLine() bool {
	return len(b.current.Segments) > 0
}

// Finish returns all built lines.
func (b *LineBuilder) Finish() []Line {
	if len(b.current.Segments) > 0 || len(b.lines) == 0 {
		b.NewLine()
	}
	return b.lines
}

// stepState represents the current state of the grapheme parser.
type stepState struct {
	unisegState int
	boundaries  int
	grossLength int
}

// LineBreak returns whether the string can be broken into the next line after
// the returned grapheme cluster.
func (s *stepState) LineBreak() (lineBreak, optional bool) {
	switch s.boundaries & uniseg.MaskLine {
	case uniseg.LineCanBreak:
		return true, true
	case uniseg.LineMustBreak:
		return true, false
	}
	return false, false
}

// Width returns the grapheme cluster's width in cells.
func (s *stepState) Width() int {
	return s.boundaries >> uniseg.ShiftWidth
}

// GrossLength returns the grapheme cluster's length in bytes.
func (s *stepState) GrossLength() int {
	return s.grossLength
}

// step iterates over grapheme clusters of a string.
func step(str string, state *stepState) (cluster, rest string, newState *stepState) {
	if state == nil {
		state = &stepState{
			unisegState: -1,
		}
	}
	if len(str) == 0 {
		newState = state
		return
	}

	preState := state.unisegState
	cluster, rest, state.boundaries, state.unisegState = uniseg.StepString(str, preState)
	state.grossLength = len(cluster)
	if rest == "" && !uniseg.HasTrailingLineBreakInString(cluster) {
		state.boundaries &^= uniseg.MaskLine
	}

	newState = state
	return
}

// TaggedStringWidth returns the width of the given string needed to print it on
// screen.
func TaggedStringWidth(text string) (width int) {
	var state *stepState
	for len(text) > 0 {
		_, text, state = step(text, state)
		width += state.Width()
	}
	return
}

// WordWrap splits a text such that each resulting line does not exceed the
// given screen width.
func WordWrap(text string, width int) (lines []string) {
	if width <= 0 {
		return
	}

	var (
		state                                              *stepState
		lineWidth, lineLength, lastOption, lastOptionWidth int
	)
	str := text
	for len(str) > 0 {
		_, str, state = step(str, state)
		cWidth := state.Width()

		if lineWidth+cWidth > width {
			if lastOptionWidth == 0 {
				lines = append(lines, text[:lineLength])
				text = text[lineLength:]
				lineWidth, lineLength, lastOption, lastOptionWidth = 0, 0, 0, 0
			} else {
				lines = append(lines, text[:lastOption])
				text = text[lastOption:]
				lineWidth -= lastOptionWidth
				lineLength -= lastOption
				lastOption, lastOptionWidth = 0, 0
			}
		}

		lineWidth += cWidth
		lineLength += state.GrossLength()

		if lineBreak, optional := state.LineBreak(); lineBreak {
			if optional {
				lastOption = lineLength
				lastOptionWidth = lineWidth
			} else {
				lines = append(lines, strings.TrimRight(text[:lineLength], "\n\r"))
				text = text[lineLength:]
				lineWidth, lineLength, lastOption, lastOptionWidth = 0, 0, 0, 0
			}
		}
	}
	lines = append(lines, text)

	return
}
