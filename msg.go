package tview

import (
	"slices"

	"github.com/gdamore/tcell/v3"
)

type Msg any

// Cmd is work done outside Update, such as I/O, whose result comes back as a message.
type Cmd func() Msg

type rawMsg struct{ data any }

// WriteRaw is a command that writes the data to the underlying TTY without any formatting.
func WriteRaw(r any) Cmd {
	return func() Msg { return rawMsg{data: r} }
}

type batchMsg []Cmd

// Batch combines multiple commands into a single command.
func Batch(cmds ...Cmd) Cmd {
	valid := compactCmds(cmds...)
	switch len(valid) {
	case 0:
		return nil
	case 1:
		return valid[0]
	default:
		return func() Msg {
			return batchMsg(valid)
		}
	}
}

type sequenceMsg []Cmd

// Sequence executes commands one at a time, in order.
func Sequence(cmds ...Cmd) Cmd {
	valid := compactCmds(cmds...)
	switch len(valid) {
	case 0:
		return nil
	case 1:
		return valid[0]
	default:
		return func() Msg {
			return sequenceMsg(valid)
		}
	}
}

func compactCmds(cmds ...Cmd) []Cmd {
	return slices.DeleteFunc(cmds, func(cmd Cmd) bool { return cmd == nil })
}

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

type quitMsg struct{}

func Quit() Cmd {
	return func() Msg { return quitMsg{} }
}

type suspendMsg Cmd

// Suspend runs a command while the terminal is suspended.
func Suspend(cmd Cmd) Cmd {
	if cmd == nil {
		return nil
	}
	return func() Msg { return suspendMsg(cmd) }
}

type setTitleMsg string

func SetTitle(title string) Cmd {
	return func() Msg { return setTitleMsg(title) }
}

type getClipboardMsg struct{}

// GetClipboard asks the terminal for the clipboard, which arrives as a ClipboardMsg.
func GetClipboard() Cmd {
	return func() Msg { return getClipboardMsg{} }
}

type setClipboardMsg []byte

func SetClipboard(data []byte) Cmd {
	return func() Msg { return setClipboardMsg(data) }
}

type notifyMsg struct {
	title, body string
}

func Notify(title, body string) Cmd {
	return func() Msg { return notifyMsg{title: title, body: body} }
}
