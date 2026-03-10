package tview

import "github.com/gdamore/tcell/v3"

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

type InitEvent struct{ tcell.EventTime }

func NewInitEvent() *InitEvent {
	event := &InitEvent{}
	event.SetEventNow()
	return event
}
