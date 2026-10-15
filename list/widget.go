// Package list shows a list of items of varying height with a cursor.
package list

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/clip"
	"github.com/ayn2op/tview/layout"
	"github.com/gdamore/tcell/v3"
)

// Widget draws items, with the one at the cursor in the selected style, and turns keys and the mouse into the item to move the cursor to.
type Widget struct {
	cursor int
	count  int
	item   func(index int) tview.Widget

	width, height layout.Length
	spacing       func(index int) int
	selectedStyle tview.Style

	keybind func(tview.KeyMsg) Action
	focused bool

	onChange func(cursor int) tview.Msg
}

var _ tview.Widget = Widget{}

// New returns a list of count items built by item, with the cursor on item cursor, or on none if it is negative. The model owns the cursor.
func New(cursor, count int, item func(index int) tview.Widget) Widget {
	return Widget{
		cursor:        cursor,
		count:         count,
		item:          item,
		width:         layout.Fill,
		height:        layout.Fill,
		selectedStyle: tcell.StyleDefault.Reverse(true),
		keybind:       DefaultKeybind,
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

// Spacing sets the function that returns the number of empty rows above each item after the first.
func (w Widget) Spacing(f func(index int) int) Widget {
	w.spacing = f
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

// OnChange makes the list interactive, with the message f returns for the item to move the cursor to.
func (w Widget) OnChange(f func(cursor int) tview.Msg) Widget {
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
		if i > 0 && w.spacing != nil {
			total += w.spacing(i)
		}
		starts[i], sizes[i] = total, w.item(i).Layout(limits).Height
		total += sizes[i]
	}
	return starts, sizes, total
}

// Target returns the rows of item index at a width, or no rows if there is no such item, for a viewport to keep in view.
func (w Widget) Target(index int) func(width int) (top, height int) {
	return func(width int) (top, height int) {
		if index < 0 || index >= w.count {
			return 0, 0
		}

		starts, sizes, _ := w.layout(width)
		return starts[index], sizes[index]
	}
}

// Draw draws the items, with the selected one in the selected style.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	starts, sizes, _ := w.layout(area.Width)
	cursor := min(w.cursor, w.count-1)
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

// Handle turns keys (while focused) and clicks on items into the OnChange message, unless the item clicked handles the click itself. Other messages pass through unchanged.
func (w Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	if w.onChange == nil {
		return msg
	}
	cursor := min(w.cursor, w.count-1)
	switch m := msg.(type) {
	case tview.KeyMsg:
		if !w.focused {
			return msg
		}
		switch w.keybind(m) {
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
		// A drag reaches every item wherever the pointer went, as does the release that ends it, for a slider pressed in one to follow the pointer.
		drag := m.Action == tview.MouseLeftUp || m.Action == tview.MouseMove && m.Buttons()&tcell.ButtonPrimary != 0
		if !drag && (m.Action != tview.MouseLeftDown || !area.Contains(x, y)) {
			return msg
		}
		starts, sizes, _ := w.layout(area.Width)
		for i := range w.count {
			if row := y - area.Y; drag || row >= starts[i] && row < starts[i]+sizes[i] {
				// The item pressed takes the press first, as a button or slider in it would.
				itemArea := tview.Rectangle{X: area.X, Y: area.Y + starts[i], Width: area.Width, Height: sizes[i]}
				if handled := w.item(i).Handle(msg, itemArea); handled != msg {
					return handled
				}
				cursor = i
			}
		}
		if drag {
			return msg
		}
	default:
		return msg
	}
	return w.onChange(cursor)
}
