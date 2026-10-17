// Package row lays out widgets left to right.
package row

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/flex"
	"github.com/ayn2op/tview/layout"
)

// Widget lays out its children left to right.
type Widget struct{ flex flex.Widget }

var _ tview.Widget = Widget{}

// New returns a row of children, skipping nil ones. It fills its parent in both directions by default.
func New(children ...tview.Widget) Widget {
	return Widget{flex.New(true, children...)}
}

// Push adds child after the others, unless it is nil.
func (w Widget) Push(child tview.Widget) Widget {
	return Widget{w.flex.Push(child)}
}

// Width sets the width of the row.
func (w Widget) Width(width layout.Length) Widget {
	return Widget{w.flex.Width(width)}
}

// Height sets the height of the row.
func (w Widget) Height(height layout.Length) Widget {
	return Widget{w.flex.Height(height)}
}

// Spacing sets the number of cells between children.
func (w Widget) Spacing(spacing int) Widget {
	return Widget{w.flex.Spacing(spacing)}
}

// AlignY sets where the children shorter than the row go: at its top unless set.
func (w Widget) AlignY(alignment tview.Alignment) Widget {
	return Widget{w.flex.Align(alignment)}
}

// Wrap returns the row starting a new line where the next child would overflow it, with as many rows between lines as there are cells between children.
func (w Widget) Wrap() Wrapping {
	return Wrapping{w.flex.Wrap()}
}

// Size returns the width and height of the row.
func (w Widget) Size() (width, height layout.Length) { return w.flex.Size() }

// Layout returns the size of the row within limits.
func (w Widget) Layout(limits layout.Limits) layout.Size { return w.flex.Layout(limits) }

func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) { w.flex.Draw(screen, area) }

// Handle passes msg through each child in turn, with the child's area. It stops when a child drops the message.
func (w Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	return w.flex.Handle(msg, area)
}

// Wrapping is a row that starts a new line where the next child would overflow it.
// Within a line, the children shorter than it go where the row's AlignY puts them.
type Wrapping struct{ flex flex.Wrapping }

var _ tview.Widget = Wrapping{}

// VerticalSpacing sets the number of rows between lines.
func (w Wrapping) VerticalSpacing(spacing int) Wrapping {
	return Wrapping{w.flex.LineSpacing(spacing)}
}

// AlignX sets where each line goes within the widest: at its left unless set.
func (w Wrapping) AlignX(alignment tview.Alignment) Wrapping {
	return Wrapping{w.flex.AlignLines(alignment)}
}

// Size returns the width and height of the row.
func (w Wrapping) Size() (width, height layout.Length) { return w.flex.Size() }

// Layout returns the size of the row within limits.
func (w Wrapping) Layout(limits layout.Limits) layout.Size { return w.flex.Layout(limits) }

func (w Wrapping) Draw(screen tview.Screen, area tview.Rectangle) { w.flex.Draw(screen, area) }

// Handle passes msg through each child in turn, with the child's area. It stops when a child drops the message.
func (w Wrapping) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	return w.flex.Handle(msg, area)
}
