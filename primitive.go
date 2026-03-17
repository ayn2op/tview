package tview

import "github.com/gdamore/tcell/v3"

// Primitive is the top-most interface for all graphical primitives.
type Primitive interface {
	// Draw draws this primitive onto the screen. Implementers can call the
	// screen's ShowCursor() function but should only do so when they have focus.
	// (They will need to keep track of this themselves.)
	Draw(screen tcell.Screen)
	// HandleEvent receives events when this primitive has focus.
	HandleEvent(event Event) Command

	// Rect returns the current position of the primitive, x, y, width, and
	// height.
	Rect() (int, int, int, int)
	// SetRect sets a new position of the primitive.
	SetRect(x, y, width, height int)

	// HasFocus determines if the primitive has focus. This function must return
	// true also if one of this primitive's child elements has focus.
	HasFocus() bool
	// Focus is called by the application when the primitive receives focus.
	// Implementers may call delegate() to pass the focus on to another primitive.
	Focus(delegate func(p Primitive))
	// Blur is called by the application when the primitive loses focus.
	Blur()
}
