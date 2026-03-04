package tview

import "github.com/gdamore/tcell/v3"

// Command is a side effect requested by a primitive during input handling.
// Commands are executed by the Application event loop.
type Command any

// BatchCommand groups multiple commands into a single command.
type BatchCommand []Command

type EventCommand func() tcell.Event

type SetFocusCommand struct {
	Target Primitive
}

type SetMouseCaptureCommand struct {
	Target Primitive
}

type RedrawCommand struct{}

type QuitCommand struct{}

type SetTitleCommand string

type SetClipboardCommand string

type GetClipboardCommand struct{}

type NotifyCommand struct{ Title, Body string }
