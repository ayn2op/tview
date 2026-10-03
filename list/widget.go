// Package list shows a scrollable list of items of varying height with a cursor.
package list

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/layout"
	"github.com/ayn2op/tview/scrollbar"
	"github.com/gdamore/tcell/v3"
)

// ScrollBarVisibility is when a list shows its scroll bar.
type ScrollBarVisibility uint8

const (
	// ScrollBarVisibilityAutomatic shows the scroll bar when the items do not fit.
	ScrollBarVisibilityAutomatic ScrollBarVisibility = iota
	ScrollBarVisibilityAlways
	ScrollBarVisibilityNever
)

// Widget draws items, selected and scrolled as its SelectionState says, and turns keys and the mouse into Changes.
type Widget struct {
	selectionState SelectionState
	count          int
	item           func(index int) tview.Element
	width, height  layout.Length
	gap            int
	selectedStyle  tview.Style
	scrollBar      scrollbar.Widget
	visibility     ScrollBarVisibility
	keybind        func(tview.KeyMsg) Action
	focused        bool
	onChange       func(Change) tview.Msg
}

var _ tview.Element = Widget{}

// New returns a list of count items built by item, with selectionState as its cursor and scroll position, showing the scroll bar when they do not fit. An item is as tall as it lays out to at the width of the list, where its height is not limited.
func New(selectionState SelectionState, count int, item func(index int) tview.Element) Widget {
	return Widget{
		selectionState: selectionState,
		count:          count,
		item:           item,
		width:          layout.Fill,
		height:         layout.Fill,
		selectedStyle:  tcell.StyleDefault.Reverse(true),
		scrollBar:      scrollbar.New(),
		keybind:        DefaultKeybind,
	}
}

// Width sets the width of the list.
func (w Widget) Width(width layout.Length) Widget {
	w.width = width
	return w
}

// Height sets the height of the list.
func (w Widget) Height(height layout.Length) Widget {
	w.height = height
	return w
}

// Gap sets the number of empty rows between items.
func (w Widget) Gap(gap int) Widget {
	w.gap = gap
	return w
}

// SelectedStyle sets the style merged into the selected item.
func (w Widget) SelectedStyle(style tview.Style) Widget {
	w.selectedStyle = style
	return w
}

// ScrollBar sets the scroll bar and when it is shown.
func (w Widget) ScrollBar(scrollBar scrollbar.Widget, visibility ScrollBarVisibility) Widget {
	w.scrollBar, w.visibility = scrollBar, visibility
	return w
}

// Keybind sets the function that turns keys into Actions, or ActionNone for keys it does not bind, DefaultKeybind unless set.
func (w Widget) Keybind(f func(tview.KeyMsg) Action) Widget {
	w.keybind = f
	return w
}

// Focused sets whether the list receives keys.
func (w Widget) Focused(focused bool) Widget {
	w.focused = focused
	return w
}

// OnChange makes the list interactive, turning keys and the mouse into the message f returns for the Change, which the model applies with SelectionState.Apply.
func (w Widget) OnChange(f func(Change) tview.Msg) Widget {
	w.onChange = f
	return w
}

// Size returns the width and height of the list.
func (w Widget) Size() (width, height layout.Length) {
	return w.width, w.height
}

// Layout returns the size of the list within limits.
func (w Widget) Layout(limits layout.Limits) layout.Size {
	width, height := w.Size()
	return layout.Atomic(limits, width, height)
}

// view is the list laid out in an area.
type view struct {
	items, bar     tview.Rectangle
	starts, sizes  []int
	total          int
	cursor, offset int
}

// maxOffset returns the largest scroll position, where the last row is at the bottom.
func (v view) maxOffset() int {
	return max(v.total-v.items.Height, 0)
}

// layout returns where each item starts and how many rows it takes in a view of a size. Items are laid out at the width of the view with an infinite height, as the list scrolls along it.
func (w Widget) layout(view layout.Size) (starts, sizes []int, total int) {
	limits := layout.Limits{Max: view, Infinite: layout.Axes{Height: true}}
	starts, sizes = make([]int, w.count), make([]int, w.count)
	for i := range w.count {
		if i > 0 {
			total += w.gap
		}
		starts[i], sizes[i] = total, w.item(i).Layout(limits).Height
		total += sizes[i]
	}
	return starts, sizes, total
}

// resolve lays the list out in area and works out the scroll position from the selection state.
func (w Widget) resolve(area tview.Rectangle) view {
	v := view{items: area}
	// Lay out beside the scroll bar first: a list long enough to need it is then laid out once.
	if area.Width > 1 && (w.visibility == ScrollBarVisibilityAlways || w.visibility == ScrollBarVisibilityAutomatic) {
		v.items.Width--
		v.starts, v.sizes, v.total = w.layout(v.items.Size())
		if w.visibility == ScrollBarVisibilityAlways || v.total > area.Height {
			v.bar = tview.Rectangle{X: area.X + v.items.Width, Y: area.Y, Width: 1, Height: area.Height}
		} else {
			v.items.Width++
			v.starts, v.sizes, v.total = w.layout(v.items.Size())
		}
	} else {
		v.starts, v.sizes, v.total = w.layout(area.Size())
	}

	c := w.selectionState
	v.cursor = min(c.cursor, w.count-1)
	v.offset = c.offset
	switch {
	case c.center >= 0 && w.count > 0:
		i := min(c.center, w.count-1)
		v.offset = v.starts[i] + v.sizes[i]/2 - area.Height/2
	case c.atEnd:
		v.offset = v.maxOffset()
	}
	v.offset = min(max(v.offset, 0), v.maxOffset())
	return v
}

