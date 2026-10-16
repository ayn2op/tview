package tabs_test

import (
	"fmt"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/tabs"
	"github.com/gdamore/tcell/v3"
)

type tabMsg int

// Tabs too wide for their area keep the active one centered, with arrows where more are hidden.
func ExampleNew() {
	t := tabs.New("General", "Appearance", "Keybinds", "About").
		Active(1).
		Arrows("◀", "▶").
		OnSelect(func(index int) tview.Msg { return tabMsg(index) })
	screentest.Print(t, 24, 1)

	next := tcell.NewEventKey(tcell.KeyCtrlL, "", tcell.ModCtrl)
	fmt.Println(t.Handle(next, tview.Rectangle{Width: 24, Height: 1}))
	// Output:
	// ◀neral Appearance Keybi▶
	// 2
}
