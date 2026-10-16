package slider_test

import (
	"fmt"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/slider"
	"github.com/gdamore/tcell/v3"
)

type volumeMsg int

// A volume slider that steps by 10, and by 25 with shift.
func ExampleNew() {
	volume := slider.New(0, 100, 50, func(v int) tview.Msg { return volumeMsg(v) }).
		Step(10).
		ShiftStep(25).
		Focused(true)
	screentest.Print(volume, 11, 1)

	area := tview.Rectangle{Width: 11, Height: 1}
	fmt.Println(volume.Handle(tcell.NewEventKey(tcell.KeyRight, "", tcell.ModNone), area))
	fmt.Println(volume.Handle(tcell.NewEventKey(tcell.KeyLeft, "", tcell.ModShift), area))
	// Output:
	// ━━━━━●─────
	// 60
	// 25
}
