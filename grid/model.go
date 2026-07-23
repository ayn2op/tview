package grid

import (
	"math"
	"slices"

	"github.com/ayn2op/tview"
	"github.com/gdamore/tcell/v3"
)

// item represents one model and its possible position on a grid.
type item struct {
	Item                        tview.Model // The item to be positioned. May be nil for an empty item.
	Row, Column                 int         // The top-left grid cell where the item is placed.
	Width, Height               int         // The number of rows and columns the item occupies.
	MinGridWidth, MinGridHeight int         // The minimum grid width/height for which this item is visible.
	Focus                       bool        // Whether or not this item attracts the layout's focus.

	visible    bool // Whether or not this item was visible the last time the grid was drawn.
	x, y, w, h int  // The last position of the item relative to the top-left corner of the grid. Undefined if visible is false.
}

// Model is an implementation of a grid-based layout. It works by defining the
// size of the rows and columns, then placing models into the grid.
//
// Some settings can lead to the grid exceeding its available space. SetOffset()
// can then be used to scroll in steps of rows and columns. These offset values
// can also be controlled with the arrow keys (or the "g","G", "j", "k", "h",
// and "l" keys) while the grid has focus and none of its contained models
// do.
type Model struct {
	*tview.Box

	// The items to be positioned.
	items []*item

	// The definition of the rows and columns of the grid. See
	// [Model.SetRows] / [Model.SetColumns] for details.
	rows, columns []int

	// The minimum sizes for rows and columns.
	minWidth, minHeight int

	// The size of the gaps between neighboring models. This is automatically
	// set to 1 if borders is true.
	gapRows, gapColumns int

	// The number of rows and columns skipped before drawing the top-left corner
	// of the grid.
	rowOffset, columnOffset int

	// Whether or not borders are drawn around grid items. If this is set to true,
	// a gap size of 1 is automatically assumed (which is filled with the border
	// graphics).
	borders bool

	// The color of the borders around grid items.
	bordersColor tcell.Color
}

// NewModel returns a new grid-based layout container with no initial models.
//
// Note that Box, the superclass of Model, will be transparent so that any grid
// areas not covered by any models will leave their background unchanged. To
// clear a Model's background before any items are drawn, reset its Box to one
// with the desired color:
//
//	grid.Box = tview.NewBox()
func NewModel() *Model {
	g := &Model{
		bordersColor: tview.Styles.GraphicsColor,
	}
	g.Box = tview.NewBox()
	g.SetDontClear(true)
	return g
}

// SetColumns defines how the columns of the grid are distributed. Each value
// defines the size of one column, starting with the leftmost column. Values
// greater than 0 represent absolute column widths (gaps not included). Values
// less than or equal to 0 represent proportional column widths or fractions of
// the remaining free space, where 0 is treated the same as -1. That is, a
// column with a value of -3 will have three times the width of a column with a
// value of -1 (or 0). The minimum width set with SetMinSize() is always
// observed.
//
// Models may extend beyond the columns defined explicitly with this
// function. A value of 0 is assumed for any undefined column. In fact, if you
// never call this function, all columns occupied by models will have the
// same width. On the other hand, unoccupied columns defined with this function
// will always take their place.
//
// Assuming a total width of the grid of 100 cells and a minimum width of 0, the
// following call will result in columns with widths of 30, 10, 15, 15, and 30
// cells:
//
//	grid.SetColumns(30, 10, -1, -1, -2)
//
// If a model were then placed in the 6th and 7th column, the resulting
// widths would be: 30, 10, 10, 10, 20, 10, and 10 cells.
//
// If you then called SetMinSize() as follows:
//
//	grid.SetMinSize(15, 20)
//
// The resulting widths would be: 30, 15, 15, 15, 20, 15, and 15 cells, a total
// of 125 cells, 25 cells wider than the available grid width.
func (m *Model) SetColumns(columns ...int) *Model {
	m.columns = columns
	return m
}

// SetRows defines how the rows of the grid are distributed. These values behave
// the same as the column values provided with [Model.SetColumns], see there
// for a definition and examples.
//
// The provided values correspond to row heights, the first value defining
// the height of the topmost row.
func (m *Model) SetRows(rows ...int) *Model {
	m.rows = rows
	return m
}

