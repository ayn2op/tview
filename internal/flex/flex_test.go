package flex

import (
	"github.com/ayn2op/tview/layout"
	"slices"
	"testing"

	"github.com/ayn2op/tview"
	"github.com/gdamore/tcell/v3"
)

// sized is an element with a fixed width and height that records the area it handles a message in.
type sized struct {
	width, height layout.Length
	area          *tview.Rectangle
}

func (s sized) Size() (width, height layout.Length) { return s.width, s.height }

func (sized) Draw(tview.Screen, tview.Rectangle) {}

func (s sized) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	if s.area != nil {
		*s.area = area
	}
	if key, ok := msg.(tview.KeyMsg); ok && key.Key() == tcell.KeyEscape {
		return nil
	}
	return msg
}

func wide(width layout.Length) tview.Element  { return sized{width: width, height: layout.Fill} }
func tall(height layout.Length) tview.Element { return sized{width: layout.Fill, height: height} }

func TestWidgetAreas(t *testing.T) {
	area := tview.Rectangle{X: 1, Y: 2, Width: 10, Height: 4}
	for _, tt := range []struct {
		name   string
		layout Widget
		want   []tview.Rectangle
	}{
		{"row portions", New(true, wide(layout.Fill), wide(layout.FillPortion(2)), wide(layout.Fill)), []tview.Rectangle{
			{X: 1, Y: 2, Width: 2, Height: 4},
			{X: 3, Y: 2, Width: 5, Height: 4},
			{X: 8, Y: 2, Width: 3, Height: 4},
		}},
		{"row fixed", New(true, wide(layout.Fixed(3)), wide(layout.Fill)), []tview.Rectangle{
			{X: 1, Y: 2, Width: 3, Height: 4},
			{X: 4, Y: 2, Width: 7, Height: 4},
		}},
		{"row spacing", New(true, wide(layout.Fill), wide(layout.Fill)).Spacing(2), []tview.Rectangle{
			{X: 1, Y: 2, Width: 4, Height: 4},
			{X: 7, Y: 2, Width: 4, Height: 4},
		}},
		{"column overflow", New(false, tall(layout.Fixed(3)), tall(layout.Fixed(3)), tall(layout.Fill)), []tview.Rectangle{
			{X: 1, Y: 2, Width: 10, Height: 3},
			{X: 1, Y: 5, Width: 10, Height: 1},
			{X: 1, Y: 6, Width: 10, Height: 0},
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.layout.areas(area); !slices.Equal(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestWidgetSize(t *testing.T) {
	t.Run("shrink fits fixed children", func(t *testing.T) {
		l := New(true,
			sized{width: layout.Fixed(2), height: layout.Fixed(1)},
			sized{width: layout.Fixed(3), height: layout.Fixed(2)},
		).Width(layout.Shrink).Height(layout.Shrink).Spacing(1)
		if width, height := l.Size(); width != layout.Fixed(6) || height != layout.Fixed(2) {
			t.Fatalf("size = %+v x %+v", width, height)
		}
	})
	t.Run("keeps other lengths", func(t *testing.T) {
		l := New(false).Height(layout.Fixed(3))
		if width, height := l.Size(); width != layout.Fill || height != layout.Fixed(3) {
			t.Fatalf("size = %+v x %+v", width, height)
		}
	})
}

func TestWidgetHandle(t *testing.T) {
	var left, right tview.Rectangle
	l := New(true,
		sized{width: layout.Fill, height: layout.Fill, area: &left},
		sized{width: layout.Fill, height: layout.Fill, area: &right},
	)
	area := tview.Rectangle{Width: 10, Height: 1}
	t.Run("child areas", func(t *testing.T) {
		enter := tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone)
		if got := l.Handle(enter, area); got != enter {
			t.Fatalf("got %v", got)
		}
		if left != (tview.Rectangle{Width: 5, Height: 1}) || right != (tview.Rectangle{X: 5, Width: 5, Height: 1}) {
			t.Fatalf("left=%+v, right=%+v", left, right)
		}
	})
	t.Run("stops when dropped", func(t *testing.T) {
		right = tview.Rectangle{}
		if got := l.Handle(tcell.NewEventKey(tcell.KeyEscape, "", tcell.ModNone), area); got != nil {
			t.Fatalf("got %v", got)
		}
		if right != (tview.Rectangle{}) {
			t.Fatal("message reached the child after the one that dropped it")
		}
	})
}

func (s sized) Layout(limits layout.Limits) layout.Size {
	return layout.Atomic(limits, s.width, s.height)
}
