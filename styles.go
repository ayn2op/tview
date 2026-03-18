package tview

import (
	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

// Theme defines the colors used when models are initialized.
type Theme struct {
	PrimitiveBackgroundColor    tcell.Color // Main background color for models.
	ContrastBackgroundColor     tcell.Color // Background color for contrasting elements.
	MoreContrastBackgroundColor tcell.Color // Background color for even more contrasting elements.
	BorderColor                 tcell.Color // Box borders.
	TitleColor                  tcell.Color // Box titles.
	GraphicsColor               tcell.Color // Graphics.
	PrimaryTextColor            tcell.Color // Primary text.
	SecondaryTextColor          tcell.Color // Secondary text (e.g. labels).
	TertiaryTextColor           tcell.Color // Tertiary text (e.g. subtitles, notes).
	InverseTextColor            tcell.Color // Text on primary-colored backgrounds.
	ContrastSecondaryTextColor  tcell.Color // Secondary text on ContrastBackgroundColor-colored backgrounds.
}

// Styles defines the theme for applications. The default is for a black
// background and some basic colors: black, white, yellow, green, cyan, and
// blue.
var Styles = Theme{
	PrimitiveBackgroundColor:    color.Black,
	ContrastBackgroundColor:     color.Blue,
	MoreContrastBackgroundColor: color.Green,
	BorderColor:                 color.White,
	TitleColor:                  color.White,
	GraphicsColor:               color.White,
	PrimaryTextColor:            color.White,
	SecondaryTextColor:          color.Yellow,
	TertiaryTextColor:           color.Green,
	InverseTextColor:            color.Blue,
	ContrastSecondaryTextColor:  color.Navy,
}
