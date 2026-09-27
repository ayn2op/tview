package tview

// Length is how much space an element takes along one axis of its parent.
type Length struct {
	cells, portion int
	shrink         bool
}

// Fill takes an equal share of the free space. It is the default for every element.
var Fill = Length{portion: 1}

// FillPortion takes a share of the free space in proportion to portion, where Fill is a portion of 1.
func FillPortion(portion int) Length {
	return Length{portion: portion}
}

// Fixed takes exactly cells cells.
func Fixed(cells int) Length {
	return Length{cells: cells}
}

// Shrink takes the size of the element's content.
var Shrink = Length{shrink: true}

// Cells returns the size of a Fixed length, and 0 for others.
func (l Length) Cells() int {
	return l.cells
}

// Portion returns the portion of a Fill or FillPortion length, and 0 for others.
func (l Length) Portion() int {
	return l.portion
}

// IsShrink reports whether l is Shrink.
func (l Length) IsShrink() bool {
	return l.shrink
}

// Sizer is implemented by elements with a width and height. An element resolves Shrink to the Fixed size of its content.
type Sizer interface {
	Size() (width, height Length)
}

// SizeOf returns the width and height of element, which are Fill unless it is a Sizer.
func SizeOf(element Element) (width, height Length) {
	if s, ok := element.(Sizer); ok {
		return s.Size()
	}
	return Fill, Fill
}
