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
