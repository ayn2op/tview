package tview

import (
	"math"

	"github.com/gdamore/tcell/v3"
)

type Alignment int

const (
	AlignmentLeft Alignment = iota
	AlignmentCenter
	AlignmentRight
)

// Print prints text onto the screen into the given box at (x,y,maxWidth,1),
// not exceeding that box. The screen's background color will not be changed.
//
// Returns the number of actual bytes of the text printed and the actual width
// used for the printed runes.
func Print(screen tcell.Screen, text string, x, y, maxWidth int, alignment Alignment, color tcell.Color) (int, int) {
	start, end, width := printWithStyle(screen, text, x, y, 0, maxWidth, alignment, tcell.StyleDefault.Foreground(color), true)
	return end - start, width
}

// printWithStyle works like [Print] but it takes a style instead of just a
// foreground color. The skipWidth parameter specifies the number of cells
// skipped at the beginning of the text. It returns the start index, end index
// (exclusively), and screen width of the text actually printed. If
// maintainBackground is "true", the existing screen background is not changed
// (i.e. the style's background color is ignored).
func printWithStyle(screen tcell.Screen, text string, x, y, skipWidth, maxWidth int, alignment Alignment, style tcell.Style, maintainBackground bool) (start, end, printedWidth int) {
	totalWidth, totalHeight := screen.Size()
	if maxWidth <= 0 || len(text) == 0 || y < 0 || y >= totalHeight {
		return 0, 0, 0
	}

	// If we don't overwrite the background, we use the default color.
	if maintainBackground {
		style = style.Background(tcell.ColorDefault)
	}

	// Skip beginning and measure width.
	var textWidth int
	state := &stepState{
		unisegState: -1,
	}
	newState := *state
	str := text
	for len(str) > 0 {
		_, str, state = step(str, state)
		if skipWidth > 0 {
			skipWidth -= state.Width()
			text = str
			newState = *state
			start += state.GrossLength()
		} else {
			textWidth += state.Width()
		}
	}
	state = &newState

	// Reduce all alignments to AlignLeft.
	switch alignment {
	case AlignmentRight:
		// Chop off characters on the left until it fits.
		for len(text) > 0 && textWidth > maxWidth {
			_, text, state = step(text, state)
			textWidth -= state.Width()
			start += state.GrossLength()
		}
		x, maxWidth = x+maxWidth-textWidth, textWidth
	case AlignmentCenter:
		// Chop off characters on the left until it fits.
		subtracted := (textWidth - maxWidth) / 2
		for len(text) > 0 && subtracted > 0 {
			_, text, state = step(text, state)
			subtracted -= state.Width()
			textWidth -= state.Width()
			start += state.GrossLength()
		}
		if textWidth < maxWidth {
			x, maxWidth = x+maxWidth/2-textWidth/2, textWidth
		}
	}

	// Draw left-aligned text.
	end = start
	rightBorder := x + maxWidth
	for len(text) > 0 && x < rightBorder && x < totalWidth {
		var c string
		c, text, state = step(text, state)
		if c == "" {
			break
		}
		width := state.Width()

		if width > 0 {
			finalStyle := style
			if maintainBackground {
				backgroundColor := finalStyle.GetBackground()
				if backgroundColor == tcell.ColorDefault {
					_, existingStyle, _ := screen.Get(x, y)
					background := existingStyle.GetBackground()
					finalStyle = finalStyle.Background(background)
				}
			}
			for offset := width - 1; offset >= 0; offset-- {
				// To avoid undesired effects, we populate all cells.
				if offset == 0 {
					screen.Put(x+offset, y, c, finalStyle)
				} else {
					screen.Put(x+offset, y, " ", finalStyle)
				}
			}
		}

		x += width
		end += state.GrossLength()
		printedWidth += width
	}

	return
}

// PrintSimple prints white text to the screen at the given position.
func PrintSimple(screen tcell.Screen, text string, x, y int) {
	Print(screen, text, x, y, math.MaxInt32, AlignmentLeft, Styles.PrimaryTextColor)
}