// SetSize is a shortcut for [Model.SetRows] and [Model.SetColumns] where
// all row and column values are set to the given size values. See
// [Model.SetColumns] for details on sizes.
func (m *Model) SetSize(numRows, numColumns, rowSize, columnSize int) *Model {
	rows := make([]int, numRows)
	for index := range rows {
		rows[index] = rowSize
	}
	columns := make([]int, numColumns)
	for index := range columns {
		columns[index] = columnSize
	}

	m.rows = rows
	m.columns = columns
	return m
}

// SetMinSize sets an absolute minimum width for rows and an absolute minimum
// height for columns. Panics if negative values are provided.
func (m *Model) SetMinSize(row, column int) *Model {
	if row < 0 || column < 0 {
		panic("Invalid minimum row/column size")
	}
	m.minHeight, m.minWidth = row, column
	return m
}

// SetGap sets the size of the gaps between neighboring models on the grid.
// If borders are drawn (see SetBorders()), these values are ignored and a gap
// of 1 is assumed. Panics if negative values are provided.
func (m *Model) SetGap(row, column int) *Model {
	if row < 0 || column < 0 {
		panic("Invalid gap size")
	}
	m.gapRows, m.gapColumns = row, column
	return m
}

// SetBorders sets whether or not borders are drawn around grid items. Setting
// this value to true will cause the gap values (see SetGap()) to be ignored and
// automatically assumed to be 1 where the border graphics are drawn.
func (m *Model) SetBorders(borders bool) *Model {
	m.borders = borders
	return m
}

// SetBordersColor sets the color of the item borders.
func (m *Model) SetBordersColor(color tcell.Color) *Model {
	m.bordersColor = color
	return m
}

// AddItem adds a model and its position to the grid. The top-left corner
// of the model will be located in the top-left corner of the grid cell at
// the given row and column and will span "rowSpan" rows and "colSpan" columns.
// For example, for a model to occupy rows 2, 3, and 4 and columns 5 and 6:
//
//	grid.AddItem(p, 2, 5, 3, 2, 0, 0, true)
//
// If rowSpan or colSpan is 0, the model will not be drawn.
//
// You can add the same model multiple times with different grid positions.
// The minGridWidth and minGridHeight values will then determine which of those
// positions will be used. This is similar to CSS media queries. These minimum
// values refer to the overall size of the grid. If multiple items for the same
// model apply, the one with the highest minimum value (width or height,
// whatever is higher) will be used, or the model added last if those values
// are the same. Example:
//
//	grid.AddItem(p, 0, 0, 0, 0, 0, 0, true). // Hide in small grids.
//	  AddItem(p, 0, 0, 1, 2, 100, 0, true).  // One-column layout for medium grids.
//	  AddItem(p, 1, 1, 3, 2, 300, 0, true)   // Multi-column layout for large grids.
//
// To use the same grid layout for all sizes, simply set minGridWidth and
// minGridHeight to 0.
//
// If the item's focus is set to true, it will receive focus when the grid
// receives focus. If there are multiple items with a true focus flag, the last
// visible one that was added will receive focus.
func (m *Model) AddItem(p tview.Model, row, column, rowSpan, colSpan, minGridHeight, minGridWidth int, focus bool) *Model {
	m.items = append(m.items, &item{
		Item:          p,
		Row:           row,
		Column:        column,
		Height:        rowSpan,
		Width:         colSpan,
		MinGridHeight: minGridHeight,
		MinGridWidth:  minGridWidth,
		Focus:         focus,
	})
	return m
}

// RemoveItem removes all items for the given model from the grid, keeping
// the order of the remaining items intact.
func (g *Model) RemoveItem(m tview.Model) *Model {
	for index := len(g.items) - 1; index >= 0; index-- {
		if g.items[index].Item == m {
			g.items = slices.Delete(g.items, index, index+1)
		}
	}
	return g
}

// Clear removes all items from the grid.
func (m *Model) Clear() *Model {
	m.items = nil
	return m
}

// Offset returns the current row and column offset (see SetOffset() for
// details).
func (m *Model) Offset() (rows, columns int) {
	return m.rowOffset, m.columnOffset
}

// SetOffset sets the number of rows and columns which are skipped before
// drawing the first grid cell in the top-left corner. As the grid will never
// completely move off the screen, these values may be adjusted the next time
// the grid is drawn. The actual position of the grid may also be adjusted such
// that contained models that have focus remain visible.
func (m *Model) SetOffset(rows, columns int) *Model {
	m.rowOffset, m.columnOffset = rows, columns
	return m
}

