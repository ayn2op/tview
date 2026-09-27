package textview

import (
	"testing"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/keybind"
	"github.com/ayn2op/tview/richtext"
	"github.com/gdamore/tcell/v3"
)

func lines(strs ...string) richtext.Text {
	text := make(richtext.Text, len(strs))
	for i, s := range strs {
		text[i] = richtext.NewLine(richtext.NewSegment(s, tcell.StyleDefault))
	}
	return text
}

func TestWidgetDraw(t *testing.T) {
	t.Run("wraps on words", func(t *testing.T) {
		screen := screentest.New(t, 6, 2)
		New(lines("ab cd ef")).Draw(screen, tview.Rectangle{Width: 6, Height: 2})
		for y, want := range []string{"ab cd ", "ef    "} {
			if got := screentest.Row(screen, y, 6); got != want {
				t.Fatalf("row %d = %q, want %q", y, got, want)
			}
		}
	})
	t.Run("centers lines", func(t *testing.T) {
		screen := screentest.New(t, 6, 1)
		New(lines("ab")).Alignment(tview.AlignmentCenter).Draw(screen, tview.Rectangle{Width: 6, Height: 1})
		if got := screentest.Row(screen, 0, 6); got != "  ab  " {
			t.Fatalf("row = %q", got)
		}
	})
	t.Run("from the scroll position", func(t *testing.T) {
		screen := screentest.New(t, 2, 1)
		var scroll ScrollState
		scroll.ScrollToEnd()
		New(lines("a", "b", "c")).ScrollState(&scroll).Draw(screen, tview.Rectangle{Width: 2, Height: 1})
		if got := screentest.Row(screen, 0, 2); got != "c " {
			t.Fatalf("row = %q", got)
		}
	})
}

func TestWidgetHandle(t *testing.T) {
	area := tview.Rectangle{Width: 4, Height: 2}
	key := func(k tcell.Key, str string) tview.KeyMsg { return tcell.NewEventKey(k, str, tcell.ModNone) }
	// scroll sends keys through a focused text view of five lines and applies the actions, returning the row reached.
	scroll := func(t *testing.T, keys ...tview.KeyMsg) int {
		t.Helper()
		text := lines("1", "2", "3", "4", "5")
		var scroll ScrollState
		for _, k := range keys {
			msg := New(text).ScrollState(&scroll).OnAction(func(a Action) tview.Msg { return a }).Focused(true).Handle(k, area)
			action, ok := msg.(Action)
			if !ok {
				t.Fatalf("key %v: got %v, want an Action", k.Name(), msg)
			}
			scroll.Perform(action)
		}
		row, _ := New(text).clamp(New(text).layout(area.Width), area, scroll.row, scroll.column, scroll.followEnd)
		return row
	}

	t.Run("down", func(t *testing.T) {
		if got := scroll(t, key(tcell.KeyDown, "")); got != 1 {
			t.Fatalf("row = %d, want 1", got)
		}
	})
	t.Run("page down by the height", func(t *testing.T) {
		if got := scroll(t, key(tcell.KeyPgDn, "")); got != 2 {
			t.Fatalf("row = %d, want 2", got)
		}
	})
	t.Run("clamped at the end", func(t *testing.T) {
		if got := scroll(t, key(tcell.KeyPgDn, ""), key(tcell.KeyPgDn, ""), key(tcell.KeyUp, "")); got != 2 {
			t.Fatalf("row = %d, want 2", got)
		}
	})
	t.Run("end then up", func(t *testing.T) {
		if got := scroll(t, key(tcell.KeyRune, "G"), key(tcell.KeyRune, "k")); got != 2 {
			t.Fatalf("row = %d, want 2", got)
		}
	})
	t.Run("modified letters pass through", func(t *testing.T) {
		var scroll ScrollState
		ctrlH := tcell.NewEventKey(tcell.KeyRune, "h", tcell.ModCtrl)
		if got := New(lines("1", "2", "3")).ScrollState(&scroll).OnAction(func(a Action) tview.Msg { return a }).Focused(true).Handle(ctrlH, area); got != ctrlH {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("rebound key", func(t *testing.T) {
		keybinds := DefaultKeybinds()
		keybinds.Down = keybind.NewSingleKeybind("n", "down")
		var scroll ScrollState
		view := New(lines("1", "2", "3")).ScrollState(&scroll).Keybinds(keybinds).OnAction(func(a Action) tview.Msg { return a }).Focused(true)
		if _, ok := view.Handle(key(tcell.KeyRune, "n"), area).(Action); !ok {
			t.Fatal("n did not scroll")
		}
		if j := key(tcell.KeyRune, "j"); view.Handle(j, area) != j {
			t.Fatal("j still scrolls")
		}
	})
	t.Run("not scrollable without OnAction", func(t *testing.T) {
		var scroll ScrollState
		down := key(tcell.KeyDown, "")
		if got := New(lines("1", "2", "3")).ScrollState(&scroll).Focused(true).Handle(down, area); got != down {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("not scrollable without ScrollState", func(t *testing.T) {
		down := key(tcell.KeyDown, "")
		if got := New(lines("1", "2", "3")).OnAction(func(a Action) tview.Msg { return a }).Focused(true).Handle(down, area); got != down {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("keys pass through unfocused", func(t *testing.T) {
		var scroll ScrollState
		down := key(tcell.KeyDown, "")
		if got := New(lines("1", "2", "3")).ScrollState(&scroll).OnAction(func(a Action) tview.Msg { return a }).Handle(down, area); got != down {
			t.Fatalf("got %v", got)
		}
	})
}

func TestWidgetRows(t *testing.T) {
	if got := New(lines("ab cd ef", "g")).Rows(6); got != 3 {
		t.Fatalf("rows = %d, want 3", got)
	}
}
