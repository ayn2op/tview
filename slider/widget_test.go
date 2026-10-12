package slider

import (
	"testing"
	"time"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/gdamore/tcell/v3"
)

func TestWidgetDraw(t *testing.T) {
	for value, want := range map[int]string{0: "●───", 50: "━●──", 100: "━━━●", 200: "━━━●"} {
		screen := screentest.New(t, 4, 1)
		New(0, 100, value, nil).Draw(screen, tview.Rectangle{Width: 4, Height: 1})

		if got := screentest.Row(screen, 0, 4); got != want {
			t.Errorf("value %d: row = %q, want %q", value, got, want)
		}
	}
}

func TestWidgetHandle(t *testing.T) {
	area := tview.Rectangle{X: 2, Width: 11, Height: 1}
	w := New(0, 10, 8, func(value int) tview.Msg { return value }).Step(5)
	right := tcell.NewEventKey(tcell.KeyRight, "", tcell.ModNone)
	left := tcell.NewEventKey(tcell.KeyLeft, "", tcell.ModNone)
	shiftLeft := tcell.NewEventKey(tcell.KeyLeft, "", tcell.ModShift)
	click := tview.MouseMsg{EventMouse: tcell.NewEventMouse(6, 0, tcell.ButtonPrimary, tcell.ModNone), Action: tview.MouseLeftDown}

	seek := New(0, time.Minute, 0, func(value time.Duration) tview.Msg { return value }).Step(time.Second)
	if got := seek.Handle(click, area); got != 24*time.Second {
		t.Errorf("duration: got %v, want 24s", got)
	}

	for name, tt := range map[string]struct {
		w    Widget[int]
		msg  tview.Msg
		want tview.Msg
	}{
		"right":             {w.Focused(true), right, 10},
		"left":              {w.Focused(true), left, 3},
		"shift step":        {w.ShiftStep(1).Focused(true), shiftLeft, 7},
		"shift, no step":    {w.Focused(true), shiftLeft, 3},
		"not focused":       {w, right, right},
		"right at max":      {New(0, 10, 10, nil).Focused(true), right, right},
		"click rounds to 5": {w, click, 5},
	} {
		if got := tt.w.Handle(tt.msg, area); got != tt.want {
			t.Errorf("%s: got %v, want %v", name, got, tt.want)
		}
	}
}
