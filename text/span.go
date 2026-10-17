package text

import (
	"slices"
	"strings"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/wordwrap"
	"github.com/rockorager/go-uucode"
)

// Span is a part of a text in one style.
type Span struct {
	Text  string
	Style tview.Style
}

// Wrapping is how lines wider than the text are broken.
type Wrapping int

const (
	// WrappingNone keeps lines whole, cut off at the edge.
	WrappingNone Wrapping = iota
	// WrappingWord breaks lines between words. A word wider than the text runs past the edge.
	WrappingWord
	// WrappingGlyph breaks lines at any character.
	WrappingGlyph
	// WrappingWordOrGlyph breaks lines between words, and words wider than the text at any character.
	WrappingWordOrGlyph
)

// Width returns the width of spans in cells.
func Width(spans []Span) int {
	width := 0
	for _, span := range spans {
		width += uucode.StringWidth(span.Text)
	}
	return width
}

// Lines splits spans at line breaks and wraps each line to width.
func Lines(spans []Span, width int, wrapping Wrapping) [][]Span {
	lines := splitLines(spans)
	if wrapping == WrappingNone || width <= 0 {
		return lines
	}

	wrapped := make([][]Span, 0, len(lines))
	for _, line := range lines {
		if wrapping == WrappingGlyph {
			wrapped = append(wrapped, wrapGlyphs(line, width)...)
		} else {
			wrapped = append(wrapped, wrapWords(line, width, wrapping == WrappingWordOrGlyph)...)
		}
	}
	return wrapped
}

// splitLines splits spans at line breaks, with a tab drawn and measured as a space.
func splitLines(spans []Span) [][]Span {
	if !slices.ContainsFunc(spans, func(span Span) bool { return strings.ContainsAny(span.Text, "\n\t") }) {
		return [][]Span{spans}
	}

	lines := [][]Span{nil}
	for _, span := range spans {
		for {
			part, rest, found := strings.Cut(span.Text, "\n")
			if found {
				part = strings.TrimSuffix(part, "\r")
			}
			if part = strings.ReplaceAll(part, "\t", " "); part != "" {
				lines[len(lines)-1] = append(lines[len(lines)-1], Span{Text: part, Style: span.Style})
			}
			if !found {
				break
			}
			lines = append(lines, nil)
			span.Text = rest
		}
	}
	return lines
}

// wrapGlyphs breaks line at any character so that each part fits width.
func wrapGlyphs(line []Span, width int) [][]Span {
	var lines [][]Span
	var current []Span
	currentWidth := 0
	for _, span := range line {
		start := 0
		clusters := uucode.NewGraphemeWidthIterator(span.Text)
		for cluster, ok := clusters.Next(); ok; cluster, ok = clusters.Next() {
			if currentWidth > 0 && currentWidth+cluster.Width > width {
				if start < cluster.Start {
					current = append(current, Span{Text: span.Text[start:cluster.Start], Style: span.Style})
				}
				lines = append(lines, current)
				current, currentWidth, start = nil, 0, cluster.Start
			}
			currentWidth += cluster.Width
		}
		if start < len(span.Text) {
			current = append(current, Span{Text: span.Text[start:], Style: span.Style})
		}
	}
	return append(lines, current)
}

// wrapWords breaks line between words so that each part fits width, keeping the style of each span. The spaces that hang past the end of a part are dropped.
func wrapWords(line []Span, width int, glyph bool) [][]Span {
	var plain strings.Builder
	for _, span := range line {
		plain.WriteString(span.Text)
	}

	var lines [][]Span
	index, offset := 0, 0 // The span and byte within it where the next line starts.
	pieces := wordwrap.Split(plain.String(), width, glyph)
	for i, piece := range pieces {
		var current []Span
		for n := len(piece); n > 0 && index < len(line); {
			span := line[index]
			take := min(n, len(span.Text)-offset)
			current = append(current, Span{Text: span.Text[offset : offset+take], Style: span.Style})
			n -= take
			if offset += take; offset == len(span.Text) {
				index, offset = index+1, 0
			}
		}
		// The line breaks that end a piece are dropped from it, so they are skipped here too.
		for index < len(line) && offset < len(line[index].Text) && line[index].Text[offset] == '\r' {
			if offset++; offset == len(line[index].Text) {
				index, offset = index+1, 0
			}
		}
		if i < len(pieces)-1 {
			current = trimSpaces(current)
		}
		lines = append(lines, current)
	}
	return lines
}

// trimSpaces returns line without the spaces at its end.
func trimSpaces(line []Span) []Span {
	for len(line) > 0 {
		last := &line[len(line)-1]
		if last.Text = strings.TrimRight(last.Text, " "); last.Text != "" {
			break
		}
		line = line[:len(line)-1]
	}
	return line
}
