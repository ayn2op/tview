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

// DoubleClickInterval specifies the maximum time between clicks to register a
// double click rather than click.
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

// Application represents the top node of an application.
//
// It is not strictly required to use this class as none of the other classes
// depend on it. However, it provides useful tools to set up an application and
// plays nicely with all widgets.
type Application struct {
	sync.RWMutex

	// The application's screen. Apart from Run(), this variable should never be
	// set directly. Always use the screenReplacement channel after calling
	// Fini(), to set a new screen (or nil to stop the application).
	screen tcell.Screen

	// The model which currently has the keyboard focus.
	focus Model

	// The root model to be seen on the screen.
	root Model

	msgs chan Msg
	cmds chan Cmd

	mouseCapturingModel    Model            // A model requested via SetMouseCaptureCommand to capture future mouse messages.
	lastMouseX, lastMouseY int              // The last position of the mouse.
	mouseDownX, mouseDownY int              // The position of the mouse when its button was last pressed.
	lastMouseClick         time.Time        // The time when a mouse button was last clicked.
	lastMouseButtons       tcell.ButtonMask // The last mouse button state.

	// forceRedraw requests a full clear before the next frame.
	forceRedraw bool

	catchPanics bool
}

// NewApplication creates and returns a new application.
func NewApplication() *Application {
	return &Application{
		cmds:        make(chan Cmd),
		catchPanics: true,
	}
}

// SetScreen sets the application's screen.
func (a *Application) SetScreen(screen tcell.Screen) *Application {
	a.Lock()
	defer a.Unlock()
	if a.screen == nil {
		a.screen = screen
		a.forceRedraw = true
	}
	return a
}

// SetCatchPanics sets whether cmd panics should be recovered.
func (a *Application) SetCatchPanics(catchPanics bool) *Application {
	a.Lock()
	defer a.Unlock()
	a.catchPanics = catchPanics
	return a
}

// Run starts the application and thus the messages loop.
func (a *Application) Run() error {
	var (
		lastRedraw  time.Time   // The time the screen was last redrawn.
		redrawTimer *time.Timer // A timer to schedule the next redraw.
	)

	// Make a screen if there is none yet.
	a.Lock()
	if a.screen == nil {
		screen, err := tcell.NewScreen()
		if err != nil {
			a.Unlock()
			return err
		}
		if err = screen.Init(); err != nil {
			a.Unlock()
			return err
		}
		a.screen = screen
	}
	a.Unlock()

	defer a.stop()

	// We catch panics to clean up because they mess up the terminal.
	defer func() {
		if p := recover(); p != nil {
			panic(p)
		}
	}()

	go a.handleCmds()

	a.Lock()
	a.msgs = a.screen.EventQ()
	a.Unlock()

	a.RLock()
	root := a.root
	a.RUnlock()

	if root != nil {
		if cmd := root.Update(&InitMsg{}); cmd != nil {
			a.cmds <- cmd
		}
		a.draw()
	}

	// Start messages loop.
	var (
		pasteBuffer strings.Builder
		pasting     bool // Set to true while we receive paste key events.
	)
	for msg := range a.msgs {
		if msg == nil {
			continue
		}
		switch msg := msg.(type) {
		case *quitMsg:
			return nil
		case *tcell.EventError:
			return msg

		case *batchMsg:
			for _, cmd := range msg.cmds {
				if cmd != nil {
					a.cmds <- cmd
				}
			}

		case *setFocusMsg:
			a.setFocus(msg.target)
		case *setMouseCaptureMsg:
			a.mouseCapturingModel = msg.target
		case *setTitleMsg:
			a.screen.SetTitle(msg.title)
		case *notifyMsg:
			a.screen.ShowNotification(msg.title, msg.body)

		case *getClipboardMsg:
			a.screen.GetClipboard()
		case *setClipboardMsg:
			a.screen.SetClipboard(msg.data)

		case *KeyMsg:
			// If we are pasting, collect runes, nothing else.
			if pasting {
				switch msg.Key() {
				case tcell.KeyRune:
					pasteBuffer.WriteString(msg.Str())
				case tcell.KeyEnter:
					pasteBuffer.WriteRune('\n')
				case tcell.KeyTab:
					pasteBuffer.WriteRune('\t')
				}
				break
			}

			a.RLock()
			root := a.root
			a.RUnlock()

			// Pass other key events to the root model.
			if root != nil && root.HasFocus() {
				if cmd := root.Update(msg); cmd != nil {
					a.cmds <- cmd
				}
			}
		case *tcell.EventPaste:
			if msg.Start() {
				pasting = true
				pasteBuffer.Reset()
			} else if msg.End() {
				pasting = false
				a.RLock()
				root := a.root
				a.RUnlock()
				if root != nil && root.HasFocus() && pasteBuffer.Len() > 0 {
					if cmd := root.Update(&PasteMsg{Content: pasteBuffer.String()}); cmd != nil {
						a.cmds <- cmd
					}
				}
			}
		case *tcell.EventResize:
			a.Lock()
			// Resize events can imply terminal state changes even when size
			// reports unchanged, so force one redraw pass.
			a.forceRedraw = true
			a.Unlock()
			if time.Since(lastRedraw) < redrawPause {
				if redrawTimer != nil {
					redrawTimer.Stop()
				}
				redrawTimer = time.AfterFunc(redrawPause, func() {
					a.Send(msg)
				})
			}
			lastRedraw = time.Now()
		case *tcell.EventMouse:
			isMouseDownAction := a.fireMouseActions(msg)
			a.lastMouseButtons = msg.Buttons()
			if isMouseDownAction {
				a.mouseDownX, a.mouseDownY = msg.Position()
			}
		default:
			a.RLock()
			root := a.root
			a.RUnlock()
			if root != nil {
				if cmd := root.Update(msg); cmd != nil {
					a.cmds <- cmd
				}
			}
		}

		a.draw()
	}
	return nil
}

