package tview

import (
	"github.com/ayn2op/tview/cmd"
	"github.com/gdamore/tcell/v3"
)

type (
	Msg = cmd.Msg
	Cmd = cmd.Cmd
)

// Events from tcell that models receive as they are.
type (
	KeyMsg       = *tcell.EventKey
	ResizeMsg    = *tcell.EventResize
	ClipboardMsg = *tcell.EventClipboard
	FocusMsg     = *tcell.EventFocus
)

type MouseMsg struct {
	*tcell.EventMouse
	Action MouseAction
}

type PasteMsg string

// TerminalInfoMsg holds the terminal's name and version, which are empty if unknown.
type TerminalInfoMsg struct {
	Name, Version string
}
