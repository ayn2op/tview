package cmd

import "github.com/gdamore/tcell/v3/color"

// SetForegroundColor sets the terminal's default foreground color. An invalid color, such as color.Default, restores the terminal's own.
type SetForegroundColor color.Color

func (SetForegroundColor) isCmd() {}

// SetBackgroundColor sets the terminal's default background color. An invalid color, such as color.Default, restores the terminal's own.
type SetBackgroundColor color.Color

func (SetBackgroundColor) isCmd() {}
