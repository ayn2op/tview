package list_test

import (
	"fmt"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/list"
	"github.com/ayn2op/tview/text"
	"github.com/gdamore/tcell/v3"
)

type cursorMsg int

// A list builds each item from its index, and reports where a key moves the cursor with OnChange.
func ExampleNew() {
	servers := []string{"Gophers", "Rustaceans", "Zig"}
	l := list.New(0, len(servers), func(i int) tview.Widget { return text.New(servers[i]) }).
		SelectedStyle(tcell.StyleDefault.Reverse(true)).
		Spacing(func(int) int { return 1 }).
		Focused(true).
		OnChange(func(cursor int) tview.Msg { return cursorMsg(cursor) })
	screentest.Print(l, 12, 5)

	fmt.Println(l.Handle(tcell.NewEventKey(tcell.KeyDown, "", tcell.ModNone), tview.Rectangle{Width: 12, Height: 5}))
	// Output:
	// Gophers
	//
	// Rustaceans
	//
	// Zig
	// 1
}
