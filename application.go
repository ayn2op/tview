package tview

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/ayn2op/tview/cmd"
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
	queueSize          int
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

// WithQueueSize sets how many messages can be queued before their senders block, 256 by default.
// A burst of messages that fits is handled before drawing, so it is drawn once rather than a frame each.
func WithQueueSize(size int) ApplicationOption {
	return func(c *applicationOptions) {
		c.queueSize = size
	}
}

// Application runs a model: it starts it, changes it with each message, and draws it.
type Application[M Model[M]] struct {
	msgs     chan Msg
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
	opts := applicationOptions{queueSize: 256}
	for _, option := range options {
		option(&opts)
	}
	return &Application[M]{
		msgs:  make(chan Msg, opts.queueSize),
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

	if a.handleCmd(a.model.Init()) {
		return nil
	}
	a.draw()

	var (
		pasteBuffer strings.Builder
		pasting     bool // Set to true while we receive paste key events.
		// Only an updated model changes the view, so a frame is drawn after an update and skipped otherwise.
		redraw bool
	)
	for msg := range a.msgs {
		if msg == nil {
			continue
		}
		var updated, quit bool
		switch msg := msg.(type) {
		case *tcell.EventError:
			return msg
		case Cmd:
			quit = a.handleCmd(msg)

		case KeyMsg:
			// If we are pasting, collect runes, nothing else.
			if pasting {
				appendPasteKey(&pasteBuffer, msg)
				break
			}
			updated, quit = a.handle(msg)
		case *tcell.EventPaste:
			if msg.Start() {
				pasting = true
				pasteBuffer.Reset()
			} else if msg.End() {
				pasting = false
				if pasteBuffer.Len() > 0 {
					updated, quit = a.handle(PasteMsg(pasteBuffer.String()))
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
			updated, quit = true, a.updateModel(msg)
		case *tcell.EventMouse:
			updated, quit = a.fireMouseActions(msg)
			a.lastMouseButtons = msg.Buttons()
		default:
			// Messages from commands reach the widget too.
			updated, quit = a.handle(msg)
		}
		if quit {
			return nil
		}
		redraw = redraw || updated

		// Queued events and messages are handled before drawing, so a burst of them, such as scrolling with a touchpad or images loading, is drawn once rather than a frame each, which would fall behind them.
		if !redraw || len(a.screen.EventQ()) > 0 || len(a.msgs) > 0 {
			continue
		}
		a.draw()
		redraw = false
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

// fireMouseActions derives mouse actions from the provided mouse event and passes them through the root's widget.
func (a *Application[M]) fireMouseActions(event *tcell.EventMouse) (updated, quit bool) {
	x, y := event.Position()
	fire := func(action MouseAction) {
		if quit {
			return
		}
		switch action {
		case MouseLeftDown, MouseMiddleDown, MouseRightDown:
			a.mouseDownX, a.mouseDownY = x, y
		}
		var u bool
		u, quit = a.handle(MouseMsg{EventMouse: event, Action: action})
		updated = updated || u
	}

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

	return updated, quit
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

// handle passes a message through the model's widget before updating the model with what it makes of it.
// Mouse input that no widget took is dropped, as no model reads it.
func (a *Application[M]) handle(msg Msg) (updated, quit bool) {
	if view := a.model.View(); view != nil {
		width, height := a.screen.Size()
		msg = view.Handle(msg, Rectangle{Width: width, Height: height})
	}
	if _, isMouse := msg.(MouseMsg); isMouse || msg == nil {
		return false, false
	}
	return true, a.updateModel(msg)
}

// updateModel replaces the model with the one its Update returns for msg and handles the returned command.
func (a *Application[M]) updateModel(msg Msg) (quit bool) {
	var c Cmd
	a.model, c = a.model.Update(msg)
	return a.handleCmd(c)
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

func (a *Application[M]) handleCmd(c Cmd) (quit bool) {
	switch c := c.(type) {
	case cmd.Quit:
		return true
	case cmd.WriteRaw:
		if tty, ok := a.screen.Tty(); ok {
			_, _ = tty.Write([]byte(c))
		}
	case cmd.Suspend:
		var next Msg
		a.suspend(func() { next = c() })
		a.queueMsg(next)
	case cmd.SetTitle:
		a.screen.SetTitle(string(c))
	case cmd.Notify:
		a.screen.ShowNotification(c.Title, c.Body)
	case cmd.RequestTerminalInfo:
		name, version := a.screen.Terminal()
		_, quit := a.handle(TerminalInfoMsg{Name: name, Version: version})
		return quit
	case cmd.GetClipboard:
		a.screen.GetClipboard()
	case cmd.SetClipboard:
		a.screen.SetClipboard(c)
	case cmd.Task:
		go a.runTask(c)
	case cmd.Batch:
		for _, c := range c {
			if a.handleCmd(c) {
				return true
			}
		}
	}
	return false
}

func (a *Application[M]) runTask(task cmd.Task) {
	if !a.disableCatchPanics {
		defer func() {
			if r := recover(); r != nil {
				text := fmt.Sprintf("goroutine panicked: %v", r)
				fmt.Fprintf(os.Stderr, "%s\nstack trace:\n%s\n", text, debug.Stack())
				a.queueMsg(tcell.NewEventError(errors.New(text)))
			}
		}()
	}
	a.queueMsg(task())
}
