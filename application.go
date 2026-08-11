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

// DoubleClickInterval specifies the maximum time between clicks to register a double-click rather than click.
var DoubleClickInterval = 500 * time.Millisecond

// MouseAction indicates one of the actions the mouse is logically doing.
type MouseAction int16

// Available mouse actions.
const (
	MouseMove MouseAction = iota
	MouseLeftDown
	MouseLeftUp
	MouseLeftClick
	MouseLeftDoubleClick
	MouseMiddleDown
	MouseMiddleUp
	MouseMiddleClick
	MouseMiddleDoubleClick
	MouseRightDown
	MouseRightUp
	MouseRightClick
	MouseRightDoubleClick
	MouseScrollUp
	MouseScrollDown
	MouseScrollLeft
	MouseScrollRight
)

type ApplicationOption func(*Application)

func WithScreen(screen tcell.Screen) ApplicationOption {
	return func(a *Application) {
		a.screen = screen
		a.forceRedraw = true
	}
}

func WithoutCatchPanics() ApplicationOption {
	return func(a *Application) {
		a.disableCatchPanics = true
	}
}

// Application represents the top node of an application.
//
// It is not strictly required to use this class as none of the other classes
// depend on it. However, it provides useful tools to set up an application and
// plays nicely with all widgets.
type Application struct {
	msgs     chan Msg
	cmds     chan Cmd
	done     chan struct{}
	doneOnce sync.Once

	root  Model
	focus Model

	mouseCapturingModel    Model            // A model requested to capture future mouse messages.
	lastMouseX, lastMouseY int              // The last position of the mouse.
	mouseDownX, mouseDownY int              // The position of the mouse when a button was last pressed.
	lastMouseClick         time.Time        // The time when a mouse button was last clicked.
	lastMouseButtons       tcell.ButtonMask // The last mouse button state.

	// forceRedraw requests a full clear before the next frame.
	forceRedraw bool

	screen             tcell.Screen
	disableCatchPanics bool
}

// NewApplication creates an application with root as its top-level model.
func NewApplication(root Model, options ...ApplicationOption) *Application {
	a := &Application{
		msgs: make(chan Msg),
		cmds: make(chan Cmd),
		done: make(chan struct{}),
		root: root,
	}
	for _, option := range options {
		option(a)
	}
	return a
}

// Run starts the application and thus the messages loop.
func (a *Application) Run() error {
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
		a.screen = screen
	}

	defer a.stop()

	go a.handleEvents()
	go a.handleCmds()

	root := a.root
	if root != nil {
		a.setFocus(root)
		terminalName, terminalVersion := a.screen.Terminal()
		a.queueCmd(root.Update(InitMsg{
			TerminalName:    terminalName,
			TerminalVersion: terminalVersion,
		}))
		a.draw()
	}

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
				data := fmt.Append(nil, msg.msg)
				_, _ = tty.Write(data)
			}
		case setFocusMsg:
			a.setFocus(msg.target)
		case suspendMsg:
			var next Msg
			a.suspend(func() { next = Cmd(msg)() })
			if next != nil {
				a.queueCmd(func() Msg { return next })
			}
		case setMouseCaptureMsg:
			a.mouseCapturingModel = msg.target
		case setTitleMsg:
			a.screen.SetTitle(string(msg))
		case notifyMsg:
			a.screen.ShowNotification(msg.title, msg.body)

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

			// Pass other key events to the root model.
			root := a.root
			if root != nil {
				a.queueCmd(root.Update(msg))
			}
		case *tcell.EventPaste:
			if msg.Start() {
				pasting = true
				pasteBuffer.Reset()
			} else if msg.End() {
				pasting = false
				root := a.root
				if root != nil && pasteBuffer.Len() > 0 {
					a.queueCmd(root.Update(PasteMsg(pasteBuffer.String())))
				}
			}
		case *tcell.EventResize:
			// Resize events can imply terminal state changes even when size
			// reports unchanged, so force one redraw pass.
			a.forceRedraw = true
			if time.Since(lastRedraw) < redrawPause {
				if redrawTimer != nil {
					redrawTimer.Stop()
				}
				redrawTimer = time.AfterFunc(redrawPause, func() {
					a.queueMsg(msg)
				})
			}
			lastRedraw = time.Now()
			if root := a.root; root != nil {
				a.queueCmd(root.Update(msg))
			}
		case *tcell.EventMouse:
			isMouseDownAction := a.fireMouseActions(msg)
			a.lastMouseButtons = msg.Buttons()
			if isMouseDownAction {
				a.mouseDownX, a.mouseDownY = msg.Position()
			}
		default:
			root := a.root
			if root != nil {
				a.queueCmd(root.Update(msg))
			}
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

