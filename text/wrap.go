package text

import (
	"strings"

	"github.com/ayn2op/tview/internal/grapheme"
)

// WordWrap splits text into lines no wider than width.
func WordWrap(text string, width int) []string {
	var lines []string
	if width <= 0 {
		return lines
	}

	state := grapheme.NewState()
	lineWidth, lineLength, lastBreak, widthAtBreak := 0, 0, 0, 0
	for rest := text; rest != ""; {
		_, rest, state = grapheme.Step(rest, state)
		clusterWidth := state.Width()
		if lineWidth+clusterWidth > width {
			if widthAtBreak == 0 {
				lines = append(lines, text[:lineLength])
				text = text[lineLength:]
				lineWidth, lineLength = 0, 0
			} else {
				lines = append(lines, text[:lastBreak])
				text = text[lastBreak:]
				lineWidth -= widthAtBreak
				lineLength -= lastBreak
			}
			lastBreak, widthAtBreak = 0, 0
		}

		lineWidth += clusterWidth
		lineLength += state.Length()
		if lineBreak, optional := state.LineBreak(); lineBreak {
			if optional {
				lastBreak, widthAtBreak = lineLength, lineWidth
			} else {
				lines = append(lines, strings.TrimRight(text[:lineLength], "\n\r"))
				text = text[lineLength:]
				lineWidth, lineLength, lastBreak, widthAtBreak = 0, 0, 0, 0
			}
		}
	}
	return append(lines, text)
}

// Wrap splits a styled line to fit width.
func Wrap(line Line, width int) Text {
	if width <= 0 || len(line) == 0 {
		return Text{line}
	}

	lines := make(Text, 0, 2)
	current := make(Line, 0, len(line))
	currentWidth := 0
	flush := func() {
		lines = append(lines, current.Clone())
		current = current[:0]
		currentWidth = 0
	}

	for _, segment := range line {
		state := grapheme.NewState()
		start := 0
		for rest := segment.Text; rest != ""; {
			offset := len(segment.Text) - len(rest)
			_, rest, state = grapheme.Step(rest, state)
			end := len(segment.Text) - len(rest)
			clusterWidth := state.Width()
			if currentWidth > 0 && currentWidth+clusterWidth > width {
				current = appendSegment(current, Segment{Text: segment.Text[start:offset], Style: segment.Style})
				flush()
				start = offset
			}
			currentWidth += clusterWidth
			if currentWidth >= width {
				current = appendSegment(current, Segment{Text: segment.Text[start:end], Style: segment.Style})
				flush()
				start = end
			}
		}
		current = appendSegment(current, Segment{Text: segment.Text[start:], Style: segment.Style})
	}

	if len(current) > 0 {
		flush()
	}
	if len(lines) == 0 {
		return Text{{}}
	}
	return lines
}
