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

// Line is a list of styled segments.
type Line []Segment

// Clone returns a copy of this line with an independent backing array.
func (l Line) Clone() Line {
	out := make(Line, len(l))
	copy(out, l)
	return out
}

// NewLine returns a line from the provided segments, skipping empty segments.
func NewLine(segments ...Segment) Line {
	line := make(Line, 0, len(segments))
	for _, segment := range segments {
		if segment.Text == "" {
			continue
		}
		line = append(line, segment)
	}
	return line
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

// WriteAll is just like [Write] but takes a list of strings/segments.
// All strings will have the same style as the first argument.
// Unknown types will be discarded.
func (b *LineBuilder) WriteAll(style tcell.Style, list ...any) {
	for _, item := range list {
		switch item := item.(type) {
		case Segment:
			b.Write(item.Text, item.Style)
		case string:
			b.Write(item, style)
		}
	}
}

func (b *LineBuilder) writeSegment(text string, style tcell.Style) {
	if text == "" {
		return
	}
	if n := len(b.current); n > 0 && b.current[n-1].Style == style {
		b.current[n-1].Text += text
		return
	}
	b.current = append(b.current, Segment{Text: text, Style: style})
}

// AppendLines appends fully built lines into the builder.
func (b *LineBuilder) AppendLines(lines []Line) {
	if len(lines) == 0 {
		return
	}
	for i, line := range lines {
		if i > 0 {
			b.NewLine()
		}
		for _, segment := range line {
			b.writeSegment(segment.Text, segment.Style)
		}
	}
}

// NewLine flushes the current line into the builder output.
func (b *LineBuilder) NewLine() {
	line := make(Line, len(b.current))
	copy(line, b.current)
	b.lines = append(b.lines, line)
	b.current = nil
}

// HasCurrentLine returns true when unflushed segments exist.
func (b *LineBuilder) HasCurrentLine() bool {
	return len(b.current) > 0
}

// Finish returns all built lines.
func (b *LineBuilder) Finish() []Line {
	if len(b.current) > 0 || len(b.lines) == 0 {
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
