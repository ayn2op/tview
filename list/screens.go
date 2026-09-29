package list

import (
	"github.com/ayn2op/tview"
	"github.com/gdamore/tcell/v3"
	"github.com/rivo/uniseg"
)

// clippedScreen drops drawing outside a rectangle, so items partly scrolled out of view are cut off.
type clippedScreen struct {
	tview.Screen
	area tview.Rectangle
}

func (s *clippedScreen) SetContent(x int, y int, primary rune, combining []rune, style tcell.Style) {
	if !s.area.Contains(x, y) {
		return
	}
	s.Screen.SetContent(x, y, primary, combining, style)
}

func (s *clippedScreen) Put(x int, y int, str string, style tcell.Style) (string, int) {
	if !s.area.Contains(x, y) {
		return str, 0
	}
	return s.Screen.Put(x, y, str, style)
}

func (s *clippedScreen) PutStr(x int, y int, str string) {
	s.PutStrStyled(x, y, str, tcell.StyleDefault)
}

func (s *clippedScreen) PutStrStyled(x int, y int, str string, style tcell.Style) {
	if y < s.area.Y || y >= s.area.Y+s.area.Height {
		return
	}

	gr := uniseg.NewGraphemes(str)
	for gr.Next() {
		cluster := gr.Str()
		width := max(uniseg.StringWidth(cluster), 1)
		if x >= s.area.X+s.area.Width {
			return
		}
		if x >= s.area.X && x+width <= s.area.X+s.area.Width {
			s.Screen.Put(x, y, cluster, style)
		}
		x += width
	}
}

func (s *clippedScreen) ShowCursor(x int, y int) {
	if !s.area.Contains(x, y) {
		s.Screen.ShowCursor(-1, -1)
		return
	}
	s.Screen.ShowCursor(x, y)
}

// styledScreen merges a style into every cell drawn through it. The list wraps the cursor item's screen with it so the selected item can be highlighted without rendering it differently from the rest.
type styledScreen struct {
	tview.Screen
	style tcell.Style
}

func (s *styledScreen) SetContent(x int, y int, primary rune, combining []rune, style tcell.Style) {
	s.Screen.SetContent(x, y, primary, combining, tview.MergeStyle(s.style, style))
}

func (s *styledScreen) Put(x int, y int, str string, style tcell.Style) (string, int) {
	return s.Screen.Put(x, y, str, tview.MergeStyle(s.style, style))
}

func (s *styledScreen) PutStr(x int, y int, str string) {
	s.PutStrStyled(x, y, str, tcell.StyleDefault)
}

func (s *styledScreen) PutStrStyled(x int, y int, str string, style tcell.Style) {
	s.Screen.PutStrStyled(x, y, str, tview.MergeStyle(s.style, style))
}
