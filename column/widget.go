// Package column lays out widgets top to bottom.
package column

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/flex"
	"github.com/ayn2op/tview/layout"
)

// Widget lays out its children top to bottom.
type Widget struct{ flex flex.Widget }

var _ tview.Widget = Widget{}

// New returns a column of children, skipping nil ones. It fills its parent in both directions by default.
func New(children ...tview.Widget) Widget {
	return Widget{flex.New(false, children...)}
}

// Push adds child after the others, unless it is nil.
func (w Widget) Push(child tview.Widget) Widget {
	return Widget{w.flex.Push(child)}
}

// Width sets the width of the column.
func (w Widget) Width(width layout.Length) Widget {
	return Widget{w.flex.Width(width)}
}

// Height sets the height of the column.
func (w Widget) Height(height layout.Length) Widget {
	return Widget{w.flex.Height(height)}
}

// Spacing sets the number of cells between children.
func (w Widget) Spacing(spacing int) Widget {
	return Widget{w.flex.Spacing(spacing)}
}

// AlignX sets where the children narrower than the column go: at its left unless set.
func (w Widget) AlignX(align tview.HorizontalAlignment) Widget {
	return Widget{w.flex.Align(flex.FromHorizontal(align))}
}

// Wrap returns the column starting a new line where the next child would overflow it, with as many columns between lines as there are cells between children.
func (w Widget) Wrap() Wrapping {
	return Wrapping{w.flex.Wrap()}
}

// Size returns the width and height of the column.
func (w Widget) Size() (width, height layout.Length) { return w.flex.Size() }

// Layout returns the size of the column within limits.
func (w Widget) Layout(limits layout.Limits) layout.Size { return w.flex.Layout(limits) }

func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) { w.flex.Draw(screen, area) }

// Handle passes msg through each child in turn, with the child's area. It stops when a child drops the message.
func (w Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	return w.flex.Handle(msg, area)
}

// Wrapping is a column that starts a new line where the next child would overflow it.
// Within a line, the children narrower than it go where the column's AlignX puts them.
type Wrapping struct{ flex flex.Wrapping }

var _ tview.Widget = Wrapping{}

// HorizontalSpacing sets the number of columns between lines.
func (w Wrapping) HorizontalSpacing(spacing int) Wrapping {
	return Wrapping{w.flex.LineSpacing(spacing)}
}

// AlignY sets where each line goes within the tallest: at its top unless set.
func (w Wrapping) AlignY(align tview.VerticalAlignment) Wrapping {
	return Wrapping{w.flex.AlignLines(flex.FromVertical(align))}
}

// Size returns the width and height of the column.
func (w Wrapping) Size() (width, height layout.Length) { return w.flex.Size() }

// Layout returns the size of the column within limits.
func (w Wrapping) Layout(limits layout.Limits) layout.Size { return w.flex.Layout(limits) }

func (w Wrapping) Draw(screen tview.Screen, area tview.Rectangle) { w.flex.Draw(screen, area) }

// Handle passes msg through each child in turn, with the child's area. It stops when a child drops the message.
func (w Wrapping) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	return w.flex.Handle(msg, area)
}
