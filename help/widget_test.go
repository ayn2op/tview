package help

import (
	"runtime"
	"strings"
	"testing"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/keybind"
)

func TestWidgetFormatKey(t *testing.T) {
	want := "^A-S-M-x"
	if runtime.GOOS == "darwin" {
		want = "⌃⌥⇧⌘x"
	}

	w := New(nil).CompactModifiers(true)
	if got := w.formatKey("ctrl+alt+shift+meta+x"); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if got := w.CompactModifiers(false).formatKey("ctrl+x"); got != "ctrl+x" {
		t.Fatalf("got %q, want %q", got, "ctrl+x")
	}
}

// keyMap has quit and help keybinds, in one column for the full help.
type keyMap struct{}

func (keyMap) ShortHelp() []keybind.Keybind {
	return []keybind.Keybind{keybind.NewSingleKeybind("q", "quit"), keybind.NewSingleKeybind("?", "help")}
}

func (k keyMap) FullHelp() [][]keybind.Keybind { return [][]keybind.Keybind{k.ShortHelp()} }

func TestWidgetDraw(t *testing.T) {
	t.Run("short help on one line", func(t *testing.T) {
		screen := screentest.New(t, 20, 2)
		New(keyMap{}).Draw(screen, tview.Rectangle{Width: 20, Height: 2})
		if got := screentest.Row(screen, 0, 20); !strings.HasPrefix(got, "q quit • ? help") {
			t.Fatalf("row = %q", got)
		}
	})
	t.Run("full help in columns", func(t *testing.T) {
		screen := screentest.New(t, 20, 2)
		New(keyMap{}).ShowAll(true).Draw(screen, tview.Rectangle{Width: 20, Height: 2})
		if got := screentest.Row(screen, 1, 20); !strings.HasPrefix(got, "? help") {
			t.Fatalf("second row = %q", got)
		}
	})
}
