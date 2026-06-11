package picker

import (
	"github.com/ayn2op/tview"
	"github.com/gdamore/tcell/v3"
)

// row is a lightweight list.Item that draws a single line of styled segments.
// Picker entries are single-line, non-wrapping and left-aligned, so wrapping each one in a full TextView only pays for scroll/wrap/form machinery it never uses.
// row holds just the segments and a rect and draws via tview's shared print helper.
type row struct {
	line       tview.Line
	x, y, w, h int
}

func newRow(line tview.Line) *row {
	return &row{line: line}
}

func (r *row) Update(tview.Msg) tview.Cmd { return nil }

func (r *row) View(screen tcell.Screen) {
	x, maxWidth := r.x, r.w
	for _, seg := range r.line {
		if maxWidth <= 0 {
			break
		}
		_, width := tview.PrintWithStyle(screen, seg.Text, x, r.y, maxWidth, tview.AlignmentLeft, seg.Style)
		x += width
		maxWidth -= width
	}
}

func (r *row) Rect() (int, int, int, int) { return r.x, r.y, r.w, r.h }
func (r *row) SetRect(x, y, w, h int)     { r.x, r.y, r.w, r.h = x, y, w, h }
func (r *row) HasFocus() bool             { return false }
func (r *row) Focus(func(tview.Model))    {}
func (r *row) Blur()                      {}
func (r *row) Height(int) int             { return 1 }
