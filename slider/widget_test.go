package slider

import (
	"fmt"
	"testing"
	"time"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/gdamore/tcell/v3"
)

func TestWidgetDraw(t *testing.T) {
	for value, want := range map[int]string{0: "●───", 50: "━●──", 100: "━━━●", 200: "━━━●"} {
		t.Run(fmt.Sprint(value), func(t *testing.T) {
			screen := screentest.New(t, 4, 1)
			New(0, 100, value, nil).Draw(screen, tview.Rectangle{Width: 4, Height: 1})
			if got := screentest.Row(screen, 0, 4); got != want {
				t.Errorf("row = %q, want %q", got, want)
			}
		})
	}
}

func TestWidgetHandle(t *testing.T) {
	area := tview.Rectangle{X: 2, Width: 11, Height: 1}
	type pressed int
	w := New(0, 10, 8, func(value int) tview.Msg { return value }).Step(5)
	dragged := w.Dragging(true).OnRelease("released")
	right := tcell.NewEventKey(tcell.KeyRight, "", tcell.ModNone)
	left := tcell.NewEventKey(tcell.KeyLeft, "", tcell.ModNone)
	shiftLeft := tcell.NewEventKey(tcell.KeyLeft, "", tcell.ModShift)
	mouse := func(x, y int, buttons tcell.ButtonMask, mod tcell.ModMask, action tview.MouseAction) tview.MouseMsg {
		return tview.MouseMsg{EventMouse: tcell.NewEventMouse(x, y, buttons, mod), Action: action}
	}
	click := mouse(6, 0, tcell.ButtonPrimary, tcell.ModNone, tview.MouseLeftDown)
	// away is a move with the button held, off the rail.
	away := mouse(40, 3, tcell.ButtonPrimary, tcell.ModNone, tview.MouseMove)
	wheel := mouse(6, 0, tcell.WheelDown, tcell.ModNone, tview.MouseScrollDown)

	t.Run("duration", func(t *testing.T) {
		seek := New(0, time.Minute, 0, func(value time.Duration) tview.Msg { return value }).Step(time.Second)
		if got := seek.Handle(click, area); got != 24*time.Second {
			t.Errorf("got %v, want 24s", got)
		}
	})

	for name, tt := range map[string]struct {
		w    Widget[int]
		msg  tview.Msg
		want tview.Msg
	}{
		"right":                  {w.Focused(true), right, 10},
		"left":                   {w.Focused(true), left, 3},
		"shift step":             {w.ShiftStep(1).Focused(true), shiftLeft, 7},
		"shift, no step":         {w.Focused(true), shiftLeft, 3},
		"not focused":            {w, right, right},
		"right at max":           {New(0, 10, 10, func(value int) tview.Msg { return value }).Focused(true), right, right},
		"click rounds to 5":      {w, click, 5},
		"press":                  {w.OnPress(func(value int) tview.Msg { return pressed(value) }), click, pressed(5)},
		"drag off the rail":      {dragged, away, 10},
		"move when not dragging": {w, away, away},
		"release":                {dragged, mouse(40, 3, tcell.ButtonNone, tcell.ModNone, tview.MouseLeftUp), "released"},
		"missed release":         {dragged, mouse(6, 0, tcell.ButtonNone, tcell.ModNone, tview.MouseMove), "released"},
		"ctrl press":             {w.Default(2), mouse(6, 0, tcell.ButtonPrimary, tcell.ModCtrl, tview.MouseLeftDown), 2},
		"ctrl wheel":             {w, mouse(6, 0, tcell.WheelDown, tcell.ModCtrl, tview.MouseScrollDown), 3},
		"wheel":                  {w, wheel, wheel},
	} {
		t.Run(name, func(t *testing.T) {
			if got := tt.w.Handle(tt.msg, area); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
