package tview

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell/v3"
)

const (
	// The minimum time between two consecutive redraws.
	redrawPause = 50 * time.Millisecond
)

// MouseAction indicates one of the actions the mouse is logically doing.
type MouseAction int16

// Available mouse actions.
const (
	MouseMove MouseAction = iota
	MouseLeftDown
	MouseLeftUp
	MouseLeftClick
	MouseMiddleDown
	MouseMiddleUp
	MouseMiddleClick
	MouseRightDown
	MouseRightUp
	MouseRightClick
	MouseScrollUp
	MouseScrollDown
	MouseScrollLeft
	MouseScrollRight
)

type applicationOptions struct {
	screen             Screen
	disableCatchPanics bool
}

type ApplicationOption func(*applicationOptions)

func WithScreen(screen Screen) ApplicationOption {
	return func(c *applicationOptions) {
		c.screen = screen
	}
}

func WithoutCatchPanics() ApplicationOption {
	return func(c *applicationOptions) {
		c.disableCatchPanics = true
	}
}

// Application runs a model: it starts it, changes it with each message, and draws it.
type Application[M Model[M]] struct {
	msgs     chan Msg
	cmds     chan Cmd
	done     chan struct{}
	doneOnce sync.Once

	model M

	lastMouseX, lastMouseY int              // The last position of the mouse.
	mouseDownX, mouseDownY int              // The position of the mouse when a button was last pressed.
	lastMouseButtons       tcell.ButtonMask // The last mouse button state.

	screen             Screen
	disableCatchPanics bool
}

// NewApplication creates an application that runs model.
func NewApplication[M Model[M]](model M, options ...ApplicationOption) *Application[M] {
	var opts applicationOptions
	for _, option := range options {
		option(&opts)
	}
	return &Application[M]{
		msgs:  make(chan Msg),
		cmds:  make(chan Cmd),
		done:  make(chan struct{}),
		model: model,

		screen:             opts.screen,
		disableCatchPanics: opts.disableCatchPanics,
	}
}

// Run starts the application and thus the messages loop.
func (a *Application[M]) Run() error {
	var (
		lastRedraw  time.Time   // The time the screen was last redrawn.
		redrawTimer *time.Timer // A timer to schedule the next redraw.
	)

	// Make a screen if there is none yet.
	if a.screen == nil {
		screen, err := tcell.NewScreen()
		if err != nil {
			return err
		}
		if err = screen.Init(); err != nil {
			return err
		}
		screen.EnableMouse()
		screen.EnablePaste()
		screen.EnableFocus()
		a.screen = screen
	}
	defer a.stop()

	go a.handleEvents()
	go a.handleCmds()

	a.queueCmd(a.model.Init())
	a.draw()

	var (
		pasteBuffer strings.Builder
		pasting     bool // Set to true while we receive paste key events.
	)
	for msg := range a.msgs {
		if msg == nil {
			continue
		}
		switch msg := msg.(type) {
		case quitMsg:
			return nil
		case *tcell.EventError:
			return msg

		case rawMsg:
			if tty, ok := a.screen.Tty(); ok {
				data := fmt.Append(nil, msg.data)
				_, _ = tty.Write(data)
			}
		case suspendMsg:
			var next Msg
			a.suspend(func() { next = Cmd(msg)() })
			if next != nil {
				a.queueCmd(func() Msg { return next })
			}
		case setTitleMsg:
			a.screen.SetTitle(string(msg))
		case notifyMsg:
			a.screen.ShowNotification(msg.title, msg.body)

		case requestTerminalInfoMsg:
			name, version := a.screen.Terminal()
			a.handle(TerminalInfoMsg{Name: name, Version: version})

		case getClipboardMsg:
			a.screen.GetClipboard()
		case setClipboardMsg:
			a.screen.SetClipboard([]byte(msg))

		case KeyMsg:
			// If we are pasting, collect runes, nothing else.
			if pasting {
				appendPasteKey(&pasteBuffer, msg)
				break
			}
			a.handle(msg)
		case *tcell.EventPaste:
			if msg.Start() {
				pasting = true
				pasteBuffer.Reset()
			} else if msg.End() {
				pasting = false
				if pasteBuffer.Len() > 0 {
					a.handle(PasteMsg(pasteBuffer.String()))
				}
			}
		case *tcell.EventResize:
			if time.Since(lastRedraw) < redrawPause {
				if redrawTimer != nil {
					redrawTimer.Stop()
				}
				redrawTimer = time.AfterFunc(redrawPause, func() {
					a.queueMsg(msg)
				})
			}
			lastRedraw = time.Now()
			a.updateModel(msg)
		case *tcell.EventMouse:
			isMouseDownAction := a.fireMouseActions(msg)
			a.lastMouseButtons = msg.Buttons()
			if isMouseDownAction {
				a.mouseDownX, a.mouseDownY = msg.Position()
			}
		default:
			// Messages from commands reach the widget too, so widgets can handle ones like tree.ActionMsg.
			a.handle(msg)
		}

		a.draw()
	}
	return nil
}

func appendPasteKey(buffer *strings.Builder, msg KeyMsg) {
	switch msg.Key() {
	case tcell.KeyRune:
		buffer.WriteString(msg.Str())
	case tcell.KeyEnter, tcell.KeyCtrlJ:
		buffer.WriteRune('\n')
	case tcell.KeyTab:
		buffer.WriteRune('\t')
	}
}

