package tabs

import (
	"testing"

	"github.com/ayn2op/tview"
	"github.com/gdamore/tcell/v3"
)

type selectMsg int

type swallowed struct{}

// greedy drops every key into swallowed and records the area it handled a message in.
type greedy struct {
	area *tview.Rectangle
}

func (greedy) Draw(tview.Screen, tview.Rectangle) {}

func (g greedy) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	*g.area = area
	if _, ok := msg.(tview.KeyMsg); ok {
		return swallowed{}
	}
	return msg
}

func TestWidgetHandle(t *testing.T) {
	area := tview.Rectangle{X: 1, Y: 2, Width: 10, Height: 5}
	var contentArea tview.Rectangle
	tabs := New("tab", "tab").
		Content(greedy{&contentArea}).
		OnSelect(func(i int) tview.Msg { return selectMsg(i) })
	next := tcell.NewEventKey(tcell.KeyRune, "l", tcell.ModCtrl)
	// The labels "tab tab" are centered: the second starts at x 1 + (10-7)/2 + 4.
	click := tview.MouseMsg{EventMouse: tcell.NewEventMouse(6, 2, tcell.ButtonNone, tcell.ModNone), Action: tview.MouseLeftClick}

	for _, tt := range []struct {
		name   string
		widget Widget
		msg    tview.Msg
		want   tview.Msg
	}{
		{"next", tabs, next, selectMsg(1)},
		{"next at the last tab goes to the content", tabs.Active(1), next, swallowed{}},
		{"clicking a label", tabs, click, selectMsg(1)},
		{"clicking the active label", tabs.Active(1), click, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.widget.Handle(tt.msg, area); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("other keys go to the content below the labels", func(t *testing.T) {
		if got := tabs.Handle(tcell.NewEventKey(tcell.KeyRune, "x", tcell.ModNone), area); got != (swallowed{}) {
			t.Fatalf("got %v", got)
		}
		if want := (tview.Rectangle{X: 1, Y: 3, Width: 10, Height: 4}); contentArea != want {
			t.Fatalf("area = %+v, want %+v", contentArea, want)
		}
	})
}
