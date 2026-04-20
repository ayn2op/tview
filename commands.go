package tview

import "github.com/gdamore/tcell/v3"

type Msg any

// Cmd is a side effect requested by a model during input handling.
type Cmd func() Msg

type batchMsg []Cmd

// Batch combines multiple commands into a single command.
func Batch(cmds ...Cmd) Cmd {
	var valid []Cmd
	for _, cmd := range cmds {
		if cmd == nil {
			continue
		}
		valid = append(valid, cmd)
	}
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

type InitMsg struct{}

type KeyMsg = *tcell.EventKey

type MouseMsg struct {
	*tcell.EventMouse
	Action MouseAction
}

type PasteMsg string

type quitMsg struct{}

func Quit() Cmd {
	return func() Msg {
		return quitMsg{}
	}
}

type setFocusMsg struct {
	target Model
}

func SetFocus(target Model) Cmd {
	return func() Msg {
		return setFocusMsg{target: target}
	}
}

type setMouseCaptureMsg struct {
	target Model
}

func SetMouseCapture(target Model) Cmd {
	return func() Msg {
		return setMouseCaptureMsg{target: target}
	}
}

type setTitleMsg string

func SetTitle(title string) Cmd {
	return func() Msg {
		return setTitleMsg(title)
	}
}

type getClipboardMsg struct{}

func GetClipboard() Cmd {
	return func() Msg {
		return getClipboardMsg{}
	}
}

type setClipboardMsg []byte

func SetClipboard(data []byte) Cmd {
	return func() Msg {
		return setClipboardMsg(data)
	}
}

type notifyMsg struct {
	title, body string
}

func Notify(title, body string) Cmd {
	return func() Msg {
		return notifyMsg{title: title, body: body}
	}
}
