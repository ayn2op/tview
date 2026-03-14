package tview

import "github.com/gdamore/tcell/v3"

// Command is a side effect requested by a primitive during input handling.
// Commands are executed by the Application event loop.
type Command func() tcell.Event

type batchEvent struct {
	tcell.EventTime
	commands []Command
}

// Batch combines multiple commands into a single command.
func Batch(cmds ...Command) Command {
	var valid []Command
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
		return func() tcell.Event {
			return &batchEvent{commands: valid}
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

func Quit() Command {
	return func() tcell.Event {
		return &quitEvent{}
	}
}

type setFocusEvent struct {
	tcell.EventTime
	target Primitive
}

func SetFocus(target Primitive) Command {
	return func() tcell.Event {
		return &setFocusEvent{target: target}
	}
}

type setMouseCaptureEvent struct {
	tcell.EventTime
	target Primitive
}

func SetMouseCapture(target Primitive) Command {
	return func() tcell.Event {
		return &setMouseCaptureEvent{target: target}
	}
}

type setTitleEvent struct {
	tcell.EventTime
	title string
}

func SetTitle(title string) Command {
	return func() tcell.Event {
		return &setTitleEvent{title: title}
	}
}

type getClipboardEvent struct{ tcell.EventTime }

func GetClipboard() Command {
	return func() tcell.Event {
		return &getClipboardEvent{}
	}
}

type setClipboardEvent struct {
	tcell.EventTime
	data []byte
}

func SetClipboard(data []byte) Command {
	return func() tcell.Event {
		return &setClipboardEvent{data: data}
	}
}

type notifyEvent struct {
	tcell.EventTime
	title, body string
}

func Notify(title, body string) Command {
	return func() tcell.Event {
		return &notifyEvent{title: title, body: body}
	}
}
