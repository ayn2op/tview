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
		if got := normalizeKey(tt.in); got != tt.want {
			t.Fatalf("normalizeKey(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
	if Matches(nil, NewSingleKeybind("a", "a")) {
		t.Fatal("Matches(nil) = true")
	}
	if !Matches(tcell.NewEventKey(tcell.KeyRune, "a", tcell.ModNone), NewSingleKeybind("a", "a")) {
		t.Fatal("Matches(rune a) = false")
	}
}
