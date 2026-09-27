// Package screentest provides a mock screen for tests.
package screentest

import (
	"testing"

	"github.com/ayn2op/tview"
	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/vt"
)

// New returns an initialized mock screen of the given size that is finalized when the test ends.
func New(t *testing.T, width, height int) tview.Screen {
	t.Helper()
	screen, err := tcell.NewTerminfoScreenFromTty(vt.NewMockTerm(vt.MockOptSize{X: vt.Col(width), Y: vt.Row(height)}))
	if err != nil {
		t.Fatal(err)
	}
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(screen.Fini)
	return screen
}

// Row returns the text of the first width cells of row y.
func Row(screen tview.Screen, y, width int) string {
	var s string
	for x := range width {
		str, _, _ := screen.Get(x, y)
		s += str
	}
	return s
}
