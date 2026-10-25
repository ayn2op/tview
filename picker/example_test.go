package picker_test

import (
	"fmt"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/picker"
	"github.com/ayn2op/tview/text"
	"github.com/gdamore/tcell/v3"
)

// A quick switcher that fuzzy filters channels as the query is typed. The model owns the search state and replaces it with the one the picker produces.
func ExampleNew() {
	var channels picker.Items
	for _, name := range []string{"general", "gaming", "random"} {
		channels = append(channels, picker.Item{Content: text.New(name), FilterText: name})
	}
	searchState := picker.NewSearchState()
	view := func() picker.Widget {
		return picker.New(channels, &searchState).
			OnChange(func(s *picker.SearchState) tview.Msg { return s }).
			OnSelect(func(item picker.Item) tview.Msg { return item })
	}

	area := tview.Rectangle{Width: 16, Height: 3}
	for _, str := range []string{"g", "m"} {
		if s, ok := view().Handle(tcell.NewEventKey(tcell.KeyRune, str, tcell.ModNone), area).(*picker.SearchState); ok {
			searchState = *s
		}
	}
	screentest.Print(view(), area.Width, area.Height)

	selected := view().Handle(tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone), area)
	fmt.Println(selected.(picker.Item).FilterText)
	// Output:
	// > gm
	// ────────────────
	// gaming
	// gaming
}
