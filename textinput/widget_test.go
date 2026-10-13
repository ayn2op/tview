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

// input is the text and cursor that typing in a text input left.
type input struct {
	text   string
	cursor int
}

func onInput(text string, cursor int) tview.Msg { return input{text, cursor} }

// apply sends msgs through a focused text input showing text with the cursor at its end in area, and returns what they leave.
func apply(t *testing.T, text string, area tview.Rectangle, msgs ...tview.Msg) input {
	t.Helper()
	got := input{text, len(text)}
	for _, msg := range msgs {
		out, ok := New(got.text, got.cursor).Focused(true).OnInput(onInput).Handle(msg, area).(input)
		if !ok {
			t.Fatalf("%v: got %v, want input", msg, out)
		}
		got = out
	}
	return got
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
			got := apply(t, tt.value, area, tt.msgs...)
			if got.text != tt.want || got.cursor != tt.wantCursor {
				t.Fatalf("got %q at %d, want %q at %d", got.text, got.cursor, tt.want, tt.wantCursor)
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
		out := New("ab", 2).Keybind(leftOnCtrlB).Focused(true).OnInput(onInput).Handle(key(tcell.KeyCtrlB, "", tcell.ModCtrl), area)
		if out != (input{"ab", 1}) {
			t.Fatalf("got %v, want the cursor at 1", out)
		}
	})

	t.Run("enter submits", func(t *testing.T) {
		enter := key(tcell.KeyEnter, "", tcell.ModNone)
		if got := New("x", 1).Focused(true).OnInput(onInput).OnSubmit(submitMsg{}).Handle(enter, area); got != (submitMsg{}) {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("passes other keys and unfocused input through", func(t *testing.T) {
		w := New("x", 1).OnInput(onInput)
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
	t.Run("scrolled to keep the cursor visible", func(t *testing.T) {
		screen := screentest.New(t, 3, 1)
		New("abcd", 4).Draw(screen, tview.Rectangle{Width: 3, Height: 1})
		if got := screentest.Row(screen, 0, 3); got != "cd " {
			t.Fatalf("row = %q", got)
		}
	})
	t.Run("masked", func(t *testing.T) {
		screen := screentest.New(t, 4, 1)
		New("hé!", 0).Mask("*").Draw(screen, tview.Rectangle{Width: 4, Height: 1})
		if got := screentest.Row(screen, 0, 4); got != "*** " {
			t.Fatalf("row = %q", got)
		}
	})
}