// Focus is called when this model receives focus.
func (m *Model) Focus(delegate func(m tview.Model)) {
	for _, item := range m.items {
		if item.Focus {
			delegate(item.Item)
			return
		}
	}
	m.Box.Focus(delegate)
}

// HasFocus returns whether or not this model has focus.
func (m *Model) HasFocus() bool {
	for _, item := range m.items {
		if item.visible && item.Item.HasFocus() {
			return true
		}
	}
	return m.Box.HasFocus()
}

// View draws this model onto the screen.
func (m *Model) View(screen tcell.Screen) {
	m.Box.View(screen)

	x, y, width, height := m.InnerRect()
	screenWidth, screenHeight := screen.Size()

	// Make a list of items which apply.
	items := make([]*item, 0, len(m.items))
ItemLoop:
	for _, item := range m.items {
		item.visible = false
		if item.Item == nil || item.Width <= 0 || item.Height <= 0 || width < item.MinGridWidth || height < item.MinGridHeight {
			continue // Disqualified.
		}

		// Check for overlaps and multiple layouts of the same item.
		for index, existing := range items {
			// Do they overlap or are identical?
			if item.Item != existing.Item &&
				(item.Row >= existing.Row+existing.Height || item.Row+item.Height <= existing.Row ||
					item.Column >= existing.Column+existing.Width || item.Column+item.Width <= existing.Column) {
				continue // They don't and aren't.
			}

			// What's their minimum size?
			itemMin := max(item.MinGridHeight, item.MinGridWidth)
			existingMin := max(existing.MinGridHeight, existing.MinGridWidth)

			// Which one is more important?
			if itemMin < existingMin {
				continue ItemLoop // This one isn't. Drop it.
			}
			items[index] = item // This one is. Replace the other.
			continue ItemLoop
		}

		// This item will be visible.
		items = append(items, item)
	}

	// How many rows and columns do we have?
	rows := len(m.rows)
	columns := len(m.columns)
	for _, item := range items {
		rowEnd := item.Row + item.Height
		if rowEnd > rows {
			rows = rowEnd
		}
		columnEnd := item.Column + item.Width
		if columnEnd > columns {
			columns = columnEnd
		}
	}
	if rows == 0 || columns == 0 {
		return // No content.
	}
	if width <= 0 || height <= 0 {
		return
	}

	// Where are they located?
	rowPos := make([]int, rows)
	rowHeight := make([]int, rows)
	columnPos := make([]int, columns)
	columnWidth := make([]int, columns)

	// How much space do we distribute?
	remainingWidth := width
	remainingHeight := height
	proportionalWidth := 0
	proportionalHeight := 0
	for index, row := range m.rows {
		if row > 0 {
			if row < m.minHeight {
				row = m.minHeight
			}
			remainingHeight -= row
			rowHeight[index] = row
		} else if row == 0 {
			proportionalHeight++
		} else {
			proportionalHeight += -row
		}
	}
	for index, column := range m.columns {
		if column > 0 {
			if column < m.minWidth {
				column = m.minWidth
			}
			remainingWidth -= column
			columnWidth[index] = column
		} else if column == 0 {
			proportionalWidth++
		} else {
			proportionalWidth += -column
		}
	}
	if m.borders {
		remainingHeight -= rows + 1
		remainingWidth -= columns + 1
	} else {
		remainingHeight -= (rows - 1) * m.gapRows
		remainingWidth -= (columns - 1) * m.gapColumns
	}
	if rows > len(m.rows) {
		proportionalHeight += rows - len(m.rows)
	}
	if columns > len(m.columns) {
		proportionalWidth += columns - len(m.columns)
	}
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	if remainingHeight < 0 {
		remainingHeight = 0
	}

	// Distribute proportional rows/columns.
	for index := range rows {
		row := 0
		if index < len(m.rows) {
			row = m.rows[index]
		}
		if row > 0 {
			continue // Not proportional. We already know the width.
		} else if row == 0 {
			row = 1
		} else {
			row = -row
		}
		rowAbs := row * remainingHeight / proportionalHeight
		remainingHeight -= rowAbs
		proportionalHeight -= row
		if rowAbs < m.minHeight {
			rowAbs = m.minHeight
		}
		rowHeight[index] = rowAbs
	}
	for index := range columns {
		column := 0
		if index < len(m.columns) {
			column = m.columns[index]
		}
		if column > 0 {
			continue // Not proportional. We already know the height.
		} else if column == 0 {
			column = 1
		} else {
			column = -column
		}
		columnAbs := column * remainingWidth / proportionalWidth
		remainingWidth -= columnAbs
		proportionalWidth -= column
		if columnAbs < m.minWidth {
			columnAbs = m.minWidth
		}
		columnWidth[index] = columnAbs
	}

	// Calculate row/column positions.
	var columnX, rowY int
	if m.borders {
		columnX++
		rowY++
	}
	for index, row := range rowHeight {
		rowPos[index] = rowY
		gap := m.gapRows
		if m.borders {
			gap = 1
		}
		rowY += row + gap
	}
	for index, column := range columnWidth {
		columnPos[index] = columnX
		gap := m.gapColumns
		if m.borders {
			gap = 1
		}
		columnX += column + gap
	}

	// Calculate model positions.
	var focus *item // The item which has focus.
	for _, item := range items {
		px := columnPos[item.Column]
		py := rowPos[item.Row]
		var pw, ph int
		for index := range item.Height {
			ph += rowHeight[item.Row+index]
		}
		for index := range item.Width {
			pw += columnWidth[item.Column+index]
		}
		if m.borders {
			pw += item.Width - 1
			ph += item.Height - 1
		} else {
			pw += (item.Width - 1) * m.gapColumns
			ph += (item.Height - 1) * m.gapRows
		}
		item.x, item.y, item.w, item.h = px, py, pw, ph
		item.visible = true
		if item.Item.HasFocus() {
			focus = item
		}
	}

	// Calculate screen offsets.
	var offsetX, offsetY int
	add := 1
	if !m.borders {
		add = m.gapRows
	}
	for index, height := range rowHeight {
		if index >= m.rowOffset {
			break
		}
		offsetY += height + add
	}
	if !m.borders {
		add = m.gapColumns
	}
	for index, width := range columnWidth {
		if index >= m.columnOffset {
			break
		}
		offsetX += width + add
	}

	// The focused item must be within the visible area.
	if focus != nil {
		if focus.y+focus.h-offsetY >= height {
			offsetY = focus.y - height + focus.h
		}
		if focus.y-offsetY < 0 {
			offsetY = focus.y
		}
		if focus.x+focus.w-offsetX >= width {
			offsetX = focus.x - width + focus.w
		}
		if focus.x-offsetX < 0 {
			offsetX = focus.x
		}
	}

	// Adjust row/column offsets based on this value.
	var from, to int
	for index, pos := range rowPos {
		if pos-offsetY < 0 {
			from = index + 1
		}
		if pos-offsetY < height {
			to = index
		}
	}
	if m.rowOffset < from {
		m.rowOffset = from
	}
	if m.rowOffset > to {
		m.rowOffset = to
	}
	from, to = 0, 0
	for index, pos := range columnPos {
		if pos-offsetX < 0 {
			from = index + 1
		}
		if pos-offsetX < width {
			to = index
		}
	}
	if m.columnOffset < from {
		m.columnOffset = from
	}
	if m.columnOffset > to {
		m.columnOffset = to
	}

	// Draw models and borders.
	borderStyle := tcell.StyleDefault.Background(m.BackgroundColor()).Foreground(m.bordersColor)
	for _, item := range items {
		// Final model position.
		if !item.visible {
			continue
		}
		item.x -= offsetX
		item.y -= offsetY
		if item.x >= width || item.x+item.w <= 0 || item.y >= height || item.y+item.h <= 0 {
			item.visible = false
			continue
		}
		if item.x+item.w > width {
			item.w = width - item.x
		}
		if item.y+item.h > height {
			item.h = height - item.y
		}
		if item.x < 0 {
			item.w += item.x
			item.x = 0
		}
		if item.y < 0 {
			item.h += item.y
			item.y = 0
		}
		if item.w <= 0 || item.h <= 0 {
			item.visible = false
			continue
		}
		item.x += x
		item.y += y
		item.Item.SetRect(item.x, item.y, item.w, item.h)

		// Draw model.
		if item == focus {
			defer item.Item.View(screen)
		} else {
			item.Item.View(screen)
		}

		// Draw border around model.
		if m.borders {
			borderSet := m.BorderSet()
			for bx := item.x; bx < item.x+item.w; bx++ { // Top/bottom lines.
				if bx < 0 || bx >= screenWidth {
					continue
				}
				by := item.y - 1
				if by >= 0 && by < screenHeight {
					tview.PrintJoinedSemigraphics(screen, bx, by, borderSet.Top, borderStyle)
				}
				by = item.y + item.h
				if by >= 0 && by < screenHeight {
					tview.PrintJoinedSemigraphics(screen, bx, by, borderSet.Bottom, borderStyle)
				}
			}
			for by := item.y; by < item.y+item.h; by++ { // Left/right lines.
				if by < 0 || by >= screenHeight {
					continue
				}
				bx := item.x - 1
				if bx >= 0 && bx < screenWidth {
					tview.PrintJoinedSemigraphics(screen, bx, by, borderSet.Left, borderStyle)
				}
				bx = item.x + item.w
				if bx >= 0 && bx < screenWidth {
					tview.PrintJoinedSemigraphics(screen, bx, by, borderSet.Right, borderStyle)
				}
			}
			bx, by := item.x-1, item.y-1 // Top-left corner.
			if bx >= 0 && bx < screenWidth && by >= 0 && by < screenHeight {
				tview.PrintJoinedSemigraphics(screen, bx, by, borderSet.TopLeft, borderStyle)
			}
			bx, by = item.x+item.w, item.y-1 // Top-right corner.
			if bx >= 0 && bx < screenWidth && by >= 0 && by < screenHeight {
				tview.PrintJoinedSemigraphics(screen, bx, by, borderSet.TopRight, borderStyle)
			}
			bx, by = item.x-1, item.y+item.h // Bottom-left corner.
			if bx >= 0 && bx < screenWidth && by >= 0 && by < screenHeight {
				tview.PrintJoinedSemigraphics(screen, bx, by, borderSet.BottomLeft, borderStyle)
			}
			bx, by = item.x+item.w, item.y+item.h // Bottom-right corner.
			if bx >= 0 && bx < screenWidth && by >= 0 && by < screenHeight {
				tview.PrintJoinedSemigraphics(screen, bx, by, borderSet.BottomRight, borderStyle)
			}
		}
	}
}

