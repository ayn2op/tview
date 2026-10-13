package row

import (
	"github.com/ayn2op/tview/layout"
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

func TestWidgetLayout(t *testing.T) {
	limits := layout.Limits{Max: layout.Size{Width: 10, Height: 10}}
	got := New(text.New("ab"), text.New("cde")).Width(layout.Shrink).Height(layout.Shrink).Spacing(1).Layout(limits)
	if got != (layout.Size{Width: 6, Height: 1}) {
		t.Fatalf("size = %+v", got)
	}
}
