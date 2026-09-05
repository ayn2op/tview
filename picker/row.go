package picker

import (
	"github.com/ayn2op/tview"
	"github.com/gdamore/tcell/v3"
)

// row is a lightweight list.Item that draws a single line of text.
// Picker entries are single-line, non-wrapping and left-aligned, so wrapping each one in a full TextView only pays for scroll/wrap/form machinery it never uses.
type row struct {
	text       string
	x, y, w, h int
}

func (r *row) Update(tview.Msg) tview.Cmd { return nil }

func (r *row) View(screen tcell.Screen) {
	if r.w <= 0 {
		return
	}
	tview.PrintWithStyle(screen, r.text, r.x, r.y, r.w, tview.AlignmentLeft, tcell.StyleDefault)
}

func (r *row) Rect() (int, int, int, int) { return r.x, r.y, r.w, r.h }
func (r *row) SetRect(x, y, w, h int)     { r.x, r.y, r.w, r.h = x, y, w, h }
func (r *row) Height(int) int             { return 1 }
