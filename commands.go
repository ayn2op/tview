package tview

import "github.com/gdamore/tcell/v3"

type Msg = tcell.Event

// Cmd is a side effect requested by a model during input handling.
type Cmd func() Msg

type batchMsg struct {
	tcell.EventTime
	cmds []Cmd
}

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
			return &batchMsg{cmds: valid}
		}
	}
}

type InitMsg struct{ tcell.EventTime }

type KeyMsg = tcell.EventKey

type MouseMsg struct {
	tcell.EventMouse
	Action MouseAction
}

type PasteMsg struct {
	tcell.EventTime
	Content string
}

type quitMsg struct{ tcell.EventTime }

func Quit() Cmd {
	return func() Msg {
		return &quitMsg{}
	}
}

type setFocusMsg struct {
	tcell.EventTime
	target Model
}

func SetFocus(target Model) Cmd {
	return func() Msg {
		return &setFocusMsg{target: target}
	}
}

type setMouseCaptureMsg struct {
	tcell.EventTime
	target Model
}

func SetMouseCapture(target Model) Cmd {
	return func() Msg {
		return &setMouseCaptureMsg{target: target}
	}
}

type setTitleMsg struct {
	tcell.EventTime
	title string
}

func SetTitle(title string) Cmd {
	return func() Msg {
		return &setTitleMsg{title: title}
	}
}

type getClipboardMsg struct{ tcell.EventTime }

func GetClipboard() Cmd {
	return func() Msg {
		return &getClipboardMsg{}
	}
}

type setClipboardMsg struct {
	tcell.EventTime
	data []byte
}

func SetClipboard(data []byte) Cmd {
	return func() Msg {
		return &setClipboardMsg{data: data}
	}
}

type notifyMsg struct {
	tcell.EventTime
	title, body string
}

func Notify(title, body string) Cmd {
	return func() Msg {
		return &notifyMsg{title: title, body: body}
	}
}
