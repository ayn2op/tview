// Package picker provides a declarative element that filters a list of items with a query.
package picker

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/box"
	"github.com/ayn2op/tview/column"
	"github.com/ayn2op/tview/keybind"
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
	keybinds            Keybinds
	scrollBar           scrollbar.Widget
	scrollBarVisibility list.ScrollBarVisibility
	onAction            func(Action) tview.Msg
	onSelect            func(Item) tview.Msg
	onCancel            tview.Msg
}

var _ tview.Element = Widget{}

// New returns a picker of items, with searchState as its query, the matches, and the selection.
func New(items Items, searchState *SearchState) Widget {
	return Widget{items: items, searchState: searchState, keybinds: defaultKeybinds, scrollBar: scrollbar.New()}
}

// Keybinds sets the keybinds of the picker.
func (w Widget) Keybinds(keybinds Keybinds) Widget {
	w.keybinds = keybinds
	return w
}

// ScrollBar sets the list's scroll bar and when it is shown.
func (w Widget) ScrollBar(scrollBar scrollbar.Widget, visibility list.ScrollBarVisibility) Widget {
	w.scrollBar, w.scrollBarVisibility = scrollBar, visibility
	return w
}

// OnAction sets the function that turns a change to the search state into a message.
func (w Widget) OnAction(onAction func(Action) tview.Msg) Widget {
	w.onAction = onAction
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

// Draw draws the query above the list.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	w.layout().Draw(screen, area)
}

// Handle returns the OnSelect and OnCancel messages for their keybinds, sends the list's keybinds to the list, and passes other messages to the query and the list.
func (w Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	if key, ok := msg.(tview.KeyMsg); ok {
		switch {
		case keybind.Matches(key, w.keybinds.Select):
			if item, ok := w.searchState.Selected(w.items); ok && w.onSelect != nil {
				return w.onSelect(item)
			}
			return nil
		case keybind.Matches(key, w.keybinds.Cancel):
			return w.onCancel
		case keybind.Matches(key, w.keybinds.moves()...):
			area.Y, area.Height = area.Y+inputHeight, max(area.Height-inputHeight, 0)
			return w.listView().Handle(msg, area)
		}
	}
	return w.layout().Handle(msg, area)
}

func (w Widget) layout() tview.Element {
	query := textinput.New(&w.searchState.query).
		Focused(true).
		OnAction(w.queryAction)
	// A line below the query separates it from the list.
	var line tview.BorderSet
	line.Bottom = tview.BoxDrawingsLightHorizontal
	line.BottomLeft, line.BottomRight = line.Bottom, line.Bottom
	header := box.New(row.New(text.New("> "), query)).Borders(tview.BordersBottom).BorderSet(line).BorderStyle(tcell.StyleDefault.Dim(true))
	return column.New(column.New(header).Height(tview.Fixed(inputHeight)), w.listView())
}

func (w Widget) listView() list.Widget {
	s, items := w.searchState, w.items
	return list.New(&s.list, s.count(items), func(i int) list.Item { return entry(items[s.index(i)].Text) }).
		SelectedStyle(tcell.StyleDefault.Reverse(true)).
		ScrollBar(w.scrollBar, w.scrollBarVisibility).
		Keybinds(w.keybinds.Keybinds).
		Focused(true).
		OnAction(func(a list.Action) tview.Msg { return w.action(Action{list: a}) })
}

// queryAction turns an edit of the query into an Action, matching the items against the query if the edit changed it.
func (w Widget) queryAction(a textinput.Action) tview.Msg {
	query := w.searchState.query
	before := query.Value()
	query.Perform(a)
	action := Action{query: a, isQuery: true}
	if value := query.Value(); value != before {
		count := len(w.items)
		if value != "" {
			for _, match := range fuzzy.FindFrom(value, w.items) {
				action.matches = append(action.matches, match.Index)
			}
			count = len(action.matches)
		}
		action.filtered, action.cursor = true, min(0, count-1)
	}
	return w.action(action)
}

func (w Widget) action(a Action) tview.Msg {
	if w.onAction == nil {
		return nil
	}
	return w.onAction(a)
}
