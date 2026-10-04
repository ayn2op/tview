// Package layout holds the lengths and limits widgets are laid out with.
package layout

// Length is how much space a widget takes along one axis of its parent.
type Length struct {
	cells, portion int
	shrink         bool
}

// Fill takes an equal share of the free space. It is the default for every widget.
var Fill = Length{portion: 1}

// FillPortion takes a share of the free space in proportion to portion, where Fill is a portion of 1.
func FillPortion(portion int) Length {
	return Length{portion: portion}
}

// Fixed takes exactly cells cells.
func Fixed(cells int) Length {
	return Length{cells: cells}
}

// Shrink takes the size of the widget's content.
var Shrink = Length{shrink: true}

// Cells returns the size of a Fixed length, and 0 for others.
func (l Length) Cells() int {
	return l.cells
}

// Portion returns the portion of a Fill or FillPortion length, and 0 for others.
func (l Length) Portion() int {
	return l.portion
}

// IsFixed reports whether l is a Fixed length.
func (l Length) IsFixed() bool {
	return l.portion == 0 && !l.shrink
}

// IsShrink reports whether l is Shrink.
func (l Length) IsShrink() bool {
	return l.shrink
}
