package tview

import "github.com/gdamore/tcell/v3"

// Model is the top-most interface for all graphical models.
type Model interface {
	// Draw draws this model onto the screen. Implementers can call the
	// screen's ShowCursor() function but should only do so when they have focus.
	// (They will need to keep track of this themselves.)
	Draw(screen tcell.Screen)
	// HandleEvent receives events when this model has focus.
	HandleEvent(event Event) Cmd

	// Rect returns the current position of the model, x, y, width, and
	// height.
	Rect() (int, int, int, int)
	// SetRect sets a new position of the model.
	SetRect(x, y, width, height int)

	// HasFocus determines if the model has focus. This function must return
	// true also if one of this model's child elements has focus.
	HasFocus() bool
	// Focus is called by the application when the model receives focus.
	// Implementers may call delegate() to pass the focus on to another model.
	Focus(delegate func(Model))
	// Blur is called by the application when the model loses focus.
	Blur()
}
