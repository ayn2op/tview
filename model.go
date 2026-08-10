package tview

import "github.com/gdamore/tcell/v3"

// Model is the top-most interface for all graphical models.
type Model interface {
	// Update receives messages when this model has focus.
	Update(Msg) Cmd
	// View draws this model onto the screen.
	View(tcell.Screen)

	// Rect returns the current position of the model, x, y, width, and
	// height.
	Rect() (int, int, int, int)
	// SetRect sets a new position of the model.
	SetRect(x, y, width, height int)

	// HasFocus determines if the model has focus. This function must return
	// true also if one of this model's child elements has focus.
	HasFocus() bool
}
