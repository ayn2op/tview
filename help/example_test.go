package help_test

import (
	"github.com/ayn2op/tview/help"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/keybind"
)

// keyMap is the keybinds of an editor, which help lists.
type keyMap struct {
	save, quit, help keybind.Keybind
}

func (k keyMap) ShortHelp() []keybind.Keybind {
	return []keybind.Keybind{k.save, k.quit, k.help}
}

func (k keyMap) FullHelp() [][]keybind.Keybind {
	return [][]keybind.Keybind{{k.save}, {k.quit, k.help}}
}

func ExampleNew() {
	keys := keyMap{
		save: keybind.NewSingleKeybind("ctrl+s", "save"),
		quit: keybind.NewSingleKeybind("ctrl+q", "quit"),
		help: keybind.NewSingleKeybind("?", "more"),
	}
	screentest.Print(help.New(keys), 40, 1)
	screentest.Print(help.New(keys).ShowAll(true), 40, 2)
	// Output:
	// ctrl+s save • ctrl+q quit • ? more
	// ctrl+s save    ctrl+q quit
	//                ?      more
}
