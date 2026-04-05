package flex

import (
	"slices"

	"github.com/ayn2op/tview"
	"github.com/gdamore/tcell/v3"
)

// Direction controls the direction of items.
type Direction uint8

const (
	// One item per row.
	DirectionRow Direction = iota
	// One item per column.
	DirectionColumn
	// As defined in CSS, items distributed along a row.
	DirectionRowCSS = DirectionColumn
	// As defined in CSS, items distributed within a column.
	DirectionColumnCSS = DirectionRow
)

// item holds layout options for one item.
type item struct {
	Item       tview.Model // The item to be positioned. May be nil for an empty item.
	FixedSize  int         // The item's fixed size which may not be changed, 0 if it has no fixed size.
	Proportion int         // The item's proportion.
	Focus      bool        // Whether or not this item attracts the layout's focus.
}

// Model is a basic implementation of the Flexbox layout. The contained
// models are arranged horizontally or vertically. The way they are
// distributed along that dimension depends on their layout settings, which is
// either a fixed length or a proportional length. See AddItem() for details.
//
// See https://github.com/ayn2op/tview/wiki/Flex for an example.
type Model struct {
	*tview.Box

	// The items to be positioned.
	items []*item

	// Layout direction.
	direction Direction

	// If set to true, Model will use the entire screen as its available space
	// instead its box dimensions.
	fullScreen bool
}

// NewModel returns a new flexbox layout container with no models and its
// direction set to DirectionColumn. To add models to this layout, see AddItem().
// To change the direction, see SetDirection().
//
// Note that Box, the superclass of Model, will not clear its contents so that
// any nil flex items will leave their background unchanged. To clear a Model's
// background before any items are drawn, set it to a box with the desired
// color:
//
//	model.Box = tview.NewBox()
func NewModel() *Model {
	m := &Model{
		direction: DirectionColumn,
	}
	m.Box = tview.NewBox()
	m.SetDontClear(true)
	return m
}

// SetDirection sets the direction in which the contained models are
// distributed. This can be either DirectionColumn (default) or DirectionRow. Note that
// these are the opposite of what you would expect coming from CSS. You may also
// use DirectionColumnCSS or DirectionRowCSS, to remain in line with the CSS definition.
func (m *Model) SetDirection(direction Direction) *Model {
	if m.direction != direction {
		m.direction = direction
	}
	return m
}

// SetFullScreen sets the flag which, when true, causes the flex layout to use
// the entire screen space instead of whatever size it is currently assigned to.
func (m *Model) SetFullScreen(fullScreen bool) *Model {
	if m.fullScreen != fullScreen {
		m.fullScreen = fullScreen
	}
	return m
}

// AddItem adds a new item to the container. The "fixedSize" argument is a width
// or height that may not be changed by the layout algorithm. A value of 0 means
// that its size is flexible and may be changed. The "proportion" argument
// defines the relative size of the item compared to other flexible-size items.
// For example, items with a proportion of 2 will be twice as large as items
// with a proportion of 1. The proportion must be at least 1 if fixedSize == 0
// (ignored otherwise).
//
// If "focus" is set to true, the item will receive focus when the Model
// model receives focus. If multiple items have the "focus" flag set to
// true, the first one will receive focus.
//
// You can provide a nil value for the model. This will still consume screen
// space but nothing will be drawn.
func (m *Model) AddItem(p tview.Model, fixedSize, proportion int, focus bool) *Model {
	m.items = append(m.items, &item{Item: p, FixedSize: fixedSize, Proportion: proportion, Focus: focus})
	return m
}

// RemoveItem removes all items for the given model from the container,
// keeping the order of the remaining items intact.
func (m *Model) RemoveItem(item tview.Model) *Model {
	for index := len(m.items) - 1; index >= 0; index-- {
		if m.items[index].Item == item {
			m.items = slices.Delete(m.items, index, index+1)
		}
	}
	return m
}

// GetItemCount returns the number of items in this container.
func (m *Model) GetItemCount() int {
	return len(m.items)
}

// GetItem returns the model at the given index, starting with 0 for the
// first model in this container.
//
// This function will panic for out of range indices.
func (m *Model) GetItem(index int) tview.Model {
	return m.items[index].Item
}

// Clear removes all items from the container.
func (m *Model) Clear() *Model {
	if len(m.items) > 0 {
		m.items = nil
	}
	return m
}

// ResizeItem sets a new size for the item(s) with the given model. If there
// are multiple Model items with the same model, they will all receive the
// same size. For details regarding the size parameters, see AddItem().
func (m *Model) ResizeItem(p tview.Model, fixedSize, proportion int) *Model {
	for _, item := range m.items {
		if item.Item == p && (item.FixedSize != fixedSize || item.Proportion != proportion) {
			item.FixedSize = fixedSize
			item.Proportion = proportion
		}
	}
	return m
}

// Draw draws this model onto the screen.
func (m *Model) Draw(screen tcell.Screen) {
	m.DrawForSubclass(screen, m)

	// Calculate size and position of the items.

	// Do we use the entire screen?
	if m.fullScreen {
		width, height := screen.Size()
		m.SetRect(0, 0, width, height)
	}

	// How much space can we distribute?
	x, y, width, height := m.InnerRect()
	var proportionSum int
	distSize := width
	if m.direction == DirectionRow {
		distSize = height
	}
	for _, item := range m.items {
		if item.FixedSize > 0 {
			distSize -= item.FixedSize
		} else {
			proportionSum += item.Proportion
		}
	}

	// Calculate positions and draw items.
	pos := x
	if m.direction == DirectionRow {
		pos = y
	}
	for _, item := range m.items {
		size := item.FixedSize
		if size <= 0 {
			if proportionSum > 0 {
				size = distSize * item.Proportion / proportionSum
				distSize -= size
				proportionSum -= item.Proportion
			} else {
				size = 0
			}
		}
		if item.Item != nil {
			if m.direction == DirectionColumn {
				item.Item.SetRect(pos, y, size, height)
			} else {
				item.Item.SetRect(x, pos, width, size)
			}
		}
		pos += size

		if item.Item != nil {
			if item.Item.HasFocus() {
				defer item.Item.Draw(screen)
			} else {
				item.Item.Draw(screen)
			}
		}
	}
}

// Focus is called when this model receives focus.
func (m *Model) Focus(delegate func(m tview.Model)) {
	for _, item := range m.items {
		if item.Item != nil && item.Focus {
			delegate(item.Item)
			return
		}
	}
	m.Box.Focus(delegate)
}

// HasFocus returns whether or not this model has focus.
func (m *Model) HasFocus() bool {
	for _, item := range m.items {
		if item.Item != nil && item.Item.HasFocus() {
			return true
		}
	}
	return m.Box.HasFocus()
}

// Update handles input events for this model.
func (m *Model) Update(msg tview.Msg) tview.Cmd {
	switch msg := msg.(type) {
	case *tview.MouseMsg:
		if !m.InRect(msg.Position()) {
			return nil
		}

		// Pass mouse events along to the first child item that takes it.
		for _, item := range m.items {
			if item.Item == nil {
				continue
			}
			childCmds := item.Item.Update(msg)
			if childCmds != nil {
				return childCmds
			}
		}
		return nil
	}

	// Forward events to the focused child.
	for _, item := range m.items {
		if item.Item != nil && item.Item.HasFocus() {
			return item.Item.Update(msg)
		}
	}
	return nil
}
