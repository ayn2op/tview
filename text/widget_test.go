package text

import (
	"testing"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

func TestWidgetDraw(t *testing.T) {
	t.Run("keeps the style beneath", func(t *testing.T) {
		screen := screentest.New(t, 5, 1)
		style := tcell.StyleDefault.Background(color.Blue)
		screen.Put(1, 0, " ", style)
		New("ab").Draw(screen, tview.Rectangle{Width: 5, Height: 1})
		if _, got, _ := screen.Get(1, 0); got != style {
			t.Fatalf("style = %v, want %v", got, style)
		}
	})
	t.Run("cut off at the edge", func(t *testing.T) {
		screen := screentest.New(t, 5, 1)
		New("abcdef").Draw(screen, tview.Rectangle{Width: 3, Height: 1})
		if got := screentest.Row(screen, 0, 5); got != "abc  " {
			t.Fatalf("row = %q", got)
		}
	})
}
