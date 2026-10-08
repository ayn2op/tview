// Package viewport scrolls a child that is larger than its area.
package viewport

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/clip"
	"github.com/ayn2op/tview/layout"
	"github.com/ayn2op/tview/scrollbar"
)

// ScrollBarVisibility is when a viewport shows its scroll bar.
type ScrollBarVisibility uint8

const (
	ScrollBarVisibilityNever ScrollBarVisibility = iota
	// ScrollBarVisibilityAutomatic shows the scroll bar when the child does not fit, keeping its column free when it does.
	ScrollBarVisibilityAutomatic
	ScrollBarVisibilityAlways
)

// Widget shows the part of its child that it is scrolled to.
type Widget struct {
	child         tview.Widget
	scrollState   ScrollState
	width, height layout.Length
	axes          layout.Axes
	target        func(width int) (top, height int)
	contentWidth  func(top, height int) int
	scrollBar     scrollbar.Widget
	visibility    ScrollBarVisibility
	trackEnd      bool
	keybind       func(tview.KeyMsg) Action
	focused       bool
	onChange      func(Change) tview.Msg
	onChildMsg    func(Change, tview.Msg) tview.Msg
}

var _ tview.Widget = Widget{}

// New returns a viewport that scrolls child vertically.
func New(child tview.Widget, scrollState ScrollState) Widget {
	return Widget{
		child:       child,
		scrollState: scrollState,
		width:       layout.Fill,
		height:      layout.Fill,
		axes:        layout.Axes{Height: true},
		scrollBar:   scrollbar.New(),
		keybind:     DefaultKeybind,
	}
}

// Width sets the width of the viewport.
func (w Widget) Width(width layout.Length) Widget {
	w.width = width
	return w
}

// Height sets the height of the viewport.
func (w Widget) Height(height layout.Length) Widget {
	w.height = height
	return w
}

// Axes sets the axes the child scrolls along.
func (w Widget) Axes(axes layout.Axes) Widget {
	w.axes = axes
	return w
}

// Target sets the rows of the child at a width that ScrollState.ScrollToTarget scrolls to, such as the Target of a tree or list.
func (w Widget) Target(f func(width int) (top, height int)) Widget {
	w.target = f
	return w
}

// ContentWidth sets the width of the rows of the child in view, such as the RowsWidth of a tree, so that it scrolls no further sideways than they reach.
func (w Widget) ContentWidth(f func(top, height int) int) Widget {
	w.contentWidth = f
	return w
}

// ScrollBar sets the vertical scroll bar and when it is shown.
func (w Widget) ScrollBar(scrollBar scrollbar.Widget, visibility ScrollBarVisibility) Widget {
	w.scrollBar, w.visibility = scrollBar, visibility
	return w
}

// TrackEnd sets whether the viewport stays at the last row as the child grows while it is there.
func (w Widget) TrackEnd(track bool) Widget {
	w.trackEnd = track
	return w
}

// Keybind sets the function that turns keys into Actions.
func (w Widget) Keybind(f func(tview.KeyMsg) Action) Widget {
	w.keybind = f
	return w
}

// Focused sets whether the viewport receives keys.
func (w Widget) Focused(focused bool) Widget {
	w.focused = focused
	return w
}

// OnChange makes the viewport scroll, with the message f returns for the Change.
func (w Widget) OnChange(f func(Change) tview.Msg) Widget {
	w.onChange = f
	return w
}

// OnChildMsg sets the function that joins a message of the child with the Change that saves the position shown, when the viewport is not scrolled to it, as while it keeps its target in view.
func (w Widget) OnChildMsg(f func(Change, tview.Msg) tview.Msg) Widget {
	w.onChildMsg = f
	return w
}

// Size returns the width and height of the viewport.
func (w Widget) Size() (width, height layout.Length) {
	return w.width, w.height
}

// Layout returns the size of the viewport within limits.
func (w Widget) Layout(limits layout.Limits) layout.Size {
	return layout.Sized(limits, w.width, w.height, func(limits layout.Limits) layout.Size {
		bar := layout.Size{}
		if w.visibility != ScrollBarVisibilityNever {
			bar.Width = 1
		}
		content := w.child.Layout(limits.Shrink(bar))
		return layout.Size{Width: content.Width + bar.Width, Height: content.Height}
	})
}

// view is the child laid out in an area.
type view struct {
	// port is where the child shows, and bar the scroll bar if it is shown.
	port, bar tview.Rectangle
	content   layout.Size
	x, y      int
}

func (w Widget) resolve(area tview.Rectangle) view {
	v := view{port: area}
	// The scroll bar column is reserved even while the bar is hidden, so the child is always laid out at the same width.
	if area.Width > 1 && w.visibility != ScrollBarVisibilityNever {
		v.port.Width--
	}
	v.content = w.child.Layout(layout.Limits{Max: v.port.Size(), Infinite: w.axes})
	if v.port.Width < area.Width && (w.visibility == ScrollBarVisibilityAlways || v.content.Height > area.Height) {
		v.bar = tview.Rectangle{X: area.X + v.port.Width, Y: area.Y, Width: 1, Height: area.Height}
	}

	v.x, v.y = w.scrollState.x, w.scrollState.y
	switch {
	case w.scrollState.atEnd:
		_, v.y = v.max()
	case w.scrollState.toTarget && w.target != nil:
		if top, height := w.target(v.content.Width); height > 0 {
			v.y = top + height/2 - v.port.Height/2
		}
	}
	return w.clamp(v, v.x, v.y)
}

