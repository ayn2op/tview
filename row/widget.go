// Package row lays out widgets left to right.
package row

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/flex"
)

// Widget lays out its children left to right.
type Widget = flex.Widget

// New returns a row of children, skipping nil ones. It fills its parent in both directions by default.
func New(children ...tview.Widget) Widget {
	return flex.New(true, children...)
}
