package button

import (
	"testing"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/text"
	"github.com/gdamore/tcell/v3"
)

type pressMsg struct{}

type contentMsg struct{}

type hoverMsg bool

func onHover(hovered bool) tview.Msg { return hoverMsg(hovered) }

func TestWidgetHandle(t *testing.T) {
	area := tview.Rectangle{X: 2, Y: 1, Width: 6, Height: 1}
	click := func(x, y int) tview.MouseMsg {
		return tview.MouseMsg{EventMouse: tcell.NewEventMouse(x, y, tcell.ButtonNone, tcell.ModNone), Action: tview.MouseLeftDown}
	}
	inside, outside := click(3, 1), click(9, 1)
	move := func(x, y int) tview.MouseMsg {
		return tview.MouseMsg{EventMouse: tcell.NewEventMouse(x, y, tcell.ButtonNone, tcell.ModNone), Action: tview.MouseMove}
	}
	moveIn, moveWithin, moveOut := move(3, 1), move(4, 1), move(9, 1)
	hoverable := New(text.New("OK")).OnHover(onHover)
	enter := tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone)
	button := New(text.New("OK")).OnClick(pressMsg{})

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
		{"move onto", hoverable, moveIn, hoverMsg(true)},
		{"move off", hoverable.Hovered(true), moveOut, hoverMsg(false)},
		{"move within", hoverable.Hovered(true), moveWithin, moveWithin},
		{"move outside", hoverable, moveOut, moveOut},
		{"move onto disabled", hoverable.Disabled(true), moveIn, moveIn},
		// The content, a button of one cell, is in the middle of the six: at x 2 + 5/2.
		{"click on content that takes it", New(New(text.New("x")).OnClick(contentMsg{})).OnClick(pressMsg{}), click(4, 1), contentMsg{}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.widget.Handle(tt.msg, area); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("uncomparable message", func(t *testing.T) {
		if got := button.Handle([]int{1}, area).([]int); len(got) != 1 {
			t.Fatalf("got %v, want [1]", got)
		}
	})
}

func TestWidgetDraw(t *testing.T) {
	screen := screentest.New(t, 10, 3)
	New(text.New("OK")).Draw(screen, tview.Rectangle{X: 2, Y: 0, Width: 6, Height: 3})
	if got := screentest.Row(screen, 1, 10)[2:8]; got != "  OK  " {
		t.Fatalf("label row = %q", got)
	}

	for _, tt := range []struct {
		name   string
		button Widget
		want   tview.Style
	}{
		{"normal", New(text.New("OK")), tcell.StyleDefault},
		{"styled", New(text.New("OK")).Style(tcell.StyleDefault.Bold(true)), tcell.StyleDefault.Bold(true)},
		{"focused, hovered and disabled keep the style", New(text.New("OK")).Style(tcell.StyleDefault.Bold(true)).Focused(true).Hovered(true).Disabled(true), tcell.StyleDefault.Bold(true)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			screen := screentest.New(t, 4, 1)
			tt.button.Draw(screen, tview.Rectangle{Width: 4, Height: 1})
			if _, style, _ := screen.Get(1, 0); style != tt.want {
				t.Fatalf("style = %v, want %v", style, tt.want)
			}
		})
	}
}

func pressOnP(key tview.KeyMsg) Action {
	if key.Str() == "p" {
		return ActionPress
	}
	return ActionNone
}
