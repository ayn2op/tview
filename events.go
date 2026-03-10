package tview

import "github.com/gdamore/tcell/v3"

type InitEvent struct{ tcell.EventTime }

func NewInitEvent() *InitEvent {
	event := &InitEvent{}
	event.SetEventNow()
	return event
}

type quitEvent struct{ tcell.EventTime }

func Quit() Command {
	return EventCommand(func() tcell.Event {
		event := &quitEvent{}
		event.SetEventNow()
		return event
	})
}

type KeyEvent = tcell.EventKey

type MouseEvent struct {
	tcell.EventMouse
	Action MouseAction
}

func newMouseEvent(mouseEvent tcell.EventMouse, action MouseAction) *MouseEvent {
	event := &MouseEvent{mouseEvent, action}
	return event
}

type PasteEvent struct {
	tcell.EventTime
	Content string
}

func newPasteEvent(content string) *PasteEvent {
	event := &PasteEvent{Content: content}
	event.SetEventNow()
	return event
}

type setTitleEvent struct {
	tcell.EventTime
	title string
}

func SetTitle(title string) Command {
	return EventCommand(func() tcell.Event {
		event := &setTitleEvent{title: title}
		event.SetEventNow()
		return event
	})
}

type getClipboardEvent struct{ tcell.EventTime }

func GetClipboard() Command {
	return EventCommand(func() tcell.Event {
		event := &getClipboardEvent{}
		event.SetEventNow()
		return event
	})
}

type setClipboardEvent struct {
	tcell.EventTime
	data []byte
}

func SetClipboard(data []byte) Command {
	return EventCommand(func() tcell.Event {
		event := &setClipboardEvent{data: data}
		event.SetEventNow()
		return event
	})
}
