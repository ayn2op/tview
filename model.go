package tview

import "github.com/gdamore/tcell/v3"

type Screen = tcell.Screen

// Model is the state of an application or a part of it, how messages change it, and how it is shown. M is the model's own type, so that Update returns it without a type assertion.
//
// Implement Model with value receivers and treat the model as immutable: only Update produces a changed model, by returning it. Init and View work on a copy, so changes they make are lost.
type Model[M any] interface {
	// Init returns a command to run when the model starts, or nil.
	Init() Cmd
	// Update returns the model changed in response to a message and a command to run, or nil.
	Update(Msg) (M, Cmd)
	// View returns the element that draws the model.
	View() Element
}

// Rectangle is a region of the screen.
type Rectangle struct {
	X, Y, Width, Height int
}

// Contains reports whether the point x, y lies within r.
func (r Rectangle) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.Width && y >= r.Y && y < r.Y+r.Height
}

// Element is a drawable part of the user interface.
type Element interface {
	// Draw draws the element onto the screen within the given area.
	Draw(Screen, Rectangle)
	// Handle translates a message, such as input received within the given area, into the message passed to Update. It returns nil to drop the message.
	Handle(Msg, Rectangle) Msg
}
