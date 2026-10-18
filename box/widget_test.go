package box

import (
	"testing"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/layout"
	"github.com/gdamore/tcell/v3"
)

// areaWidget records the area it is drawn and handles messages in.
type areaWidget struct{ area *tview.Rectangle }

func (e areaWidget) Draw(_ tview.Screen, area tview.Rectangle) { *e.area = area }

func (e areaWidget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	*e.area = area
	return msg
}

func TestWidgetDraw(t *testing.T) {
	screen := screentest.New(t, 8, 3)
	var inner tview.Rectangle
	New(areaWidget{&inner}).Borders(tview.BordersAll).Title("ab").Draw(screen, tview.Rectangle{Width: 8, Height: 3})
	t.Run("title in the border", func(t *testing.T) {
		if got := screentest.Row(screen, 0, 8); got != "┌──ab──┐" {
			t.Fatalf("top = %q", got)
		}
	})
	t.Run("sides and bottom", func(t *testing.T) {
		for y, want := range map[int]string{1: "│      │", 2: "└──────┘"} {
			if got := screentest.Row(screen, y, 8); got != want {
				t.Fatalf("row %d = %q, want %q", y, got, want)
			}
		}
	})
	t.Run("child inside the border", func(t *testing.T) {
		if want := (tview.Rectangle{X: 1, Y: 1, Width: 6, Height: 1}); inner != want {
			t.Fatalf("inner = %+v, want %+v", inner, want)
		}
	})
}

func TestWidgetDrawCutTitle(t *testing.T) {
	screen := screentest.New(t, 8, 3)
	New(areaWidget{new(tview.Rectangle)}).Borders(tview.BordersAll).Title("abcdefgh").Draw(screen, tview.Rectangle{Width: 8, Height: 3})
	if got := screentest.Row(screen, 0, 8); got != "┌abcdef┐" {
		t.Fatalf("top = %q", got)
	}
}

func TestWidgetHandle(t *testing.T) {
	var inner, title tview.Rectangle
	b := New(areaWidget{&inner}).Padding(1, 0, 2, 0).TitleWidget(areaWidget{&title}).TitleAlignment(tview.HorizontalAlignmentLeft)
	area := tview.Rectangle{Width: 8, Height: 3}

	t.Run("child inside the padding", func(t *testing.T) {
		b.Handle(nil, area)
		if want := (tview.Rectangle{X: 2, Y: 2, Width: 6, Height: 1}); inner != want {
			t.Fatalf("inner = %+v, want %+v", inner, want)
		}
	})
	t.Run("a click on the title goes to it", func(t *testing.T) {
		b.Handle(tview.MouseMsg{EventMouse: tcell.NewEventMouse(1, 0, tcell.Button1, tcell.ModNone), Action: tview.MouseLeftDown}, area)
		if want := (tview.Rectangle{X: 1, Y: 0, Width: 6, Height: 1}); title != want {
			t.Fatalf("title = %+v, want %+v", title, want)
		}
	})
}

func (areaWidget) Size() (width, height layout.Length) { return layout.Fill, layout.Fill }

func (areaWidget) Layout(limits layout.Limits) layout.Size {
	return layout.Atomic(limits, layout.Fill, layout.Fill)
}
