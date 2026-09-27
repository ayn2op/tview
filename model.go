package tview

import "github.com/gdamore/tcell/v3"

type Screen = tcell.Screen

// Model is the state of an application or a part of it, how messages change it, and how it is shown.
type Model interface {
	// Init returns a command to run when the model starts, or nil.
	Init() Cmd
	// Update changes the model in response to a message and returns a command to run, or nil.
	Update(Msg) Cmd
	// View returns the element that draws this model.
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
	// Handle translates an input message received within the given area into the message passed to Update. It returns nil to drop the message.
	Handle(Msg, Rectangle) Msg
}