// Update handles input events for this model.
func (m *Model) Update(msg tview.Msg) tview.Cmd {
	switch msg := msg.(type) {
	case tview.MouseMsg:
		x, y := msg.Position()
		if !m.InRect(x, y) {
			return nil
		}

		// Pass mouse events along to the first child item that takes it.
		for _, item := range m.items {
			if item.Item == nil || !item.visible {
				continue
			}
			if tview.ModelInRect(item.Item, x, y) {
				return item.Item.Update(msg)
			}
		}
	case tview.KeyMsg:
		previousRowOffset, previousColumnOffset := m.rowOffset, m.columnOffset
		if !m.Box.HasFocus() {
			// Pass event on to child model.
			for _, item := range m.items {
				if item != nil && item.Item.HasFocus() {
					return item.Item.Update(msg)
				}
			}
			return nil
		}

		// Process our own key events if we have direct focus.
		switch msg.Key() {
		case tcell.KeyRune:
			switch msg.Str() {
			case "g":
				m.rowOffset, m.columnOffset = 0, 0
			case "G":
				m.rowOffset = math.MaxInt32
			case "j":
				m.rowOffset++
			case "k":
				m.rowOffset--
			case "h":
				m.columnOffset--
			case "l":
				m.columnOffset++
			}
		case tcell.KeyHome:
			m.rowOffset, m.columnOffset = 0, 0
		case tcell.KeyEnd:
			m.rowOffset = math.MaxInt32
		case tcell.KeyUp:
			m.rowOffset--
		case tcell.KeyDown:
			m.rowOffset++
		case tcell.KeyLeft:
			m.columnOffset--
		case tcell.KeyRight:
			m.columnOffset++
		}
		if m.rowOffset != previousRowOffset || m.columnOffset != previousColumnOffset {
			return nil
		}
	}

	// Forward events to the focused child.
	for _, item := range m.items {
		if item != nil && item.Item.HasFocus() {
			return item.Item.Update(msg)
		}
	}
	return nil
}
