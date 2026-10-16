package viewport_test

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/list"
	"github.com/ayn2op/tview/richtext"
	"github.com/ayn2op/tview/textview"
	"github.com/ayn2op/tview/viewport"
	"github.com/gdamore/tcell/v3"
)

// Scrolling one text view suits a single block of text, such as a help page. The text view wraps all of its text on every draw, so long text that changes often is better split into a list of text views.
func ExampleNew_textView() {
	type model struct{ scrollState viewport.ScrollState }
	view := func(m model) viewport.Widget {
		text := richtext.New(
			richtext.NewLine(richtext.NewSegment("one", tcell.StyleDefault)),
			richtext.NewLine(richtext.NewSegment("two", tcell.StyleDefault)),
			richtext.NewLine(richtext.NewSegment("three", tcell.StyleDefault)),
			richtext.NewLine(richtext.NewSegment("four", tcell.StyleDefault)),
		)
		return viewport.New(textview.New(text), m.scrollState).
			Focused(true).
			OnChange(func(s *viewport.ScrollState) tview.Msg { return *s })
	}

	area := tview.Rectangle{Width: 5, Height: 2}
	var m model
	// The model stores the state the viewport produces, as Update would.
	if s, ok := view(m).Handle(tcell.NewEventKey(tcell.KeyDown, "", tcell.ModNone), area).(viewport.ScrollState); ok {
		m.scrollState = s
	}
	screentest.Print(view(m), area.Width, area.Height)
	// Output:
	// two
	// three
}

// A list of text views suits blocks of text that are each their own item, such as chat messages. Each can be selected and styled on its own, and the viewport keeps the one at the cursor in view.
func ExampleNew_list() {
	messages := []string{"hi", "hello", "how are you?", "good"}
	type model struct {
		cursor      int
		scrollState viewport.ScrollState
	}
	view := func(m model) viewport.Widget {
		l := list.New(m.cursor, len(messages), func(i int) tview.Widget {
			text := richtext.New(richtext.NewLine(richtext.NewSegment(messages[i], tcell.StyleDefault)))
			return textview.New(text)
		})
		return viewport.New(l, m.scrollState).
			Target(l.Target(m.cursor)).
			OnChange(func(s *viewport.ScrollState) tview.Msg { return *s })
	}

	area := tview.Rectangle{Width: 6, Height: 2}
	m := model{cursor: 3}
	m.scrollState.ScrollToTarget()
	screentest.Print(view(m), area.Width, area.Height)
	// Output:
	// you?
	// good
}
