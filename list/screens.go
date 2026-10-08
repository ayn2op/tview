package list

import (
	"github.com/ayn2op/tview"
	"github.com/gdamore/tcell/v3"
)

// styledScreen merges a style into every cell drawn through it. The list wraps the cursor item's screen with it so the selected item can be highlighted without rendering it differently from the rest.
type styledScreen struct {
	tview.Screen
	style tview.Style
}

func (s *styledScreen) SetContent(x int, y int, primary rune, combining []rune, style tview.Style) {
	s.Screen.SetContent(x, y, primary, combining, tview.MergeStyle(s.style, style))
}

func (s *styledScreen) Put(x int, y int, str string, style tview.Style) (string, int) {
	return s.Screen.Put(x, y, str, tview.MergeStyle(s.style, style))
}

func (s *styledScreen) FillArea(x, y, width, height int, r rune, style tview.Style) {
	s.Screen.FillArea(x, y, width, height, r, tview.MergeStyle(s.style, style))
}

func (s *styledScreen) PutStr(x int, y int, str string) {
	s.PutStrStyled(x, y, str, tcell.StyleDefault)
}

func (s *styledScreen) PutStrStyled(x int, y int, str string, style tview.Style) {
	s.Screen.PutStrStyled(x, y, str, tview.MergeStyle(s.style, style))
}
