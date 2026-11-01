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
	t.Run("compact", func(t *testing.T) {
		if got := w.formatKey("ctrl+alt+shift+meta+x"); got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})
	t.Run("spelled out", func(t *testing.T) {
		if got := w.CompactModifiers(false).formatKey("ctrl+x"); got != "ctrl+x" {
			t.Fatalf("got %q, want %q", got, "ctrl+x")
		}
	})
}

// keyMap has quit and help keybinds, in one column for the full help.
type keyMap struct{}

func (keyMap) ShortHelp() []keybind.Keybind {
	return []keybind.Keybind{keybind.NewSingleKeybind("q", "quit"), keybind.NewSingleKeybind("?", "help")}
}

func (k keyMap) FullHelp() [][]keybind.Keybind { return [][]keybind.Keybind{k.ShortHelp()} }

// longKeyMap has a dozen keybinds, as a screen with many actions shows.
type longKeyMap struct{}

func (longKeyMap) ShortHelp() []keybind.Keybind {
	var keybinds []keybind.Keybind
	for key := range strings.FieldsSeq("a b c d e f g h i j k l") {
		keybinds = append(keybinds, keybind.NewSingleKeybind("ctrl+"+key, "action "+key))
	}
	return keybinds
}

func (k longKeyMap) FullHelp() [][]keybind.Keybind { return [][]keybind.Keybind{k.ShortHelp()} }

func BenchmarkWidgetDraw(b *testing.B) {
	screen := screentest.New(b, 200, 1)
	widget := New(longKeyMap{})
	b.ReportAllocs()
	for b.Loop() {
		widget.Draw(screen, tview.Rectangle{Width: 200, Height: 1})
	}
}
