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

// Line is a list of styled segments.
type Line []Segment

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
