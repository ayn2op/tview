package textarea

import (
	"testing"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/gdamore/tcell/v3"
)

func key(k tcell.Key, str string) tview.KeyMsg { return tcell.NewEventKey(k, str, tcell.ModNone) }

// input is the text and cursor that typing in a text area left.
type input struct {
	text   string
	cursor int
}

func onInput(text string, cursor int) tview.Msg { return input{text, cursor} }

// apply sends msgs through a focused text area showing text with the cursor at its end in area, and returns what they leave.
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
			got := apply(t, tt.value, area, tt.msgs...)
			if got.text != tt.want || got.cursor != tt.wantCursor {
				t.Fatalf("got %q at %d, want %q at %d", got.text, got.cursor, tt.want, tt.wantCursor)
			}
		})
	}

	t.Run("passes other keys and unfocused input through", func(t *testing.T) {
		w := New("x", 1).OnInput(onInput)
		tab := key(tcell.KeyTab, "")
		if got := w.Focused(true).Handle(tab, area); got != tab {
			t.Fatalf("got %v", got)
		}
		if msg := key(tcell.KeyRune, "y"); w.Handle(msg, area) != msg {
			t.Fatal("unfocused text area took a key")
		}
	})
}

func TestWidgetDraw(t *testing.T) {
	draw := func(t *testing.T, text string, want ...string) {
		screen := screentest.New(t, 5, 2)
		New(text, len(text)).Draw(screen, tview.Rectangle{Width: 5, Height: 2})
		for y, want := range want {
			if got := screentest.Row(screen, y, 5); got != want {
				t.Fatalf("row %d = %q, want %q", y, got, want)
			}
		}
	}

	t.Run("wrapped", func(t *testing.T) { draw(t, "ab cd ef", "ab   ", "cd ef") })
	t.Run("scrolled to keep the cursor visible", func(t *testing.T) { draw(t, "a\nb\nc", "b    ", "c    ") })
}
