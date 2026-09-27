package grapheme

import "github.com/rivo/uniseg"

type State struct {
	unisegState int
	boundaries  int
	length      int
}

func NewState() State {
	return State{unisegState: -1}
}

func (s State) LineBreak() (lineBreak, optional bool) {
	switch s.boundaries & uniseg.MaskLine {
	case uniseg.LineCanBreak:
		return true, true
	case uniseg.LineMustBreak:
		return true, false
	}
	return false, false
}

func (s State) Width() int  { return s.boundaries >> uniseg.ShiftWidth }
func (s State) Length() int { return s.length }

func Step(str string, state State) (cluster, rest string, next State) {
	if str == "" {
		return "", "", state
	}

	cluster, rest, state.boundaries, state.unisegState = uniseg.StepString(str, state.unisegState)
	state.length = len(cluster)
	if rest == "" && !uniseg.HasTrailingLineBreakInString(cluster) {
		state.boundaries &^= uniseg.MaskLine
	}
	return cluster, rest, state
}

// Next returns the end of the grapheme cluster starting at offset in str.
func Next(str string, offset int) int {
	cluster, _, _, _ := uniseg.FirstGraphemeClusterInString(str[offset:], -1)
	return offset + len(cluster)
}

// Previous returns the start of the grapheme cluster ending at offset in str.
func Previous(str string, offset int) int {
	start := 0
	for end := Next(str, 0); end < offset; end = Next(str, end) {
		start = end
	}
	return start
}
