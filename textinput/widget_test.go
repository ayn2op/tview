package textinput

import (
	"testing"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/gdamore/tcell/v3"
)

type submitMsg struct{}

func key(k tcell.Key, str string, mods tcell.ModMask) tview.KeyMsg {
	return tcell.NewEventKey(k, str, mods)
}

func typed(str string) tview.KeyMsg { return key(tcell.KeyRune, str, tcell.ModNone) }

// apply sends msgs through a focused text input showing editState in area, applying each Change, and returns the editState.
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
	area := tview.Rectangle{Width: 10, Height: 1}
	for _, tt := range []struct {
		name       string
		value      string
		msgs       []tview.Msg
		want       string
		wantCursor int
	}{
		{"typing", "", []tview.Msg{typed("h"), typed("i")}, "hi", 2},
		{"backspace", "hé", []tview.Msg{key(tcell.KeyBackspace, "", tcell.ModNone)}, "h", 1},
		{"insert in the middle", "ac", []tview.Msg{key(tcell.KeyLeft, "", tcell.ModNone), typed("b")}, "abc", 2},
		{"home and delete", "ab", []tview.Msg{key(tcell.KeyHome, "", tcell.ModNone), key(tcell.KeyDelete, "", tcell.ModNone)}, "b", 0},
		{"paste on one line", "", []tview.Msg{tview.PasteMsg("a\nb")}, "a b", 3},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := apply(t, NewEditState(tt.value), area, tt.msgs...)
			if got.value != tt.want || got.cursor != tt.wantCursor {
				t.Fatalf("got %q at %d, want %q at %d", got.value, got.cursor, tt.want, tt.wantCursor)
			}
		})
	}

	t.Run("rebound key", func(t *testing.T) {
		leftOnCtrlB := func(key tview.KeyMsg) Action {
			if key.Key() == tcell.KeyCtrlB {
				return ActionLeft
			}
			return ActionNone
		}
		editState := NewEditState("ab")
		out := New(&editState).Keybind(leftOnCtrlB).Focused(true).OnChange(func(a Change) tview.Msg { return a }).Handle(key(tcell.KeyCtrlB, "", tcell.ModCtrl), area)
		change, ok := out.(Change)
		if !ok {
			t.Fatalf("got %v, want a Change", out)
		}
		if editState.Apply(change); editState.cursor != 1 {
			t.Fatalf("cursor = %d, want 1", editState.cursor)
		}
	})

	t.Run("scrolls to keep the cursor visible", func(t *testing.T) {
		got := apply(t, NewEditState(""), tview.Rectangle{Width: 3, Height: 1}, typed("a"), typed("b"), typed("c"), typed("d"))
		if got.offset != 2 {
			t.Fatalf("offset = %d, want 2", got.offset)
		}
	})
	t.Run("enter submits", func(t *testing.T) {
		editState := NewEditState("x")
		enter := key(tcell.KeyEnter, "", tcell.ModNone)
		if got := New(&editState).Focused(true).OnChange(func(a Change) tview.Msg { return a }).OnSubmit(submitMsg{}).Handle(enter, area); got != (submitMsg{}) {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("passes other keys and unfocused input through", func(t *testing.T) {
		editState := NewEditState("x")
		w := New(&editState).OnChange(func(a Change) tview.Msg { return a })
		for _, msg := range []tview.Msg{key(tcell.KeyUp, "", tcell.ModNone), key(tcell.KeyRune, "l", tcell.ModCtrl)} {
			if got := w.Focused(true).Handle(msg, area); got != msg {
				t.Fatalf("focused %v: got %v", msg, got)
			}
		}
		if msg := typed("y"); w.Handle(msg, area) != msg {
			t.Fatal("unfocused input took a key")
		}
	})
}

func TestWidgetDraw(t *testing.T) {
	t.Run("from the scroll offset", func(t *testing.T) {
		screen := screentest.New(t, 3, 1)
		editState := EditState{value: "abcd", cursor: 4, offset: 2}
		New(&editState).Draw(screen, tview.Rectangle{Width: 3, Height: 1})
		if got := screentest.Row(screen, 0, 3); got != "cd " {
			t.Fatalf("row = %q", got)
		}
	})
	t.Run("masked", func(t *testing.T) {
		screen := screentest.New(t, 4, 1)
		editState := NewEditState("hé!")
		New(&editState).Mask("*").Draw(screen, tview.Rectangle{Width: 4, Height: 1})
		if got := screentest.Row(screen, 0, 4); got != "*** " {
			t.Fatalf("row = %q", got)
		}
	})
}
