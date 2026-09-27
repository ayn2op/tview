// Package tabs provides a row of tab labels above the content of the active tab.
package tabs

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/keybind"
	"github.com/gdamore/tcell/v3"
)

// Widget draws the tab labels, centered on the first row, above the content of the active tab. The model owns the tabs and switches between them.
type Widget struct {
	labels   []string
	active   int
	content  tview.Element
	keybinds Keybinds
	onSelect func(index int) tview.Msg
}

var _ tview.Element = Widget{}

// New returns tabs with labels.
func New(labels ...string) Widget {
	return Widget{labels: labels, keybinds: defaultKeybinds}
}

// Active sets the index of the active tab.
func (w Widget) Active(index int) Widget {
	w.active = index
	return w
}

// Content sets the element of the active tab.
func (w Widget) Content(content tview.Element) Widget {
	w.content = content
	return w
}

// Keybinds sets the keys that switch tabs.
func (w Widget) Keybinds(keybinds Keybinds) Widget {
	w.keybinds = keybinds
	return w
}

// OnSelect sets the function that turns the index of a tab to switch to into a message.
func (w Widget) OnSelect(onSelect func(index int) tview.Msg) Widget {
	w.onSelect = onSelect
	return w
}

// Draw draws the labels on the first row of area, highlighting the active one, and the content below them.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	x := area.X + w.stripOffset(area.Width)
	for i, label := range w.labels {
		style := tcell.StyleDefault
		if i == w.active {
			style = style.Reverse(true)
		}
		tview.Print(screen, label, x, area.Y, len(label), tview.AlignmentLeft, style)
		x += len(label) + 1
	}
	if w.content != nil {
		w.content.Draw(screen, contentArea(area))
	}
}

// Handle turns the tab keybinds and clicks and scrolling on the labels into the OnSelect message, and passes other messages to the content below the labels.
func (w Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	switch msg := msg.(type) {
	case tview.KeyMsg:
		switch {
		case keybind.Matches(msg, w.keybinds.Previous) && w.active > 0:
			return w.selectTab(w.active - 1)
		case keybind.Matches(msg, w.keybinds.Next) && w.active < len(w.labels)-1:
			return w.selectTab(w.active + 1)
		}
	case tview.MouseMsg:
		x, y := msg.Position()
		if tab, ok := w.tabAt(area, x, y); ok {
			switch msg.Action {
			case tview.MouseLeftClick:
				return w.selectTab(tab)
			case tview.MouseScrollUp, tview.MouseScrollLeft:
				return w.selectTab(max(w.active-1, 0))
			case tview.MouseScrollDown, tview.MouseScrollRight:
				return w.selectTab(min(w.active+1, len(w.labels)-1))
			}
			return nil
		}
	}
	if w.content == nil {
		return msg
	}
	return w.content.Handle(msg, contentArea(area))
}

// selectTab returns the OnSelect message for index, or nil if index is already active.
func (w Widget) selectTab(index int) tview.Msg {
	if index == w.active || w.onSelect == nil {
		return nil
	}
	return w.onSelect(index)
}

// contentArea returns the area of the content, below the labels.
func contentArea(area tview.Rectangle) tview.Rectangle {
	area.Y++
	area.Height = max(area.Height-1, 0)
	return area
}

// tabAt returns the tab whose label is at x, y in area, whose first row holds the labels.
func (w Widget) tabAt(area tview.Rectangle, x, y int) (int, bool) {
	if y != area.Y || x < area.X || x >= area.X+area.Width {
		return 0, false
	}
	start := area.X + w.stripOffset(area.Width)
	for i, label := range w.labels {
		if x >= start && x < start+len(label) {
			return i, true
		}
		start += len(label) + 1
	}
	return 0, false
}

// stripOffset returns where the labels, separated by a single space, start so that they are centered as a group within width.
func (w Widget) stripOffset(width int) int {
	stripWidth := -1 // no trailing space after the last label
	for _, label := range w.labels {
		stripWidth += len(label) + 1
	}
	return max((width-stripWidth)/2, 0)
}
