package button

import (
	"testing"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/gdamore/tcell/v3"
)

type pressMsg struct{}

func TestWidgetHandle(t *testing.T) {
	area := tview.Rectangle{X: 2, Y: 1, Width: 6, Height: 1}
	click := func(x, y int) tview.MouseMsg {
		return tview.MouseMsg{EventMouse: tcell.NewEventMouse(x, y, tcell.ButtonNone, tcell.ModNone), Action: tview.MouseLeftClick}
	}
	inside, outside := click(3, 1), click(9, 1)
	enter := tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone)
	button := New().Label("OK").OnClick(pressMsg{})

	for _, tt := range []struct {
		name   string
		widget Widget
		msg    tview.Msg
		want   tview.Msg
	}{
		{"click inside", button, inside, pressMsg{}},
		{"click outside", button, outside, outside},
		{"enter unfocused", button, enter, enter},
		{"enter focused", button.Focused(true), enter, pressMsg{}},
		{"enter rebound away", button.Focused(true).Keybind(pressOnP), enter, enter},
		{"rebound key", button.Focused(true).Keybind(pressOnP), tcell.NewEventKey(tcell.KeyRune, "p", tcell.ModNone), pressMsg{}},
		{"disabled", button.Disabled(true), inside, inside},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.widget.Handle(tt.msg, area); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWidgetDraw(t *testing.T) {
	screen := screentest.New(t, 10, 3)
	New().Label("OK").Draw(screen, tview.Rectangle{X: 2, Y: 0, Width: 6, Height: 3})
	if got := screentest.Row(screen, 1, 10)[2:8]; got != "  OK  " {
		t.Fatalf("label row = %q", got)
	}

	for _, tt := range []struct {
		name   string
		button Widget
		want   tcell.Style
	}{
		{"normal", New(), tcell.StyleDefault},
		{"focused", New().Focused(true), tcell.StyleDefault.Reverse(true)},
		{"disabled", New().Focused(true).Disabled(true), tcell.StyleDefault.Dim(true)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			screen := screentest.New(t, 4, 1)
			tt.button.Label("OK").Draw(screen, tview.Rectangle{Width: 4, Height: 1})
			if _, style, _ := screen.Get(1, 0); style != tt.want {
				t.Fatalf("style = %v, want %v", style, tt.want)
			}
		})
	}
}

func pressOnP(key tview.KeyMsg) (Action, bool) {
	return ActionPress, key.Str() == "p"
}
