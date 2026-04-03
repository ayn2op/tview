package tview

import "github.com/gdamore/tcell/v3"

type Event = tcell.Event

// Cmd is a side effect requested by a model during input handling.
type Cmd func() Event

type batchEvent struct {
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
		return func() Event {
			return &batchEvent{cmds: valid}
		}
	}
}

type InitEvent struct{ tcell.EventTime }

type KeyEvent = tcell.EventKey

type MouseEvent struct {
	tcell.EventMouse
	Action MouseAction
}

func newMouseEvent(mouseEvent tcell.EventMouse, action MouseAction) *MouseEvent {
	return &MouseEvent{mouseEvent, action}
}

type PasteEvent struct {
	tcell.EventTime
	Content string
}

func newPasteEvent(content string) *PasteEvent {
	return &PasteEvent{Content: content}
}

type quitEvent struct{ tcell.EventTime }

func Quit() Cmd {
	return func() Event {
		return &quitEvent{}
	}
}

type setFocusEvent struct {
	tcell.EventTime
	target Model
}

func SetFocus(target Model) Cmd {
	return func() Event {
		return &setFocusEvent{target: target}
	}
}

type setMouseCaptureEvent struct {
	tcell.EventTime
	target Model
}

func SetMouseCapture(target Model) Cmd {
	return func() Event {
		return &setMouseCaptureEvent{target: target}
	}
}

type setTitleEvent struct {
	tcell.EventTime
	title string
}

func SetTitle(title string) Cmd {
	return func() Event {
		return &setTitleEvent{title: title}
	}
}

type getClipboardEvent struct{ tcell.EventTime }

func GetClipboard() Cmd {
	return func() Event {
		return &getClipboardEvent{}
	}
}

type setClipboardEvent struct {
	tcell.EventTime
	data []byte
}

func SetClipboard(data []byte) Cmd {
	return func() Event {
		return &setClipboardEvent{data: data}
	}
}

type notifyEvent struct {
	tcell.EventTime
	title, body string
}

func Notify(title, body string) Cmd {
	return func() Event {
		return &notifyEvent{title: title, body: body}
	}
}
