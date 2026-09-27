package flex

import (
	"slices"
	"testing"

	"github.com/ayn2op/tview"
	"github.com/gdamore/tcell/v3"
)

// sized is an element with a fixed width and height that records the area it handles a message in.
type sized struct {
	width, height tview.Length
	area          *tview.Rectangle
}

func (s sized) Size() (width, height tview.Length) { return s.width, s.height }

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

func wide(width tview.Length) tview.Element  { return sized{width: width, height: tview.Fill} }
func tall(height tview.Length) tview.Element { return sized{width: tview.Fill, height: height} }

func TestLayoutAreas(t *testing.T) {
	area := tview.Rectangle{X: 1, Y: 2, Width: 10, Height: 4}
	for _, tt := range []struct {
		name   string
		layout Layout
		want   []tview.Rectangle
	}{
		{"row portions", Layout{Horizontal: true, Children: []tview.Element{wide(tview.Fill), wide(tview.FillPortion(2)), wide(tview.Fill)}}, []tview.Rectangle{
			{X: 1, Y: 2, Width: 2, Height: 4},
			{X: 3, Y: 2, Width: 5, Height: 4},
			{X: 8, Y: 2, Width: 3, Height: 4},
		}},
		{"row fixed", Layout{Horizontal: true, Children: []tview.Element{wide(tview.Fixed(3)), wide(tview.Fill)}}, []tview.Rectangle{
			{X: 1, Y: 2, Width: 3, Height: 4},
			{X: 4, Y: 2, Width: 7, Height: 4},
		}},
		{"row spacing", Layout{Horizontal: true, Spacing: 2, Children: []tview.Element{wide(tview.Fill), wide(tview.Fill)}}, []tview.Rectangle{
			{X: 1, Y: 2, Width: 4, Height: 4},
			{X: 7, Y: 2, Width: 4, Height: 4},
		}},
		{"column overflow", Layout{Children: []tview.Element{tall(tview.Fixed(3)), tall(tview.Fixed(3)), tall(tview.Fill)}}, []tview.Rectangle{
			{X: 1, Y: 2, Width: 10, Height: 3},
			{X: 1, Y: 5, Width: 10, Height: 1},
			{X: 1, Y: 6, Width: 10, Height: 0},
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.layout.Areas(area); !slices.Equal(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestLayoutSize(t *testing.T) {
	t.Run("shrink fits fixed children", func(t *testing.T) {
		l := Layout{Horizontal: true, Width: tview.Shrink, Height: tview.Shrink, Spacing: 1, Children: []tview.Element{
			sized{width: tview.Fixed(2), height: tview.Fixed(1)},
			sized{width: tview.Fixed(3), height: tview.Fixed(2)},
		}}
		if width, height := l.Size(); width != tview.Fixed(6) || height != tview.Fixed(2) {
			t.Fatalf("size = %+v x %+v", width, height)
		}
	})
	t.Run("keeps other lengths", func(t *testing.T) {
		l := Layout{Width: tview.Fill, Height: tview.Fixed(3)}
		if width, height := l.Size(); width != tview.Fill || height != tview.Fixed(3) {
			t.Fatalf("size = %+v x %+v", width, height)
		}
	})
}

func TestLayoutHandle(t *testing.T) {
	var left, right tview.Rectangle
	l := Layout{Horizontal: true, Children: []tview.Element{
		sized{width: tview.Fill, height: tview.Fill, area: &left},
		sized{width: tview.Fill, height: tview.Fill, area: &right},
	}}
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
