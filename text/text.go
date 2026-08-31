// Package text provides styled text values.
package text

import (
	"slices"
	"strings"

	"github.com/gdamore/tcell/v3"
	"github.com/rivo/uniseg"
)

// Segment is a styled string.
type Segment struct {
	Text  string
	Style tcell.Style
}

// NewSegment returns a styled segment.
func NewSegment(text string, style tcell.Style) Segment {
	return Segment{Text: text, Style: style}
}

// Width returns the segment's cell width.
func (s Segment) Width() int {
	return uniseg.StringWidth(s.Text)
}

// Line is a sequence of styled segments.
type Line []Segment

// NewLine returns a line from segments.
func NewLine(segments ...Segment) Line {
	return segments
}

// Clone returns an independent copy.
func (l Line) Clone() Line {
	return slices.Clone(l)
}

// Width returns the line's cell width.
func (l Line) Width() int {
	width := 0
	for _, segment := range l {
		width += segment.Width()
	}
	return width
}

// Text is a sequence of styled lines.
type Text []Line

// New returns text from lines.
func New(lines ...Line) Text {
	return lines
}

// Clone returns an independent copy.
func (t Text) Clone() Text {
	clone := make(Text, len(t))
	for i, line := range t {
		clone[i] = line.Clone()
	}
	return clone
}

// Width returns the widest line's cell width.
func (t Text) Width() int {
	width := 0
	for _, line := range t {
		width = max(width, line.Width())
	}
	return width
}

// Height returns the line count.
func (t Text) Height() int {
	return len(t)
}

// Builder builds styled text.
type Builder struct {
	lines   Text
	current Line
}

// Write appends styled text.
func (b *Builder) Write(value string, style tcell.Style) {
	for {
		line, rest, found := strings.Cut(value, "\n")
		b.WriteSegment(NewSegment(line, style))
		if !found {
			return
		}
		b.NewLine()
		value = rest
	}
}

// WriteSegment appends a segment.
func (b *Builder) WriteSegment(segment Segment) {
	b.current = appendSegment(b.current, segment)
}

// WriteText appends text.
func (b *Builder) WriteText(text Text) {
	for i, line := range text {
		if i > 0 {
			b.NewLine()
		}
		for _, segment := range line {
			b.WriteSegment(segment)
		}
	}
}

// LineEmpty reports whether the current line is empty.
func (b *Builder) LineEmpty() bool {
	return len(b.current) == 0
}

// NewLine starts a new line.
func (b *Builder) NewLine() {
	b.lines = append(b.lines, b.current)
	b.current = nil
}

// Finish returns the built text.
func (b *Builder) Finish() Text {
	if len(b.current) > 0 || len(b.lines) == 0 {
		b.NewLine()
	}
	return b.lines
}

func appendSegment(line Line, segment Segment) Line {
	if segment.Text == "" {
		return line
	}
	if n := len(line); n > 0 && line[n-1].Style == segment.Style {
		line[n-1].Text += segment.Text
		return line
	}
	return append(line, segment)
}