func (a *Application) handleCmds() {
	for cmd := range a.cmds {
		if cmd == nil {
			continue
		}

		go func() {
			if a.catchPanics {
				defer func() {
					if r := recover(); r != nil {
						text := fmt.Sprintf("goroutine panicked: %v", r)
						fmt.Fprintf(os.Stderr, "%s\nstack trace:\n%s\n", text, debug.Stack())
						a.Send(tcell.NewEventError(errors.New(text)))
					}
				}()
			}
			if msg := cmd(); msg != nil {
				a.Send(msg)
			}
		}()
	}
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
			if cmd := model.Update(&MouseMsg{EventMouse: *event, Action: action}); cmd != nil {
				a.cmds <- cmd
			}
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
	a.Lock()
	defer a.Unlock()
	screen := a.screen
	if screen == nil {
		return
	}
	screen.Fini()
	a.screen = nil
	close(a.cmds)
}

// Suspend temporarily suspends the application by exiting terminal UI mode and
// invoking the provided function "f". When "f" returns, terminal UI mode is
// entered again and the application resumes.
//
// A return value of true indicates that the application was suspended and "f"
// was called. If false is returned, the application was already suspended,
// terminal UI mode was not exited, and "f" was not called.
func (a *Application) Suspend(f func()) bool {
	a.RLock()
	screen := a.screen
	a.RUnlock()
	if screen == nil {
		return false // Screen has not yet been initialized.
	}

	// Enter suspended mode.
	if err := screen.Suspend(); err != nil {
		return false // Suspension failed.
	}

	// Wait for "f" to return.
	f()

	// If the screen object has changed in the meantime, we need to do more.
	a.RLock()
	defer a.RUnlock()
	if a.screen != screen {
		// Calling stop() while in suspend mode currently still leads to a
		// panic, see https://github.com/gdamore/tcell/issues/440.
		screen.Fini()
		if a.screen == nil {
			return true // If stop was called (a.screen is nil), we're done already.
		}
	} else {
		// It hasn't changed. Resume.
		screen.Resume() // Not much we can do in case of an error.
	}

	// Continue application loop.
	return true
}

// draw actually does what Draw() promises to do.
func (a *Application) draw() *Application {
	a.RLock()
	screen := a.screen
	root := a.root
	forceRedraw := a.forceRedraw
	a.RUnlock()

	// Maybe we're not ready yet or not anymore.
	if screen == nil || root == nil {
		return a
	}

	drawWidth, drawHeight := screen.Size()
	root.SetRect(0, 0, drawWidth, drawHeight)

	// tcell already keeps a logical back buffer and emits only visual deltas in
	// Show(). Avoid clearing on regular redraws so we don't rewrite the full
	// logical screen every frame; keep full clears for forced redraws.
	if forceRedraw {
		screen.Clear()
	}
	root.Draw(screen)
	screen.Show()

	a.Lock()
	a.forceRedraw = false
	a.Unlock()

	return a
}

// SetRoot sets the root model for this application. This function must be called at least once or nothing will be displayed when
// the application starts.
//
// It also calls SetFocus() on the model.
func (a *Application) SetRoot(root Model) *Application {
	a.Lock()
	a.root = root
	if a.screen != nil {
		a.forceRedraw = true
	}
	a.Unlock()

	a.setFocus(root)
	return a
}

// setFocus sets the focus to a new model. All key events will be directed
// down the hierarchy (starting at the root) until a model handles them,
// which per default goes towards the focused model.
//
// Blur() will be called on the previously focused model. Focus() will be
// called on the new model.
func (a *Application) setFocus(m Model) *Application {
	a.Lock()
	if a.focus != nil {
		a.focus.Blur()
	}
	a.focus = m
	if a.screen != nil {
		a.screen.HideCursor()
	}
	a.Unlock()
	if m != nil {
		m.Focus(func(m Model) {
			a.setFocus(m)
		})
	}

	return a
}

// Focused returns the model which has the current focus. If none has it,
// nil is returned.
func (a *Application) Focused() Model {
	a.RLock()
	defer a.RUnlock()
	return a.focus
}

// Send sends a message to the internal messages loop.
func (a *Application) Send(msg Msg) *Application {
	a.RLock()
	msgs := a.msgs
	a.RUnlock()
	if msgs == nil {
		return a
	}
	msgs <- msg
	return a
}
