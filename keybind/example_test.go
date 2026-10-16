package keybind_test

import (
	"fmt"

	"github.com/ayn2op/tview/keybind"
	"github.com/gdamore/tcell/v3"
)

// Update matches a key against keybinds, each of which may have several keys.
func ExampleMatches() {
	quit := keybind.New("q", "ctrl+c")
	save := keybind.New("ctrl+s")

	key := tcell.NewEventKey(tcell.KeyCtrlC, "", tcell.ModCtrl)
	fmt.Println(keybind.String(key))
	fmt.Println(keybind.Matches(key, quit))
	fmt.Println(keybind.Matches(key, save))
	// Output:
	// ctrl+c
	// true
	// false
}
