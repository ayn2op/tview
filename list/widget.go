// Package list shows a list of items of varying height with a cursor.
package list

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/clip"
	"github.com/ayn2op/tview/layout"
	"github.com/gdamore/tcell/v3"
)

// Widget draws items, with the one its SelectionState selects in the selected style, and turns keys and the mouse into Changes.
type Widget struct {
	selectionState SelectionState
	count          int
	item           func(index int) tview.Widget
	width, height  layout.Length
	gap            int
	selectedStyle  tview.Style
	keybind        func(tview.KeyMsg) Action
	focused        bool
	onChange       func(Change) tview.Msg
}

var _ tview.Widget = Widget{}

// New returns a list of count items built by item, with selectionState as its cursor.
func New(selectionState SelectionState, count int, item func(index int) tview.Widget) Widget {
	return Widget{
		selectionState: selectionState,
		count:          count,
		item:           item,
		width:          layout.Fill,
		height:         layout.Fill,
		selectedStyle:  tcell.StyleDefault.Reverse(true),
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

// Keybind sets the function that turns keys into Actions.
func (w Widget) Keybind(f func(tview.KeyMsg) Action) Widget {
	w.keybind = f
	return w
}

// Focused sets whether the list receives keys.
func (w Widget) Focused(focused bool) Widget {
	w.focused = focused
	return w
}

// OnChange makes the list interactive, with the message f returns for the Change.
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
	return layout.Sized(limits, w.width, w.height, func(limits layout.Limits) layout.Size {
		_, _, total := w.layout(limits.Max.Width)
		return layout.Size{Width: limits.Max.Width, Height: total}
	})
}

// layout returns where each item starts and how many rows it takes at a width.
func (w Widget) layout(width int) (starts, sizes []int, total int) {
	limits := layout.Limits{Max: layout.Size{Width: width}, Infinite: layout.Axes{Height: true}}
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

// Target returns the rows of the item selected last at a width, or no rows for none.
func (w Widget) Target(width int) (top, height int) {
	i := w.selectionState.target
	if i < 0 || i >= w.count {
		return 0, 0
	}
	starts, sizes, _ := w.layout(width)
	return starts[i], sizes[i]
}

// Draw draws the items, with the selected one in the selected style.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	starts, sizes, _ := w.layout(area.Width)
	cursor := min(w.selectionState.cursor, w.count-1)
	clipped := &clip.Screen{Screen: screen, Area: area}
	// Items off the screen, as in a viewport, are skipped.
	_, bottom := screen.Size()
	for i := range w.count {
		itemArea := tview.Rectangle{X: area.X, Y: area.Y + starts[i], Width: area.Width, Height: sizes[i]}
		if itemArea.Y >= bottom {
			break
		}
		if itemArea.Y+itemArea.Height <= 0 {
			continue
		}
		if i != cursor || w.selectedStyle == tcell.StyleDefault {
			w.item(i).Draw(clipped, itemArea)
			continue
		}
		styled := &styledScreen{Screen: clipped, style: w.selectedStyle}
		styled.FillArea(itemArea.X, itemArea.Y, itemArea.Width, itemArea.Height, ' ', tcell.StyleDefault)
		w.item(i).Draw(styled, itemArea)
	}
}

// Handle turns keys and ActionMsgs (while focused) and clicks on items into a Change once OnChange is set. Other messages pass through unchanged.
func (w Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	if w.onChange == nil {
		return msg
	}
	if key, ok := msg.(tview.KeyMsg); ok && w.focused {
		if action := w.keybind(key); action != ActionNone {
			msg = ActionMsg(action)
		}
	}
	cursor := min(w.selectionState.cursor, w.count-1)
	switch m := msg.(type) {
	case ActionMsg:
		if !w.focused {
			return msg
		}
		switch Action(m) {
		case ActionSelectDown:
			cursor = min(cursor+1, w.count-1)
		case ActionSelectUp:
			cursor = max(cursor-1, min(0, w.count-1))
		case ActionSelectTop:
			cursor = min(0, w.count-1)
		case ActionSelectBottom:
			cursor = w.count - 1
		default:
			return msg
		}
	case tview.MouseMsg:
		x, y := m.Position()
		if m.Action != tview.MouseLeftDown || !area.Contains(x, y) {
			return msg
		}
		starts, sizes, _ := w.layout(area.Width)
		for i := range w.count {
			if row := y - area.Y; row >= starts[i] && row < starts[i]+sizes[i] {
				cursor = i
			}
		}
	default:
		return msg
	}
	return w.onChange(Change{cursor: cursor})
}
