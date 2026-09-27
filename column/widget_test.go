package column

import (
	"testing"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/text"
)

func TestColumn(t *testing.T) {
	screen := screentest.New(t, 2, 3)
	New(text.New("ab"), nil, text.New("cd")).Spacing(1).Draw(screen, tview.Rectangle{Width: 2, Height: 3})
	for y, want := range []string{"ab", "  ", "cd"} {
		if got := screentest.Row(screen, y, 2); got != want {
			t.Fatalf("row %d = %q, want %q", y, got, want)
		}
	}
}
