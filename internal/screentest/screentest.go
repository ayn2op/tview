// Package screentest provides a mock screen for tests.
package screentest

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ayn2op/tview"
	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/vt"
)

// New returns an initialized mock screen of the given size that is finalized when the test ends.
func New(t testing.TB, width, height int) tview.Screen {
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
	var s strings.Builder
	for x := range width {
		str, _, _ := screen.Get(x, y)
		s.WriteString(str)
	}
	return s.String()
}

// Print draws widget on a mock screen of the given size and prints its rows without trailing spaces, for examples to show what a widget draws.
func Print(widget tview.Widget, width, height int) {
	screen, err := tcell.NewTerminfoScreenFromTty(vt.NewMockTerm(vt.MockOptSize{X: vt.Col(width), Y: vt.Row(height)}))
	if err != nil {
		panic(err)
	}
	if err := screen.Init(); err != nil {
		panic(err)
	}
	defer screen.Fini()

	widget.Draw(screen, tview.Rectangle{Width: width, Height: height})
	for y := range height {
		fmt.Println(strings.TrimRight(Row(screen, y, width), " "))
	}
}
