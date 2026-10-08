// Package clip cuts drawing off at the edges of a rectangle.
package clip

import (
	"github.com/ayn2op/tview"
	"github.com/gdamore/tcell/v3"
	"github.com/rivo/uniseg"
)

// Screen drops drawing outside Area.
type Screen struct {
	tview.Screen
	Area tview.Rectangle
}

func (s *Screen) SetContent(x int, y int, primary rune, combining []rune, style tview.Style) {
	if !s.Area.Contains(x, y) {
		return
	}
	s.Screen.SetContent(x, y, primary, combining, style)
}

func (s *Screen) Put(x int, y int, str string, style tview.Style) (string, int) {
	if !s.Area.Contains(x, y) {
		return str, 0
	}
	return s.Screen.Put(x, y, str, style)
}

func (s *Screen) FillArea(x, y, width, height int, r rune, style tview.Style) {
	x0, y0 := max(x, s.Area.X), max(y, s.Area.Y)
	x1, y1 := min(x+width, s.Area.X+s.Area.Width), min(y+height, s.Area.Y+s.Area.Height)
	s.Screen.FillArea(x0, y0, x1-x0, y1-y0, r, style)
}

func (s *Screen) PutStr(x int, y int, str string) {
	s.PutStrStyled(x, y, str, tcell.StyleDefault)
}

func (s *Screen) PutStrStyled(x int, y int, str string, style tview.Style) {
	if y < s.Area.Y || y >= s.Area.Y+s.Area.Height {
		return
	}

	gr := uniseg.NewGraphemes(str)
	for gr.Next() {
		cluster := gr.Str()
		width := max(uniseg.StringWidth(cluster), 1)
		if x >= s.Area.X+s.Area.Width {
			return
		}
		if x >= s.Area.X && x+width <= s.Area.X+s.Area.Width {
			s.Screen.Put(x, y, cluster, style)
		}
		x += width
	}
}

func (s *Screen) ShowCursor(x int, y int) {
	if !s.Area.Contains(x, y) {
		s.Screen.ShowCursor(-1, -1)
		return
	}
	s.Screen.ShowCursor(x, y)
}
