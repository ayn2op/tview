package backdrop

import (
	"testing"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/gdamore/tcell/v3"
)

func TestWidgetDraw(t *testing.T) {
	screen := screentest.New(t, 4, 2)
	screen.Put(2, 1, "a", tcell.StyleDefault)
	New().Draw(screen, tview.Rectangle{Width: 4, Height: 2})
	str, style, _ := screen.Get(2, 1)
	if str != "a" || !style.HasDim() {
		t.Fatalf("cell = %q dim=%v, want %q dim=true", str, style.HasDim(), "a")
	}
}