// clamp returns v scrolled to x and y, limited to what can be shown.
func (w Widget) clamp(v view, x, y int) view {
	maxX, maxY := v.max()
	v.y = min(max(y, 0), maxY)
	if w.contentWidth != nil {
		v.content.Width = max(w.contentWidth(v.y, v.port.Height), v.port.Width)
		maxX, _ = v.max()
	}
	v.x = min(max(x, 0), maxX)
	return v
}

// max returns the furthest the child scrolls.
func (v view) max() (x, y int) {
	return max(v.content.Width-v.port.Width, 0), max(v.content.Height-v.port.Height, 0)
}

// child returns the area of the child, which the viewport shows a part of.
func (v view) child() tview.Rectangle {
	return tview.Rectangle{X: v.port.X - v.x, Y: v.port.Y - v.y, Width: v.content.Width, Height: v.content.Height}
}

// Draw draws the part of the child that is scrolled into area, and the scroll bar.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	v := w.resolve(area)
	w.child.Draw(&clip.Screen{Screen: screen, Area: v.port}, v.child())
	if v.bar.Width > 0 {
		w.bar(v).Draw(screen, v.bar)
	}
}

// Handle passes msg to the child and turns what it leaves into a Change.
func (w Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	switch msg.(type) {
	case tview.KeyMsg, tview.MouseMsg:
	default:
		// Laying the child out is the costly part, so messages that are not input skip it.
		return w.child.Handle(msg, area)
	}
	v := w.resolve(area)
	a := Change{x: v.x, y: v.y, dragging: w.scrollState.dragging, grab: w.scrollState.grab}
	m, isMouse := msg.(tview.MouseMsg)
	if isMouse && w.onChange != nil && w.barMouse(m, v, &a) {
		return w.change(v, a)
	}
	if !isMouse || v.port.Contains(m.Position()) {
		received := msg
		if msg = w.child.Handle(msg, v.child()); msg == nil {
			return nil
		}
		if w.onChildMsg != nil && msg != received && v.x != w.scrollState.x {
			a.toTarget, a.atEnd = w.scrollState.toTarget, w.scrollState.atEnd
			return w.onChildMsg(a, msg)
		}
	}
	if w.onChange == nil {
		return msg
	}
	switch m := msg.(type) {
	case tview.KeyMsg:
		if !w.focused {
			return msg
		}
		switch w.keybind(m) {
		case ActionUp:
			a.y--
		case ActionDown:
			a.y++
		case ActionLeft:
			a.x--
		case ActionRight:
			a.x++
		case ActionTop:
			a.y = 0
		case ActionBottom:
			_, a.y = v.max()
		case ActionPageUp:
			a.y -= v.port.Height
		case ActionPageDown:
			a.y += v.port.Height
		default:
			return msg
		}
	case tview.MouseMsg:
		if !v.port.Contains(m.Position()) {
			return msg
		}
		switch m.Action {
		case tview.MouseScrollUp:
			a.y--
		case tview.MouseScrollDown:
			a.y++
		case tview.MouseScrollLeft:
			a.x -= v.port.Width / 2
		case tview.MouseScrollRight:
			a.x += v.port.Width / 2
		default:
			return msg
		}
	default:
		return msg
	}
	return w.change(v, a)
}

func (w Widget) change(v view, a Change) tview.Msg {
	v = w.clamp(v, a.x, a.y)
	a.x, a.y = v.x, v.y
	_, maxY := v.max()
	a.atEnd = w.trackEnd && a.y == maxY
	return w.onChange(a)
}

// bar returns the scroll bar for the view.
func (w Widget) bar(v view) scrollbar.Widget {
	return w.scrollBar.Lengths(v.content.Height, v.port.Height).Offset(v.y)
}

// barMouse applies m to a if it is on the scroll bar or drags its thumb, and reports whether it was: the begin and end symbols scroll a row, the track pages, and the thumb starts a drag.
func (w Widget) barMouse(m tview.MouseMsg, v view, a *Change) bool {
	x, y := m.Position()
	row := y - v.bar.Y
	if a.dragging {
		// Dragging the thumb follows the pointer anywhere until the button is released.
		switch m.Action {
		case tview.MouseMove:
			a.y = w.thumbOffset(v, row, a.grab)
		case tview.MouseLeftUp:
			a.dragging = false
		}
		return true
	}
	if !v.bar.Contains(x, y) {
		return false
	}
	bar := w.bar(v)
	if bar.TrackCells(v.bar.Height) == 0 || m.Action != tview.MouseLeftDown {
		return true
	}
	if bar.HasBegin() {
		row--
	}
	thumbStart, thumbSize := bar.Thumb(v.bar.Height)
	switch {
	case row < 0:
		a.y--
	case row >= bar.TrackCells(v.bar.Height):
		a.y++
	case row < thumbStart:
		a.y -= v.port.Height
	case row >= thumbStart+thumbSize:
		a.y += v.port.Height
	default:
		a.dragging, a.grab = true, row-thumbStart
	}
	return true
}

// thumbOffset returns the scroll position for the thumb grabbed at grab when the pointer is at row of the scroll bar.
func (w Widget) thumbOffset(v view, row, grab int) int {
	bar := w.bar(v)
	if bar.HasBegin() {
		row--
	}
	cells := bar.TrackCells(v.bar.Height)
	thumbStart, thumbSize := bar.Thumb(v.bar.Height)
	travel := cells - thumbSize
	start := min(max(row-grab, 0), travel)
	if travel <= 0 || start == thumbStart {
		// The thumb has not moved a cell, so the scroll position, which is finer than a cell, is kept.
		return v.y
	}
	_, maxY := v.max()
	return start * maxY / travel
}