// Draw draws the visible items, with the selected one in the selected style, and the scroll bar.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	v := w.resolve(area)
	clipped := &clippedScreen{Screen: screen, area: v.items}
	for i := range w.count {
		top := v.items.Y + v.starts[i] - v.offset
		if top >= v.items.Y+v.items.Height {
			break
		}
		if top+v.sizes[i] <= v.items.Y {
			continue
		}
		itemArea := tview.Rectangle{X: v.items.X, Y: top, Width: v.items.Width, Height: v.sizes[i]}
		if i != v.cursor || w.selectedStyle == tcell.StyleDefault {
			w.item(i).Draw(clipped, itemArea)
			continue
		}
		styled := &styledScreen{Screen: clipped, style: w.selectedStyle}
		styled.FillArea(itemArea.X, itemArea.Y, itemArea.Width, itemArea.Height, ' ', tcell.StyleDefault)
		w.item(i).Draw(styled, itemArea)
	}
	if v.bar.Width > 0 {
		w.bar(v).Draw(screen, v.bar)
	}
}

// Handle turns keys and ActionMsgs (while focused), the mouse wheel, clicks on items, and clicks and drags on the scroll bar into a Change once OnChange is set. Other messages pass through unchanged.
func (w Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	if w.onChange == nil {
		return msg
	}
	if key, ok := msg.(tview.KeyMsg); ok && w.focused {
		if action := w.keybind(key); action != ActionNone {
			msg = ActionMsg(action)
		}
	}
	// Laying the list out is the costly part, so messages the list ignores return before it.
	switch msg.(type) {
	case ActionMsg:
		if !w.focused {
			return msg
		}
	case tview.MouseMsg:
	default:
		return msg
	}
	v := w.resolve(area)
	a := Change{cursor: v.cursor, offset: v.offset, dragging: w.selectionState.dragging, grab: w.selectionState.grab}
	center := false
	switch m := msg.(type) {
	case ActionMsg:
		switch Action(m) {
		case ActionSelectDown:
			a.cursor, center = min(a.cursor+1, w.count-1), true
		case ActionSelectUp:
			a.cursor, center = max(a.cursor-1, min(0, w.count-1)), true
		case ActionSelectTop:
			a.cursor, center = min(0, w.count-1), true
		case ActionSelectBottom:
			a.cursor, center = w.count-1, true
		case ActionScrollDown:
			a.offset++
		case ActionScrollUp:
			a.offset--
		case ActionScrollTop:
			a.offset = 0
		case ActionScrollBottom:
			a.offset = v.maxOffset()
		default:
			return msg
		}
	case tview.MouseMsg:
		if !w.mouse(m, v, &a) {
			return msg
		}
	default:
		return msg
	}
	if center && a.cursor >= 0 {
		a.offset = v.starts[a.cursor] + v.sizes[a.cursor]/2 - v.items.Height/2
	}
	a.offset = min(max(a.offset, 0), v.maxOffset())
	a.atEnd = w.selectionState.trackEnd && a.offset == v.maxOffset()
	return w.onChange(a)
}

// mouse applies m to a and reports whether the list used it.
func (w Widget) mouse(m tview.MouseMsg, v view, a *Change) bool {
	x, y := m.Position()
	if a.dragging {
		// Dragging the thumb follows the pointer anywhere until the button is released.
		switch m.Action {
		case tview.MouseMove:
			a.offset = w.thumbOffset(v, y-v.bar.Y, a.grab)
		case tview.MouseLeftUp:
			a.dragging = false
		}
		return true
	}
	if v.bar.Contains(x, y) {
		return w.barMouse(m.Action, v, y-v.bar.Y, a)
	}
	if !v.items.Contains(x, y) {
		return false
	}
	switch m.Action {
	case tview.MouseLeftDown:
		row := y - v.items.Y + v.offset
		for i := range w.count {
			if row >= v.starts[i] && row < v.starts[i]+v.sizes[i] {
				a.cursor = i
			}
		}
	case tview.MouseScrollUp:
		a.offset--
	case tview.MouseScrollDown:
		a.offset++
	default:
		return false
	}
	return true
}

// bar returns the scroll bar for the view.
func (w Widget) bar(v view) scrollbar.Widget {
	return w.scrollBar.Lengths(v.total, v.items.Height).Offset(v.offset)
}

// barMouse applies a mouse action at row of the scroll bar to a: the begin and end symbols scroll a row, the track pages, and the thumb starts a drag.
func (w Widget) barMouse(action tview.MouseAction, v view, row int, a *Change) bool {
	bar := w.bar(v)
	if bar.TrackCells(v.bar.Height) == 0 {
		// The scroll bar is too short to be drawn.
		return true
	}
	if bar.HasBegin() {
		row--
	}
	if row < 0 || row >= bar.TrackCells(v.bar.Height) {
		// The begin or end symbol.
		if action == tview.MouseLeftDown {
			if row < 0 {
				a.offset--
			} else {
				a.offset++
			}
		}
		return true
	}
	thumbStart, thumbSize := bar.Thumb(v.bar.Height)
	onThumb := row >= thumbStart && row < thumbStart+thumbSize
	switch {
	case action == tview.MouseLeftDown && onThumb:
		a.dragging, a.grab = true, row-thumbStart
	case action == tview.MouseLeftDown && row < thumbStart:
		a.offset -= v.items.Height
	case action == tview.MouseLeftDown && !onThumb:
		a.offset += v.items.Height
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
		return v.offset
	}
	return start * v.maxOffset() / travel
}