func (a *Application) handleEvents() {
	for event := range a.screen.EventQ() {
		a.queueMsg(event)
	}
}

func (a *Application) handleCmds() {
	for {
		select {
		case <-a.done:
			return
		case cmd := <-a.cmds:
			go a.execCmd(cmd)
		}
	}
}

func (a *Application) execCmd(cmd Cmd) {
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

func (a *Application) execSequenceMsg(msg sequenceMsg) {
	for _, cmd := range msg {
		a.execCmd(cmd)
	}
}

func (a *Application) execBatchMsg(msg batchMsg) {
	var wg sync.WaitGroup
	for _, cmd := range msg {
		wg.Go(func() {
			a.execCmd(cmd)
		})
	}
	wg.Wait()
}

// fireMouseActions analyzes the provided mouse event, derives mouse actions
// from it and then forwards them to the corresponding models.
func (a *Application) fireMouseActions(event *tcell.EventMouse) (isMouseDownAction bool) {
	// We want to relay follow-up events to the same target model.
	var targetPrimitive Model

	// Helper function to fire a mouse action.
	fire := func(action MouseAction) {
		switch action {
		case MouseLeftDown, MouseMiddleDown, MouseRightDown:
			isMouseDownAction = true
		}

		// Determine the target model.
		var model Model
		if a.mouseCapturingModel != nil {
			model = a.mouseCapturingModel
			targetPrimitive = a.mouseCapturingModel
		} else if targetPrimitive != nil {
			model = targetPrimitive
		} else {
			model = a.root
		}
		if model != nil {
			a.queueCmd(model.Update(MouseMsg{EventMouse: event, Action: action}))
		}
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
		button                  tcell.ButtonMask
		down, up, click, dclick MouseAction
	}{
		{tcell.ButtonPrimary, MouseLeftDown, MouseLeftUp, MouseLeftClick, MouseLeftDoubleClick},
		{tcell.ButtonMiddle, MouseMiddleDown, MouseMiddleUp, MouseMiddleClick, MouseMiddleDoubleClick},
		{tcell.ButtonSecondary, MouseRightDown, MouseRightUp, MouseRightClick, MouseRightDoubleClick},
	} {
		if buttonChanges&buttonMsg.button != 0 {
			if buttons&buttonMsg.button != 0 {
				fire(buttonMsg.down)
			} else {
				fire(buttonMsg.up)
				if !clickMoved {
					if a.lastMouseClick.Add(DoubleClickInterval).Before(time.Now()) {
						fire(buttonMsg.click)
						a.lastMouseClick = time.Now()
					} else {
						fire(buttonMsg.dclick)
						a.lastMouseClick = time.Time{} // reset
					}
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
func (a *Application) stop() {
	a.doneOnce.Do(func() {
		a.screen.Fini()
		a.screen = nil
		close(a.done)
	})
}

func (a *Application) suspend(f func()) {
	screen := a.screen
	if screen.Suspend() != nil {
		return
	}
	f()
	screen.Resume()
}

func (a *Application) draw() {
	screen := a.screen
	root := a.root

	if root == nil {
		return
	}

	drawWidth, drawHeight := screen.Size()
	root.SetRect(0, 0, drawWidth, drawHeight)

	// tcell.Show emits only visual deltas; clear only when forced.
	if a.forceRedraw {
		screen.Clear()
	}
	root.View(screen)
	screen.Show()

	a.forceRedraw = false
}

func (a *Application) setFocus(m Model) {
	previous := a.focus
	a.focus = m
	a.screen.HideCursor()
	var blur Cmd
	if previous != nil {
		blur = previous.Update(BlurMsg{})
	}
	var focus Cmd
	if m != nil {
		focus = m.Update(FocusMsg{})
	}
	a.queueCmd(Sequence(blur, focus))
}

func (a *Application) queueMsg(msg Msg) {
	if msg == nil {
		return
	}
	select {
	case <-a.done:
	case a.msgs <- msg:
	}
}

func (a *Application) queueCmd(cmd Cmd) {
	if cmd == nil {
		return
	}
	select {
	case <-a.done:
	case a.cmds <- cmd:
	}
}
