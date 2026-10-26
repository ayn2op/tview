package column

import (
	"fmt"
	"testing"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/text"
)

func TestWidgetDraw(t *testing.T) {
	screen := screentest.New(t, 2, 3)
	New(text.New("ab"), nil, text.New("cd")).Spacing(1).Draw(screen, tview.Rectangle{Width: 2, Height: 3})
	for y, want := range []string{"ab", "  ", "cd"} {
		t.Run(fmt.Sprintf("row %d", y), func(t *testing.T) {
			if got := screentest.Row(screen, y, 2); got != want {
				t.Fatalf("got %q, want %q", got, want)
			}
		})
	}
}
