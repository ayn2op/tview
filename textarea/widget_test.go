package textarea

import (
	"testing"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/gdamore/tcell/v3"
)

func key(k tcell.Key, str string) tview.KeyMsg { return tcell.NewEventKey(k, str, tcell.ModNone) }

// apply sends msgs through a focused text area showing editState in area, applying each Change, and returns the editState.
func apply(t *testing.T, editState EditState, area tview.Rectangle, msgs ...tview.Msg) EditState {
	t.Helper()
	for _, msg := range msgs {
		out := New(&editState).Focused(true).OnChange(func(a Change) tview.Msg { return a }).Handle(msg, area)
		change, ok := out.(Change)
		if !ok {
			t.Fatalf("%v: got %v, want a Change", msg, out)
		}
		editState.Apply(change)
	}
	return editState
}

func TestWidgetHandle(t *testing.T) {
	area := tview.Rectangle{Width: 5, Height: 2}
	for _, tt := range []struct {
		name       string
		value      string
		msgs       []tview.Msg
		want       string
		wantCursor int
	}{
		{"typing", "", []tview.Msg{key(tcell.KeyRune, "a"), key(tcell.KeyEnter, ""), key(tcell.KeyRune, "b")}, "a\nb", 3},
		{"backspace joins lines", "a\nb", []tview.Msg{key(tcell.KeyLeft, ""), key(tcell.KeyBackspace, "")}, "ab", 1},
		{"up keeps the column", "abc\nde", []tview.Msg{key(tcell.KeyUp, "")}, "abc\nde", 2},
		{"down over a wrapped line", "ab cd ef", []tview.Msg{key(tcell.KeyHome, ""), key(tcell.KeyUp, ""), key(tcell.KeyUp, ""), key(tcell.KeyDown, "")}, "ab cd ef", 3},
		{"paste", "", []tview.Msg{tview.PasteMsg("a\r\nb")}, "a\nb", 3},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := apply(t, NewEditState(tt.value), area, tt.msgs...)
			if got.value != tt.want || got.cursor != tt.wantCursor {
				t.Fatalf("got %q at %d, want %q at %d", got.value, got.cursor, tt.want, tt.wantCursor)
			}
		})
	}

	t.Run("scrolls to keep the cursor visible", func(t *testing.T) {
		got := apply(t, NewEditState(""), tview.Rectangle{Width: 5, Height: 2}, key(tcell.KeyEnter, ""), key(tcell.KeyEnter, ""))
		if got.row != 1 {
			t.Fatalf("row = %d, want 1", got.row)
		}
	})
	t.Run("passes other keys and unfocused input through", func(t *testing.T) {
		editState := NewEditState("x")
		w := New(&editState).OnChange(func(a Change) tview.Msg { return a })
		tab := key(tcell.KeyTab, "")
		if got := w.Focused(true).Handle(tab, area); got != tab {
			t.Fatalf("got %v", got)
		}
		if msg := key(tcell.KeyRune, "y"); w.Handle(msg, area) != msg {
			t.Fatal("unfocused text area took a key")
		}
	})
}

func TestStateUndo(t *testing.T) {
	editState := apply(t, NewEditState(""), tview.Rectangle{Width: 5, Height: 1}, key(tcell.KeyRune, "a"), key(tcell.KeyRune, "b"))
	editState.Undo()
	if editState.Value() != "a" || editState.Cursor() != 1 {
		t.Fatalf("got %q at %d, want %q at 1", editState.Value(), editState.Cursor(), "a")
	}
}

func TestWidgetDraw(t *testing.T) {
	screen := screentest.New(t, 5, 2)
	editState := NewEditState("ab cd ef")
	New(&editState).Draw(screen, tview.Rectangle{Width: 5, Height: 2})
	for y, want := range []string{"ab   ", "cd ef"} {
		if got := screentest.Row(screen, y, 5); got != want {
			t.Fatalf("row %d = %q, want %q", y, got, want)
		}
	}
}
