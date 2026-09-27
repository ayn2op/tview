package center

import (
	"testing"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/text"
)

func TestWidgetDraw(t *testing.T) {
	screen := screentest.New(t, 7, 3)
	New(text.New("ab")).Draw(screen, tview.Rectangle{Width: 7, Height: 3})
	if got := screentest.Row(screen, 1, 7); got != "  ab   " {
		t.Fatalf("middle row = %q", got)
	}
}
