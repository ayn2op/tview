package text

import "slices"

// Paragraph is text laid out at a width, kept so that it is laid out again only when the spans, the wrapping or the width change.
type Paragraph struct {
	spans    []Span
	wrapping Wrapping
	// lines are laid out at width, the widest of them longest cells wide.
	lines          [][]Span
	width, longest int
}

// update returns the lines of spans within width, laying them out again only if something changed.
func (p *Paragraph) update(spans []Span, width int, wrapping Wrapping) [][]Span {
	// Lines are the same at any narrower width they still fit, such as the width of their layout that the text is drawn at.
	if p.lines != nil && p.wrapping == wrapping && p.longest <= width && width <= p.width && slices.Equal(p.spans, spans) {
		return p.lines
	}

	p.spans, p.wrapping, p.width = spans, wrapping, width
	p.lines = Lines(spans, width, wrapping)
	p.longest = longest(p.lines)
	return p.lines
}
