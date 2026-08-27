package help

import (
	"runtime"
	"testing"
)

func TestCompactModifiers(t *testing.T) {
	want := "^A-S-M-x"
	if runtime.GOOS == "darwin" {
		want = "⌃⌥⇧⌘x"
	}

	m := NewModel().SetCompactModifiers(true)
	if got := m.formatKey("ctrl+alt+shift+meta+x"); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if got := m.SetCompactModifiers(false).formatKey("ctrl+x"); got != "ctrl+x" {
		t.Fatalf("got %q, want %q", got, "ctrl+x")
	}
}