func (a *Application[M]) handleEvents() {
	for event := range a.screen.EventQ() {
		a.queueMsg(event)
	}
}

func (a *Application[M]) handleCmds() {
	for {
		select {
		case <-a.done:
			return
		case cmd := <-a.cmds:
			go a.execCmd(cmd)
		}
	}
}

func (a *Application[M]) execCmd(cmd Cmd) {
	if !a.disableCatchPanics {
		defer func() {
			if r := recover(); r != nil {
				text := fmt.Sprintf("goroutine panicked: %v", r)
				fmt.Fprintf(os.Stderr, "%s\nstack trace:\n%s\n", text, debug.Stack())
				a.queueMsg(tcell.NewEventError(errors.New(text)))
			}
		}()
	}

	switch msg := cmd().(type) {
	case batchMsg:
		a.execBatchMsg(msg)
	case sequenceMsg:
		a.execSequenceMsg(msg)
	default:
		a.queueMsg(msg)
	}
}

func (a *Application[M]) execSequenceMsg(msg sequenceMsg) {
	for _, cmd := range msg {
		a.execCmd(cmd)
	}
}

func (a *Application[M]) execBatchMsg(msg batchMsg) {
	var wg sync.WaitGroup
	for _, cmd := range msg {
		wg.Go(func() {
			a.execCmd(cmd)
		})
	}
	wg.Wait()
}

// fireMouseActions derives mouse actions from the provided mouse event and passes them through the root's widget.
func (a *Application[M]) fireMouseActions(event *tcell.EventMouse) (isMouseDownAction bool) {
	fire := func(action MouseAction) {
		switch action {
		case MouseLeftDown, MouseMiddleDown, MouseRightDown:
			isMouseDownAction = true
		}
		a.handle(MouseMsg{EventMouse: event, Action: action})
	}

	x, y := event.Position()
	buttons := event.Buttons()
	clickMoved := x != a.mouseDownX || y != a.mouseDownY
	buttonChanges := buttons ^ a.lastMouseButtons

	if x != a.lastMouseX || y != a.lastMouseY {
		fire(MouseMove)
		a.lastMouseX = x
		a.lastMouseY = y
	}

	for _, buttonMsg := range []struct {
		button          tcell.ButtonMask
		down, up, click MouseAction
	}{
		{tcell.ButtonPrimary, MouseLeftDown, MouseLeftUp, MouseLeftClick},
		{tcell.ButtonMiddle, MouseMiddleDown, MouseMiddleUp, MouseMiddleClick},
		{tcell.ButtonSecondary, MouseRightDown, MouseRightUp, MouseRightClick},
	} {
		if buttonChanges&buttonMsg.button != 0 {
			if buttons&buttonMsg.button != 0 {
				fire(buttonMsg.down)
			} else {
				fire(buttonMsg.up)
				if !clickMoved {
					fire(buttonMsg.click)
				}
			}
		}
	}

	for _, wheelMsg := range []struct {
		button tcell.ButtonMask
		action MouseAction
	}{
		{tcell.WheelUp, MouseScrollUp},
		{tcell.WheelDown, MouseScrollDown},
		{tcell.WheelLeft, MouseScrollLeft},
		{tcell.WheelRight, MouseScrollRight}} {
		if buttons&wheelMsg.button != 0 {
			fire(wheelMsg.action)
		}
	}

	return isMouseDownAction
}

// stop finalizes the active screen and leaves terminal UI mode.
func (a *Application[M]) stop() {
	a.doneOnce.Do(func() {
		a.screen.Fini()
		a.screen = nil
		close(a.done)
	})
}

func (a *Application[M]) suspend(f func()) {
	screen := a.screen
	if screen.Suspend() != nil {
		return
	}
	f()
	screen.Resume()
}

func (a *Application[M]) draw() {
	screen := a.screen
	drawWidth, drawHeight := screen.Size()

	// Each frame starts blank, so nothing a widget leaves undrawn shows the frame before. Show still emits only the cells that changed.
	screen.Clear()
	// Each frame starts without a cursor, so only a widget drawn in it can show one.
	screen.HideCursor()
	if view := a.model.View(); view != nil {
		view.Draw(screen, Rectangle{Width: drawWidth, Height: drawHeight})
	}
	screen.Show()
}

// handle passes a message through the model's widget before updating the model with it.
func (a *Application[M]) handle(msg Msg) {
	if view := a.model.View(); view != nil {
		width, height := a.screen.Size()
		msg = view.Handle(msg, Rectangle{Width: width, Height: height})
	}
	if msg != nil {
		a.updateModel(msg)
	}
}

// updateModel replaces the model with the one its Update returns for msg and queues the returned command.
func (a *Application[M]) updateModel(msg Msg) {
	var cmd Cmd
	a.model, cmd = a.model.Update(msg)
	a.queueCmd(cmd)
}

func (a *Application[M]) queueMsg(msg Msg) {
	if msg == nil {
		return
	}
	select {
	case <-a.done:
	case a.msgs <- msg:
	}
}

func (a *Application[M]) queueCmd(cmd Cmd) {
	if cmd == nil {
		return
	}
	select {
	case <-a.done:
	case a.cmds <- cmd:
	}
}
