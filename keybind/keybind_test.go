package keybind

import (
	"testing"

	"github.com/gdamore/tcell/v3"
)

func TestNormalizeKey(t *testing.T) {
	for _, tt := range []struct{ in, want string }{
		{"", ""},
		{"  ", ""},
		{"a", "a"},
		{"Enter", "enter"},
		{"ctrl-c", "ctrl+c"},
		{"CTRL-C", "ctrl+c"},
		{"ctrl+ctrl+a", "ctrl+a"},
		{"shift+backtab", "shift+tab"},
	} {
		t.Run(tt.in, func(t *testing.T) {
			if got := normalizeKey(tt.in); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMatches(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		if Matches(nil, NewSingleKeybind("a", "a")) {
			t.Fatal("matched")
		}
	})
	t.Run("rune", func(t *testing.T) {
		if !Matches(tcell.NewEventKey(tcell.KeyRune, "a", tcell.ModNone), NewSingleKeybind("a", "a")) {
			t.Fatal("did not match")
		}
	})
	t.Run("space", func(t *testing.T) {
		if !Matches(tcell.NewEventKey(tcell.KeyRune, " ", tcell.ModNone), New("space")) {
			t.Fatal("did not match")
		}
	})
}
