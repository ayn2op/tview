// Package column lays out elements top to bottom.
package column

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/flex"
)

// Widget lays out its children top to bottom.
type Widget = flex.Widget

// New returns a column of children, skipping nil ones. It fills its parent in both directions by default.
func New(children ...tview.Element) Widget {
	return flex.New(false, children...)
}
