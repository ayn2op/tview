// Package picker provides a declarative element that filters a list of items with a query.
package picker

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/box"
	"github.com/ayn2op/tview/column"
	"github.com/ayn2op/tview/layout"
	"github.com/ayn2op/tview/list"
	"github.com/ayn2op/tview/row"
	"github.com/ayn2op/tview/scrollbar"
	"github.com/ayn2op/tview/text"
	"github.com/ayn2op/tview/textinput"
	"github.com/gdamore/tcell/v3"
	"github.com/sahilm/fuzzy"
)

// bottom border + value
const inputHeight = 2

// Widget is a query above the list of items that match it.
type Widget struct {
	items               Items
	searchState         *SearchState
	keybind             func(tview.KeyMsg) (Action, bool)
	listKeybind         func(tview.KeyMsg) (list.Action, bool)
	scrollBar           scrollbar.Widget
	scrollBarVisibility list.ScrollBarVisibility
	onChange            func(Change) tview.Msg
	onSelect            func(Item) tview.Msg
	onCancel            tview.Msg
}

var _ tview.Element = Widget{}

// New returns a picker of items, with searchState as its query, the matches, and the selection.
func New(items Items, searchState *SearchState) Widget {
	return Widget{items: items, searchState: searchState, keybind: DefaultKeybind, listKeybind: list.DefaultKeybind, scrollBar: scrollbar.New()}
}

// Keybind sets the function that turns keys into Actions, DefaultKeybind unless set.
func (w Widget) Keybind(f func(tview.KeyMsg) (Action, bool)) Widget {
	w.keybind = f
	return w
}

// ListKeybind sets the function that turns keys into the list's Actions, list.DefaultKeybind unless set.
func (w Widget) ListKeybind(f func(tview.KeyMsg) (list.Action, bool)) Widget {
	w.listKeybind = f
	return w
}

// ScrollBar sets the list's scroll bar and when it is shown.
func (w Widget) ScrollBar(scrollBar scrollbar.Widget, visibility list.ScrollBarVisibility) Widget {
	w.scrollBar, w.scrollBarVisibility = scrollBar, visibility
	return w
}

// OnChange sets the function that turns a change to the search state into a message.
func (w Widget) OnChange(onChange func(Change) tview.Msg) Widget {
	w.onChange = onChange
	return w
}

// OnSelect sets the function that turns the selected item into a message.
func (w Widget) OnSelect(onSelect func(Item) tview.Msg) Widget {
	w.onSelect = onSelect
	return w
}

// OnCancel sets the message the picker returns when it is cancelled.
func (w Widget) OnCancel(msg tview.Msg) Widget {
	w.onCancel = msg
	return w
}

// Size returns Fill, as the picker takes its whole area.
func (Widget) Size() (width, height layout.Length) {
	return layout.Fill, layout.Fill
}

// Layout returns the size of limits, as the picker takes its whole area.
func (Widget) Layout(limits layout.Limits) layout.Size {
	return layout.Atomic(limits, layout.Fill, layout.Fill)
}

// Draw draws the query above the list.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	w.layout().Draw(screen, area)
}

// Handle returns the OnSelect and OnCancel messages for their Actions, sends the list's keys to the list, and passes other messages to the query and the list.
func (w Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	if key, ok := msg.(tview.KeyMsg); ok {
		action, ok := w.keybind(key)
		switch {
		case ok && action == ActionSelect:
			if item, ok := w.searchState.Selected(w.items); ok && w.onSelect != nil {
				return w.onSelect(item)
			}
			return nil
		case ok && action == ActionCancel:
			return w.onCancel
		}
		if _, ok := w.listKeybind(key); ok {
			area.Y, area.Height = area.Y+inputHeight, max(area.Height-inputHeight, 0)
			return w.listView().Handle(msg, area)
		}
	}
	return w.layout().Handle(msg, area)
}

func (w Widget) layout() tview.Element {
	query := textinput.New(&w.searchState.query).
		Focused(true).
		OnChange(w.queryChange)
	// A line below the query separates it from the list.
	var line tview.BorderSet
	line.Bottom = tview.BoxDrawingsLightHorizontal
	line.BottomLeft, line.BottomRight = line.Bottom, line.Bottom
	header := box.New(row.New(text.New("> "), query)).Borders(tview.BordersBottom).BorderSet(line).BorderStyle(tcell.StyleDefault.Dim(true))
	return column.New(column.New(header).Height(layout.Fixed(inputHeight)), w.listView())
}

func (w Widget) listView() list.Widget {
	s, items := w.searchState, w.items
	return list.New(s.list, s.count(items), func(i int) tview.Element { return text.New(items[s.index(i)].Text) }).
		SelectedStyle(tcell.StyleDefault.Reverse(true)).
		ScrollBar(w.scrollBar, w.scrollBarVisibility).
		Keybind(w.listKeybind).
		Focused(true).
		OnChange(func(a list.Change) tview.Msg { return w.change(Change{list: a}) })
}

// queryChange turns an edit of the query into a Change, matching the items against the query if the edit changed it.
func (w Widget) queryChange(a textinput.Change) tview.Msg {
	query := w.searchState.query
	before := query.Value()
	query.Apply(a)
	change := Change{query: a, isQuery: true}
	if value := query.Value(); value != before {
		count := len(w.items)
		if value != "" {
			for _, match := range fuzzy.FindFrom(value, w.items) {
				change.matches = append(change.matches, match.Index)
			}
			count = len(change.matches)
		}
		change.filtered, change.cursor = true, min(0, count-1)
	}
	return w.change(change)
}

func (w Widget) change(a Change) tview.Msg {
	if w.onChange == nil {
		return nil
	}
	return w.onChange(a)
}
