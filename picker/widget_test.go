package picker

import (
	"testing"

	"github.com/ayn2op/tview"
	"github.com/gdamore/tcell/v3"
)

type cancelMsg struct{}

func TestWidgetHandle(t *testing.T) {
	area := tview.Rectangle{Width: 20, Height: 5}
	items := Items{{Text: "apple"}, {Text: "banana"}}
	// send handles msg the way a model does: it performs the Action the picker produced and returns any other message.
	send := func(s *SearchState, msg tview.Msg) tview.Msg {
		msg = New(items, s).
			OnAction(func(a Action) tview.Msg { return a }).
			OnSelect(func(item Item) tview.Msg { return item }).
			OnCancel(cancelMsg{}).
			Handle(msg, area)
		if a, ok := msg.(Action); ok {
			s.Perform(a)
			return nil
		}
		return msg
	}
	typed := func(str string) tview.KeyMsg { return tcell.NewEventKey(tcell.KeyRune, str, tcell.ModNone) }
	enter := tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone)

	t.Run("typing filters items", func(t *testing.T) {
		s := NewSearchState()
		send(&s, typed("b"))
		if got := send(&s, enter); got != items[1] {
			t.Fatalf("selected %v, want banana", got)
		}
	})
	t.Run("no matches selects nothing", func(t *testing.T) {
		s := NewSearchState()
		send(&s, typed("z"))
		if got := send(&s, enter); got != nil {
			t.Fatalf("selected %v", got)
		}
	})
	t.Run("clearing the query shows every item", func(t *testing.T) {
		s := NewSearchState()
		send(&s, typed("b"))
		send(&s, tcell.NewEventKey(tcell.KeyBackspace, "", tcell.ModNone))
		if got := s.count(items); got != 2 {
			t.Fatalf("count = %d, want 2", got)
		}
	})
	t.Run("list keybinds move the list, not the query", func(t *testing.T) {
		s := NewSearchState()
		send(&s, tcell.NewEventKey(tcell.KeyEnd, "", tcell.ModNone))
		if s.list.Cursor() != 1 {
			t.Fatalf("list cursor = %d, want 1", s.list.Cursor())
		}
	})
	t.Run("select returns the item under the cursor", func(t *testing.T) {
		s := NewSearchState()
		if got := send(&s, enter); got != items[0] {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("cancel", func(t *testing.T) {
		s := NewSearchState()
		if got := send(&s, tcell.NewEventKey(tcell.KeyEscape, "", tcell.ModNone)); got != (cancelMsg{}) {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("reset clears the query", func(t *testing.T) {
		s := NewSearchState()
		send(&s, typed("z"))
		s.Reset()
		if got := send(&s, enter); got != items[0] {
			t.Fatalf("got %v", got)
		}
	})
}
