package tview

import "github.com/ayn2op/tview/layout"

// Length and Size are defined in the layout package, which the elements and this package share.
type (
	Length = layout.Length
	Size   = layout.Size
)

// Fill takes an equal share of the free space, and Shrink the size of the element's content.
var (
	Fill   = layout.Fill
	Shrink = layout.Shrink
)

// FillPortion takes a share of the free space in proportion to portion, where Fill is a portion of 1.
func FillPortion(portion int) Length {
	return layout.FillPortion(portion)
}

// Fixed takes exactly cells cells.
func Fixed(cells int) Length {
	return layout.Fixed(cells)
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
