package row

import (
	"testing"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/ayn2op/tview/text"
)

func TestRow(t *testing.T) {
	screen := screentest.New(t, 7, 1)
	New(text.New("ab"), nil, text.New("cd")).Spacing(1).Draw(screen, tview.Rectangle{Width: 7, Height: 1})
	if got := screentest.Row(screen, 0, 7); got != "ab cd  " {
		t.Fatalf("row = %q", got)
	}
}

func TestWidgetSize(t *testing.T) {
	width, height := New(text.New("ab"), text.New("cde")).Width(tview.Shrink).Height(tview.Shrink).Spacing(1).Size()
	if width != tview.Fixed(6) || height != tview.Fixed(1) {
		t.Fatalf("size = %+v x %+v", width, height)
	}
}
