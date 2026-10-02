package tview

import "github.com/ayn2op/tview/layout"

type mapped[T any] struct {
	child Element
	f     func(T) Msg
}

// Map turns the messages of type T that child produces into f's result, like Elm's Html.map. Anything else is returned as is, including a T that was passed to child, which is taken to have passed through unchanged.
func Map[T any](child Element, f func(T) Msg) Element {
	return mapped[T]{child: child, f: f}
}

// Size returns the size of the child.
func (m mapped[T]) Size() (width, height layout.Length) {
	return m.child.Size()
}

// Layout lays out the child.
func (m mapped[T]) Layout(limits layout.Limits) layout.Size {
	return m.child.Layout(limits)
}

// Draw draws the child.
func (m mapped[T]) Draw(screen Screen, area Rectangle) {
	m.child.Draw(screen, area)
}

// Handle passes msg to the child and maps what it produces.
func (m mapped[T]) Handle(msg Msg, area Rectangle) Msg {
	out := m.child.Handle(msg, area)
	if _, passed := msg.(T); passed {
		return out
	}
	if t, ok := out.(T); ok {
		return m.f(t)
	}
	return out
}
