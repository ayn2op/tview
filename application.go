package tview

import (
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell/v3"
)

const (
	// The size of the queued updates channel.
	updatesQueueSize = 100
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

// queuedUpdate represented the execution of f queued by
// Application.QueueUpdate(). If "done" is not nil, it receives exactly one
// element after f has executed.
type queuedUpdate struct {
	f    func()
	done chan struct{}
}

// Application represents the top node of an application.
//
// It is not strictly required to use this class as none of the other classes
// depend on it. However, it provides useful tools to set up an application and
// plays nicely with all widgets.
//
// The following command displays a primitive p on the screen until the
// application is stopped (for example via Quit()):
//
//	if err := tview.NewApplication().SetRoot(p, true).Run(); err != nil {
//	    panic(err)
//	}
type Application struct {
	sync.RWMutex

	// The application's screen. Apart from Run(), this variable should never be
	// set directly. Always use the screenReplacement channel after calling
	// Fini(), to set a new screen (or nil to stop the application).
	screen tcell.Screen

	// The primitive which currently has the keyboard focus.
	focus Primitive

	// The root primitive to be seen on the screen.
	root Primitive

	events chan tcell.Event

	// Functions queued from goroutines, used to serialize updates to primitives.
	updates chan queuedUpdate

	mouseCapturingPrimitive Primitive        // A primitive requested via SetMouseCaptureCommand to capture future mouse events.
	lastMouseX, lastMouseY  int              // The last position of the mouse.
	mouseDownX, mouseDownY  int              // The position of the mouse when its button was last pressed.
	lastMouseClick          time.Time        // The time when a mouse button was last clicked.
	lastMouseButtons        tcell.ButtonMask // The last mouse button state.

	// forceRedraw requests a full clear before the next frame.
	forceRedraw bool
}

// NewApplication creates and returns a new application.
func NewApplication() *Application {
	return &Application{
		updates: make(chan queuedUpdate, updatesQueueSize),
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

// Run starts the application and thus the event loop. This function returns
// when [Application.Stop] was called.
//
// Note that while an application is running, it fully claims stdin, stdout, and
// stderr. If you use these standard streams, they may not work as expected.
// Consider stopping the application first or suspending it (using
// [Application.Suspend]) if you have to interact with the standard streams, for
// example when needing to print a call stack during a panic.
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

	defer a.Stop()

	// We catch panics to clean up because they mess up the terminal.
	defer func() {
		if p := recover(); p != nil {
			panic(p)
		}
	}()

	a.Lock()
	a.events = a.screen.EventQ()
	a.Unlock()

	a.RLock()
	root := a.root
	a.RUnlock()

	if root != nil {
		if cmd := root.HandleEvent(NewInitEvent()); cmd != nil {
			_ = a.executeCommand(cmd)
		}
		a.draw()
	}

	// Start event loop.
	var (
		pasteBuffer strings.Builder
		pasting     bool // Set to true while we receive paste key events.
	)
EventLoop:
	for {
		select {
		// If we received an event, handle it.
		case event := <-a.events:
			if event == nil {
				break EventLoop
			}

			switch event := event.(type) {
			case *quitEvent:
				break EventLoop

			case *tcell.EventKey:
				// If we are pasting, collect runes, nothing else.
				if pasting {
					switch event.Key() {
					case tcell.KeyRune:
						pasteBuffer.WriteString(event.Str())
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

				// Pass other key events to the root primitive.
				if root != nil && root.HasFocus() {
					if cmd := root.HandleEvent(event); cmd != nil {
						if a.executeCommand(cmd) {
							a.draw()
						}
					}
				}
			case *tcell.EventPaste:
				if event.Start() {
					pasting = true
					pasteBuffer.Reset()
				} else if event.End() {
					pasting = false
					a.RLock()
					root := a.root
					a.RUnlock()
					if root != nil && root.HasFocus() && pasteBuffer.Len() > 0 {
						// Pass paste event to the root primitive.
						cmd := root.HandleEvent(newPasteEvent(pasteBuffer.String()))
						if a.executeCommand(cmd) {
							a.draw()
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
						a.QueueEvent(event)
					})
				}
				lastRedraw = time.Now()
				a.draw()
			case *tcell.EventMouse:
				handled, isMouseDownAction := a.fireMouseActions(event)
				if handled {
					a.draw()
				}
				a.lastMouseButtons = event.Buttons()
				if isMouseDownAction {
					a.mouseDownX, a.mouseDownY = event.Position()
				}
			default:
				a.RLock()
				root := a.root
				a.RUnlock()
				if root != nil {
					if cmd := root.HandleEvent(event); cmd != nil {
						if a.executeCommand(cmd) {
							a.draw()
						}
					}
				}
			}

		// If we have updates, now is the time to execute them.
		case update := <-a.updates:
			update.f()
			if update.done != nil {
				update.done <- struct{}{}
			}
		}
	}
	return nil
}

// fireMouseActions analyzes the provided mouse event, derives mouse actions
// from it and then forwards them to the corresponding primitives.
func (a *Application) fireMouseActions(event *tcell.EventMouse) (handled, isMouseDownAction bool) {
	// We want to relay follow-up events to the same target primitive.
	var targetPrimitive Primitive

	// Helper function to fire a mouse action.
	fire := func(action MouseAction) {
		switch action {
		case MouseLeftDown, MouseMiddleDown, MouseRightDown:
			isMouseDownAction = true
		}

		// Determine the target primitive.
		var primitive Primitive
		if a.mouseCapturingPrimitive != nil {
			primitive = a.mouseCapturingPrimitive
			targetPrimitive = a.mouseCapturingPrimitive
		} else if targetPrimitive != nil {
			primitive = targetPrimitive
		} else {
			primitive = a.root
		}
		if primitive != nil {
			cmd := primitive.HandleEvent(newMouseEvent(*event, action))
			if a.executeCommand(cmd) {
				handled = true
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

	for _, buttonEvent := range []struct {
		button                  tcell.ButtonMask
		down, up, click, dclick MouseAction
	}{
		{tcell.ButtonPrimary, MouseLeftDown, MouseLeftUp, MouseLeftClick, MouseLeftDoubleClick},
		{tcell.ButtonMiddle, MouseMiddleDown, MouseMiddleUp, MouseMiddleClick, MouseMiddleDoubleClick},
		{tcell.ButtonSecondary, MouseRightDown, MouseRightUp, MouseRightClick, MouseRightDoubleClick},
	} {
		if buttonChanges&buttonEvent.button != 0 {
			if buttons&buttonEvent.button != 0 {
				fire(buttonEvent.down)
			} else {
				fire(buttonEvent.up)
				if !clickMoved {
					if a.lastMouseClick.Add(DoubleClickInterval).Before(time.Now()) {
						fire(buttonEvent.click)
						a.lastMouseClick = time.Now()
					} else {
						fire(buttonEvent.dclick)
						a.lastMouseClick = time.Time{} // reset
					}
				}
			}
		}
	}

	for _, wheelEvent := range []struct {
		button tcell.ButtonMask
		action MouseAction
	}{
		{tcell.WheelUp, MouseScrollUp},
		{tcell.WheelDown, MouseScrollDown},
		{tcell.WheelLeft, MouseScrollLeft},
		{tcell.WheelRight, MouseScrollRight}} {
		if buttons&wheelEvent.button != 0 {
			fire(wheelEvent.action)
		}
	}

	return handled, isMouseDownAction
}

// Stop stops the application, causing Run() to return.
func (a *Application) Stop() {
	a.Lock()
	defer a.Unlock()
	screen := a.screen
	if screen == nil {
		return
	}
	screen.Fini()
	a.screen = nil
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
		// Calling Stop() while in suspend mode currently still leads to a
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

// Draw refreshes the screen (during the next update cycle). It calls the Draw()
// function of the application's root primitive and then syncs the screen
// buffer. It is almost never necessary to call this function. It can actually
// deadlock your application if you call it from the main thread (e.g. in a
// callback function of a widget). Please see
// https://github.com/ayn2op/tview/wiki/Concurrency for details.
func (a *Application) Draw() *Application {
	a.QueueUpdate(func() {
		a.draw()
	})
	return a
}

// ForceDraw refreshes the screen immediately. Use this function with caution as
// it may lead to race conditions with updates to primitives in other
// goroutines. It is always preferable to call [Application.Draw] instead.
// Never call this function from a goroutine.
//
// It is safe to call this function during queued updates and direct event
// handling.
func (a *Application) ForceDraw() *Application {
	return a.draw()
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

// Sync forces a full re-sync of the screen buffer with the actual screen during
// the next event cycle. This is useful for when the terminal screen is
// corrupted so you may want to offer your users a keyboard shortcut to refresh
// the screen.
func (a *Application) Sync() *Application {
	a.updates <- queuedUpdate{f: func() {
		a.Lock()
		screen := a.screen
		a.forceRedraw = true
		a.Unlock()
		if screen == nil {
			return
		}
		screen.Sync()
	}}
	return a
}

// SetRoot sets the root primitive for this application. This function must be called at least once or nothing will be displayed when
// the application starts.
//
// It also calls SetFocus() on the primitive.
func (a *Application) SetRoot(root Primitive) *Application {
	a.Lock()
	a.root = root
	if a.screen != nil {
		a.forceRedraw = true
	}
	a.Unlock()

	a.SetFocus(root)
	return a
}

// SetFocus sets the focus to a new primitive. All key events will be directed
// down the hierarchy (starting at the root) until a primitive handles them,
// which per default goes towards the focused primitive.
//
// Blur() will be called on the previously focused primitive. Focus() will be
// called on the new primitive.
func (a *Application) SetFocus(p Primitive) *Application {
	a.Lock()
	if a.focus != nil {
		a.focus.Blur()
	}
	a.focus = p
	if a.screen != nil {
		a.screen.HideCursor()
	}
	a.Unlock()
	if p != nil {
		p.Focus(func(p Primitive) {
			a.SetFocus(p)
		})
	}

	return a
}

// GetFocus returns the primitive which has the current focus. If none has it,
// nil is returned.
func (a *Application) GetFocus() Primitive {
	a.RLock()
	defer a.RUnlock()
	return a.focus
}

// QueueUpdate is used to synchronize access to primitives from non-main
// goroutines. The provided function will be executed as part of the event loop
// and thus will not cause race conditions with other such update functions or
// the Draw() function.
//
// Note that Draw() is not implicitly called after the execution of f as that
// may not be desirable. You can call Draw() from f if the screen should be
// refreshed after each update. Alternatively, use QueueUpdateDraw() to follow
// up with an immediate refresh of the screen.
//
// This function returns after f has executed.
func (a *Application) QueueUpdate(f func()) *Application {
	ch := make(chan struct{})
	a.updates <- queuedUpdate{f: f, done: ch}
	<-ch
	return a
}

// QueueUpdateDraw works like QueueUpdate() except it refreshes the screen
// immediately after executing f.
func (a *Application) QueueUpdateDraw(f func()) *Application {
	a.QueueUpdate(func() {
		f()
		a.draw()
	})
	return a
}

// QueueEvent sends an event to the Application event loop.
//
// It is not recommended for event to be nil.
func (a *Application) QueueEvent(event tcell.Event) *Application {
	a.RLock()
	events := a.events
	a.RUnlock()
	if events == nil {
		return a
	}
	events <- event
	return a
}

func (a *Application) executeCommand(command Command) bool {
	a.RLock()
	screen := a.screen
	a.RUnlock()

	switch command := command.(type) {
	case BatchCommand:
		handled := false
		for _, item := range command {
			if a.executeCommand(item) {
				handled = true
			}
		}
		return handled

	case RedrawCommand:
		return true

	case SetFocusCommand:
		if command.Target == nil {
			return false
		}
		a.RLock()
		changed := a.focus != command.Target
		a.RUnlock()
		a.SetFocus(command.Target)
		return changed

	case SetMouseCaptureCommand:
		a.Lock()
		a.mouseCapturingPrimitive = command.Target
		a.Unlock()

	case SetTitleCommand:
		if screen != nil {
			screen.SetTitle(string(command))
		}

	case SetClipboardCommand:
		if screen != nil && screen.HasClipboard() {
			screen.SetClipboard([]byte(string(command)))
			return true
		}
	case GetClipboardCommand:
		if screen != nil && screen.HasClipboard() {
			screen.GetClipboard()
			return true
		}

	case NotifyCommand:
		if screen != nil {
			screen.ShowNotification(command.Title, command.Body)
		}
	case EventCommand:
		go func() {
			if event := command(); event != nil {
				a.QueueEvent(event)
			}
		}()
		return false
	}

	return false
}
