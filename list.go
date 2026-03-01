package tview

import (
	"github.com/gdamore/tcell/v3"
	"github.com/rivo/uniseg"
)

// ListItem represents a primitive which can be measured for a given width.
//
// Scroll list items are responsible for reporting their own height so the list can
// layout and scroll variable-height items.
type ListItem interface {
	Primitive
	Height(width int) int
}

// ListBuilder returns a list item for the given index and cursor position.
// It must return nil when the index is out of range.
type ListBuilder func(index int, cursor int) ListItem

// List displays a virtual list of primitives returned by a builder function.
type List struct {
	*Box

	Builder      ListBuilder
	gap          int
	snapToItems  bool
	centerCursor bool
	trackEnd     bool
	atEnd        bool

	cursor int
	scroll listState

	changed func(index int)

	lastDraw []listDrawnItem
	lastRect listRect

	scrollBarVisibility  ScrollBarVisibility
	scrollBar            *ScrollBar
	scrollBarInteraction scrollBarInteractionState
}

// ScrollBarVisibility controls when List renders its vertical scrollBar.
type ScrollBarVisibility uint8

const (
	ScrollBarVisibilityAutomatic ScrollBarVisibility = iota
	ScrollBarVisibilityAlways
	ScrollBarVisibilityNever
)

type listState struct {
	// Index of the top item in the viewport.
	top int
	// Line offset into the top item; negative values mean the item is scrolled up.
	offset int
	// Pending scroll delta in lines to apply on the next draw.
	pending int
	// Ensure the cursor is visible on the next draw.
	wantsCursor bool
}

type listDrawnItem struct {
	index  int
	item   ListItem
	row    int
	height int
}

type listRect struct {
	x      int
	y      int
	width  int
	height int
}

type listScrollBarState struct {
	contentWidth   int
	viewportHeight int
	position       int
	contentLength  int
	viewportLength int
	metrics        scrollMetrics
}

type scrollBarInteractionState struct {
	dragDelta int
	dragMoved bool
	state     listScrollBarState
}

const (
	listScrollBarNoDrag = -1
)

// NewList returns a new scroll list.
func NewList() *List {
	return &List{
		Box:                 NewBox(),
		centerCursor:        true,
		cursor:              -1,
		scrollBarVisibility: ScrollBarVisibilityAutomatic,
		scrollBar:           NewScrollBar(),
		scrollBarInteraction: scrollBarInteractionState{
			dragDelta: listScrollBarNoDrag,
		},
	}
}

// SetScrollBarVisibility sets when the list scrollBar is rendered.
func (l *List) SetScrollBarVisibility(visibility ScrollBarVisibility) *List {
	if l.scrollBarVisibility != visibility {
		l.scrollBarVisibility = visibility
	}
	return l
}

// SetScrollBar sets the ScrollBar primitive used by this list.
func (l *List) SetScrollBar(scrollBar *ScrollBar) *List {
	if l.scrollBar != scrollBar {
		l.scrollBar = scrollBar
	}
	return l
}

// SetBuilder sets the builder used to create list items on demand.
func (l *List) SetBuilder(builder ListBuilder) *List {
	if l.Builder != nil || builder != nil {
		l.Builder = builder
	}
	return l
}

// Clear removes all items from the list by clearing the builder and resetting
// scroll state.
func (l *List) Clear() *List {
	l.Builder = nil
	l.cursor = -1
	l.scroll = listState{}
	l.setLastDraw(nil)
	l.lastRect = listRect{}
	l.atEnd = false
	return l
}

// SetGap sets the number of blank rows between items.
func (l *List) SetGap(gap int) *List {
	if gap < 0 {
		gap = 0
	}
	if l.gap != gap {
		l.gap = gap
	}
	return l
}

// SetSnapToItems toggles snapping so only fully visible items are shown.
func (l *List) SetSnapToItems(snap bool) *List {
	if l.snapToItems != snap {
		l.snapToItems = snap
	}
	return l
}

// SetCenterCursor controls whether the cursor is kept centered whenever
// possible.
func (l *List) SetCenterCursor(center bool) *List {
	if l.centerCursor != center {
		l.centerCursor = center
	}
	return l
}

// SetTrackEnd toggles auto-scrolling when the view is already at the end.
func (l *List) SetTrackEnd(track bool) *List {
	if l.trackEnd != track {
		l.trackEnd = track
	}
	return l
}

// ScrollToStart resets the scroll position to the top (index 0), without
// changing the cursor.
func (l *List) ScrollToStart() *List {
	if l.scroll.top != 0 || l.scroll.offset != 0 || l.scroll.wantsCursor || l.atEnd {
		l.scroll.top = 0
		l.scroll.offset = 0
		l.scroll.wantsCursor = false
		l.atEnd = false
	}
	return l
}

// ScrollToEnd scrolls the view so the last items are visible.
func (l *List) ScrollToEnd() *List {
	_, _, width, height := l.GetInnerRect()
	if width <= 0 || height <= 0 {
		return l
	}
	top, offset := l.endScrollState(width, height)
	if l.scroll.top != top || l.scroll.offset != offset || l.scroll.wantsCursor || !l.atEnd {
		l.scroll.top, l.scroll.offset = top, offset
		l.scroll.wantsCursor = false
		l.atEnd = true
	}
	return l
}

// SetCursor sets the currently selected item index.
func (l *List) SetCursor(index int) *List {
	if index < -1 {
		index = -1
	}
	if l.cursor != index {
		l.cursor = index
		l.atEnd = false
		l.ensureScroll()
		if l.changed != nil {
			l.changed(l.cursor)
		}
	}
	return l
}

// Cursor returns the current cursor index.
func (l *List) Cursor() int {
	return l.cursor
}

// SetPendingScroll sets a pending scroll amount, in lines. Positive numbers
// scroll down.
func (l *List) SetPendingScroll(lines int) *List {
	if l.scroll.pending != lines {
		l.scroll.pending = lines
	}
	return l
}

// ScrollUp scrolls the list up by one line.
func (l *List) ScrollUp() *List {
	l.scroll.pending -= 1
	return l
}

// ScrollDown scrolls the list down by one line.
func (l *List) ScrollDown() *List {
	l.scroll.pending += 1
	return l
}

// NextItem moves the cursor to the next item, if any.
func (l *List) NextItem() bool {
	if l.Builder == nil {
		return false
	}
	if l.cursor < 0 {
		if l.Builder(0, l.cursor) == nil {
			return false
		}
		l.cursor = 0
		l.ensureScroll()
		if l.changed != nil {
			l.changed(l.cursor)
		}
		return true
	}
	if l.Builder(l.cursor+1, l.cursor) == nil {
		return false
	}
	l.cursor++
	l.ensureScroll()
	if l.changed != nil {
		l.changed(l.cursor)
	}
	return true
}

// PrevItem moves the cursor to the previous item, if any.
func (l *List) PrevItem() bool {
	if l.cursor <= 0 {
		return false
	}
	if l.Builder == nil {
		return false
	}
	if l.Builder(l.cursor-1, l.cursor) == nil {
		return false
	}
	l.cursor--
	l.ensureScroll()
	if l.changed != nil {
		l.changed(l.cursor)
	}
	return true
}

// SetChangedFunc sets a handler that is called when the cursor changes.
func (l *List) SetChangedFunc(handler func(index int)) *List {
	l.changed = handler
	return l
}

func (l *List) setLastDraw(children []listDrawnItem) {
	l.lastDraw = children
}

// Draw draws this primitive onto the screen.
func (l *List) Draw(screen tcell.Screen) {
	l.DrawForSubclass(screen, l)
	l.scrollBarInteraction.state = listScrollBarState{}

	x, y, width, height := l.GetInnerRect()
	if width <= 0 || height <= 0 || l.Builder == nil {
		return
	}

	usableWidth := width
	scrollBarX := x + width - 1
	drawScrollBar := false
	if width > 1 {
		switch l.scrollBarVisibility {
		case ScrollBarVisibilityAlways:
			drawScrollBar = true
		case ScrollBarVisibilityAutomatic:
			drawScrollBar = l.totalContentHeight(width) > height
		case ScrollBarVisibilityNever:
			drawScrollBar = false
		}
		if drawScrollBar {
			usableWidth, scrollBarX = l.scrollBarLayout(x, width)
		}
	}
	if usableWidth <= 0 {
		return
	}

	// If we were already at the end, keep following new items without
	// forcing full scans during normal scrolling.
	if l.trackEnd && l.atEnd {
		l.scroll.top, l.scroll.offset = l.endScrollState(usableWidth, height)
		l.scroll.wantsCursor = false
	}

	// In snap mode, ensure the cursor item is within the fully visible window.
	if l.snapToItems && l.scroll.wantsCursor && l.cursor >= 0 {
		visible := l.visibleItemCount(usableWidth, height)
		if l.cursor < l.scroll.top || l.cursor >= l.scroll.top+visible {
			l.scroll.top = l.cursor
			l.scroll.offset = 0
		}
		l.scroll.wantsCursor = false
	}

	// In non-snap mode, try to center the cursor when there is room.
	if !l.snapToItems && l.centerCursor && l.scroll.wantsCursor && l.cursor >= 0 {
		if top, offset, centered := l.centerScrollState(usableWidth, height); centered {
			l.scroll.top = top
			l.scroll.offset = offset
			l.scroll.wantsCursor = false
		}
	}

	pendingDelta := l.scroll.pending
	ah := -(l.scroll.offset + pendingDelta)
	l.scroll.pending = 0

	if ah > 0 && l.scroll.top == 0 {
		ah = 0
		l.scroll.offset = 0
	}

rebuild:
	// Rebuild the viewport whenever we change top/offset to keep the cursor in view.
	children := make([]listDrawnItem, 0, 16)
	startIndex := l.scroll.top

	if ah > 0 {
		// We scrolled upward into the previous top item; prepend enough items above.
		l.insertChildren(&children, usableWidth, ah)
		if len(children) > 0 {
			last := children[len(children)-1]
			ah = last.row + last.height + l.gap
		}
	}

	endReached := false
	for i := startIndex; ; i++ {
		item := l.Builder(i, l.cursor)
		if item == nil {
			endReached = true
			break
		}

		itemHeight := l.itemHeight(item, usableWidth)
		children = append(children, listDrawnItem{
			index:  i,
			item:   item,
			row:    ah,
			height: itemHeight,
		})
		ah += itemHeight + l.gap

		if l.scroll.wantsCursor && i <= l.cursor {
			continue
		}
		if ah >= height {
			break
		}
	}

	if len(children) == 0 {
		l.scroll.top = 0
		l.scroll.offset = 0
		l.setLastDraw(nil)
		l.lastRect = listRect{x: x, y: y, width: width, height: height}
		l.atEnd = false
		return
	}

	// If the cursor item didn't make it into the built slice, restart from it.
	if l.snapToItems && l.scroll.wantsCursor && l.cursor >= 0 {
		found := false
		for _, child := range children {
			if child.index == l.cursor {
				found = true
				break
			}
		}
		if !found {
			l.scroll.top = l.cursor
			l.scroll.offset = 0
			l.scroll.wantsCursor = false
			goto rebuild
		}
	}

	if l.snapToItems {
		// Drop partial items so only fully visible ones remain.
		children = l.trimToFullItems(children, height)
		if len(children) == 0 {
			l.scroll.top = 0
			l.scroll.offset = 0
			l.setLastDraw(nil)
			l.lastRect = listRect{x: x, y: y, width: width, height: height}
			l.atEnd = false
			return
		}

		// Fill remaining space with fully visible items if possible.
		nextIndex := children[len(children)-1].index + 1
		currentBottom := children[len(children)-1].row + children[len(children)-1].height
		for {
			item := l.Builder(nextIndex, l.cursor)
			if item == nil {
				break
			}
			itemHeight := l.itemHeight(item, usableWidth)
			nextRow := currentBottom + l.gap
			if nextRow+itemHeight > height {
				break
			}
			children = append(children, listDrawnItem{
				index:  nextIndex,
				item:   item,
				row:    nextRow,
				height: itemHeight,
			})
			currentBottom = nextRow + itemHeight
			nextIndex++
		}
	}

	// When scrolling down at the end, clamp so the last item aligns to the bottom.
	if endReached && pendingDelta > 0 {
		last := children[len(children)-1]
		bottom := last.row + last.height
		if children[0].row < 0 && bottom < height {
			adj := height - bottom
			for i := range children {
				children[i].row += adj
			}
		}
	}

	// Non-snap mode: adjust rows so the cursor item is fully visible.
	if l.scroll.wantsCursor {
		for _, child := range children {
			if child.index != l.cursor {
				continue
			}
			bottom := child.row + child.height
			if bottom > height {
				adj := height - bottom
				for i := range children {
					children[i].row += adj
				}
			}
			l.scroll.wantsCursor = false
			break
		}
	}

	if l.snapToItems {
		// Snap mode uses the first item as the top anchor.
		l.scroll.top = children[0].index
		l.scroll.offset = 0
	} else {
		// Non-snap mode keeps the first partially visible item as the top anchor.
		for i := range children {
			child := children[i]
			span := child.height
			if l.gap > 0 {
				span += l.gap
			}
			if child.row <= 0 && child.row+span > 0 {
				l.scroll.top = child.index
				l.scroll.offset = -child.row
				break
			}
		}
	}

	last := children[len(children)-1]
	if !endReached && l.Builder(last.index+1, l.cursor) == nil {
		endReached = true
	}
	l.atEnd = endReached && last.row+last.height <= height

	l.setLastDraw(children)
	l.lastRect = listRect{x: x, y: y, width: width, height: height}

	clipped := newClippedScreen(screen, x, y, width, height)
	for _, child := range children {
		child.item.SetRect(x, y+child.row, usableWidth, child.height)
		child.item.Draw(clipped)
	}

	if drawScrollBar {
		if l.scrollBar == nil {
			l.scrollBar = NewScrollBar().
				SetArrows(ScrollBarArrowsNone)
		}
		scrollBarState, ok := l.computeScrollBarState(usableWidth, height, children)
		if !ok {
			return
		}
		l.scrollBarInteraction.state = scrollBarState
		l.scrollBar.SetRect(scrollBarX, y, 1, height)
		l.scrollBar.SetLengths(ScrollLengths{
			ContentLen:  scrollBarState.contentLength,
			ViewportLen: scrollBarState.viewportLength,
		})
		l.scrollBar.SetOffset(scrollBarState.position)
		l.scrollBar.Draw(screen)
	}
}

func (l *List) itemHeight(item ListItem, width int) int {
	if item == nil {
		return 0
	}
	height := max(item.Height(width), 1)
	return height
}

func (l *List) totalContentHeight(width int) int {
	if l.Builder == nil || width <= 0 {
		return 0
	}
	total := 0
	for i := 0; ; i++ {
		item := l.Builder(i, l.cursor)
		if item == nil {
			break
		}
		if i > 0 {
			total += l.gap
		}
		total += l.itemHeight(item, width)
	}
	return total
}

func (l *List) scrollBarMetrics(width int, viewport int, children []listDrawnItem) (position int, contentLength int, viewportContentLength int) {
	content := l.totalContentHeight(width)
	if len(children) == 0 || content <= 0 || viewport <= 0 {
		return 0, 0, max(viewport, 0)
	}

	first := children[0]
	for i := 0; i < first.index; i++ {
		item := l.Builder(i, l.cursor)
		if item == nil {
			break
		}
		if i > 0 {
			position += l.gap
		}
		position += l.itemHeight(item, width)
	}

	position -= first.row
	if position < 0 {
		position = 0
	}

	maxOffset := max(content-viewport, 0)
	if position > maxOffset {
		position = maxOffset
	}

	contentLength = content
	viewportContentLength = viewport
	return position, contentLength, viewportContentLength
}

func (l *List) insertChildren(children *[]listDrawnItem, width int, ah int) {
	if l.scroll.top <= 0 {
		return
	}

	l.scroll.top--
	for ah > 0 {
		// Account for the gap between the inserted item and the current top.
		if l.gap > 0 {
			ah -= l.gap
		}
		item := l.Builder(l.scroll.top, l.cursor)
		if item == nil {
			break
		}
		height := l.itemHeight(item, width)
		ah -= height
		entry := listDrawnItem{
			index:  l.scroll.top,
			item:   item,
			row:    ah,
			height: height,
		}
		*children = append([]listDrawnItem{entry}, *children...)

		if l.scroll.top == 0 {
			break
		}
		l.scroll.top--
	}

	l.scroll.offset = ah

	if l.scroll.top == 0 && ah > 0 {
		// We hit the absolute top; normalize rows to avoid overscrolling.
		l.scroll.offset = 0
		row := 0
		for i := range *children {
			child := (*children)[i]
			child.row = row
			(*children)[i] = child
			row += child.height + l.gap
		}
	}
}

func (l *List) ensureScroll() {
	if l.cursor < 0 {
		l.scroll.wantsCursor = false
		return
	}
	if l.cursor < l.scroll.top {
		l.scroll.top = l.cursor
		l.scroll.offset = 0
	}
	l.scroll.wantsCursor = true
}

func (l *List) centerScrollState(width int, height int) (int, int, bool) {
	if l.Builder == nil || l.cursor < 0 || width <= 0 || height <= 0 {
		return 0, 0, false
	}
	cursorItem := l.Builder(l.cursor, l.cursor)
	if cursorItem == nil {
		return 0, 0, false
	}
	cursorHeight := l.itemHeight(cursorItem, width)
	// Compute the space above the cursor so its center aligns to the viewport center.
	targetCenter := height / 2
	desiredBefore := max(targetCenter-cursorHeight/2, 0)

	// Build a top/offset that leaves desiredBefore rows ahead of the cursor.
	top := l.cursor
	offset := 0
	remaining := desiredBefore
	for remaining > 0 && top > 0 {
		prevIndex := top - 1
		prevItem := l.Builder(prevIndex, l.cursor)
		if prevItem == nil {
			break
		}
		prevHeight := l.itemHeight(prevItem, width)
		span := prevHeight
		if l.gap > 0 {
			span += l.gap
		}
		if remaining >= span {
			remaining -= span
			top = prevIndex
			offset = 0
			continue
		}
		top = prevIndex
		if remaining > l.gap {
			// Scroll partway into the previous item if needed.
			withinItem := remaining - l.gap
			offset = max(prevHeight-withinItem, 0)
		} else {
			offset = prevHeight
		}
		remaining = 0
	}

	// If we ran out of items above, skip centering.
	if remaining > 0 {
		return 0, 0, false
	}

	// Verify there is enough content below to keep the viewport filled.
	ah := -offset
	for i := top; ; i++ {
		item := l.Builder(i, l.cursor)
		if item == nil {
			return 0, 0, false
		}
		itemHeight := l.itemHeight(item, width)
		if ah+itemHeight >= height {
			break
		}
		ah += itemHeight + l.gap
	}

	return top, offset, true
}

func (l *List) scrollByItems(delta int, count int, width int, height int) {
	if l.Builder == nil {
		return
	}
	if count < 1 {
		count = 1
	}
	if delta > 0 {
		// Step the top index downward without going past the end.
		for i := 0; i < count; i++ {
			if l.Builder(l.scroll.top+1, l.cursor) == nil {
				break
			}
			l.scroll.top++
		}
	} else {
		// Step the top index upward without going below zero.
		for i := 0; i < count; i++ {
			if l.scroll.top <= 0 {
				break
			}
			l.scroll.top--
		}
	}
	l.scroll.offset = 0
	l.scroll.wantsCursor = false
	l.setLastDraw(nil)
	l.lastRect = listRect{x: 0, y: 0, width: width, height: height}
}

func (l *List) visibleItemCount(width int, height int) int {
	if l.Builder == nil || width <= 0 || height <= 0 {
		return 0
	}
	total := 0
	count := 0
	for idx := l.scroll.top; ; idx++ {
		item := l.Builder(idx, l.cursor)
		if item == nil {
			break
		}
		if count > 0 {
			total += l.gap
		}
		itemHeight := l.itemHeight(item, width)
		if total+itemHeight > height {
			break
		}
		total += itemHeight
		count++
	}
	// Always move at least one item so navigation feels responsive.
	if count == 0 {
		return 1
	}
	return count
}

func (l *List) endScrollState(width int, height int) (int, int) {
	if l.Builder == nil || width <= 0 || height <= 0 {
		return 0, 0
	}
	start := max(l.scroll.top, 0)
	// If the current top is past the end, restart from the beginning.
	if l.Builder(start, l.cursor) == nil && start != 0 {
		start = 0
	}
	last := start
	for {
		if l.Builder(last, l.cursor) == nil {
			last--
			break
		}
		last++
	}
	if last < 0 {
		return 0, 0
	}

	// Walk upward from the last item until we fill a viewport.
	total := 0
	for i := last; i >= 0; i-- {
		item := l.Builder(i, l.cursor)
		if item == nil {
			continue
		}
		if total > 0 {
			total += l.gap
		}
		itemHeight := l.itemHeight(item, width)
		if total+itemHeight > height {
			offset := max(total+itemHeight-height, 0)
			return i, offset
		}
		total += itemHeight
		if i == 0 {
			break
		}
	}
	return 0, 0
}

// InputHandler returns the handler for this primitive.
func (l *List) InputHandler(event *tcell.EventKey) Command {
	switch event.Key() {
	case tcell.KeyDown:
		l.NextItem()
	case tcell.KeyUp:
		l.PrevItem()
	case tcell.KeyPgDn:
		_, _, width, height := l.GetInnerRect()
		if l.snapToItems {
			l.scrollByItems(1, l.visibleItemCount(width, height), width, height)
		} else {
			if height < 1 {
				height = 1
			}
			l.scroll.pending += height
		}
	case tcell.KeyPgUp:
		_, _, width, height := l.GetInnerRect()
		if l.snapToItems {
			l.scrollByItems(-1, l.visibleItemCount(width, height), width, height)
		} else {
			if height < 1 {
				height = 1
			}
			l.scroll.pending -= height
		}
	}
	return BatchCommand{RedrawCommand{}, ConsumeEventCommand{}}
}

// MouseHandler returns the mouse handler for this primitive.
func (l *List) MouseHandler(action MouseAction, event *tcell.EventMouse) (Primitive, Command) {
	var cmd Command
	x, y := event.Position()
	if l.scrollBarInteraction.dragDelta >= 0 {
		_, innerY, innerWidth, innerHeight := l.GetInnerRect()
		contentWidth, _ := l.scrollBarLayout(0, innerWidth)
		row := y - innerY
		switch action {
		case MouseMove:
			l.dragScrollBarTo(row, innerHeight, contentWidth)
			return l, AppendCommand(nil, BatchCommand{RedrawCommand{}, ConsumeEventCommand{}})
		case MouseLeftUp:
			l.scrollBarInteraction.dragDelta = listScrollBarNoDrag
			return nil, BatchCommand{RedrawCommand{}, ConsumeEventCommand{}}
		case MouseLeftClick:
			if l.scrollBarInteraction.dragMoved {
				l.scrollBarInteraction.dragMoved = false
				return nil, BatchCommand{RedrawCommand{}, ConsumeEventCommand{}}
			}
		}
	}

	if !l.InRect(x, y) {
		return nil, nil
	}

	innerX, innerY, innerWidth, innerHeight := l.GetInnerRect()
	contentWidth, scrollBarX := l.scrollBarLayout(innerX, innerWidth)
	drawScrollBar := l.shouldDrawScrollBar(innerWidth, innerHeight)
	if drawScrollBar && x == scrollBarX && y >= innerY && y < innerY+innerHeight {
		row := y - innerY
		switch action {
		case MouseLeftDown:
			cmd = AppendCommand(cmd, SetFocusCommand{Target: l})
			if l.startScrollBarDrag(row, innerHeight, contentWidth) {
				return l, AppendCommand(cmd, BatchCommand{RedrawCommand{}, ConsumeEventCommand{}})
			}
			return nil, BatchCommand{RedrawCommand{}, ConsumeEventCommand{}}
		case MouseLeftClick:
			if l.scrollBarInteraction.dragMoved {
				l.scrollBarInteraction.dragMoved = false
				return nil, BatchCommand{RedrawCommand{}, ConsumeEventCommand{}}
			}
		}
		if l.handleScrollBarMouse(action, row, innerHeight, contentWidth) {
			return nil, BatchCommand{RedrawCommand{}, ConsumeEventCommand{}}
		}
		if action == MouseLeftClick {
			return nil, BatchCommand{RedrawCommand{}, ConsumeEventCommand{}}
		}
	}

	switch action {
	case MouseLeftClick:
		index := l.indexAtPoint(x, y)
		if index >= 0 {
			previous := l.cursor
			l.cursor = index
			l.ensureScroll()
			if l.changed != nil && l.cursor != previous {
				l.changed(l.cursor)
			}
		}
		return nil, BatchCommand{RedrawCommand{}, ConsumeEventCommand{}}
	case MouseScrollUp:
		_, _, width, height := l.GetInnerRect()
		if l.snapToItems {
			l.scrollByItems(-1, 1, width, height)
		} else {
			l.scroll.pending -= l.mouseScrollStep()
		}
		return nil, BatchCommand{RedrawCommand{}, ConsumeEventCommand{}}
	case MouseScrollDown:
		_, _, width, height := l.GetInnerRect()
		if l.snapToItems {
			l.scrollByItems(1, 1, width, height)
		} else {
			l.scroll.pending += l.mouseScrollStep()
		}
		return nil, BatchCommand{RedrawCommand{}, ConsumeEventCommand{}}
	}

	return nil, nil
}

func (l *List) startScrollBarDrag(row int, height int, contentWidth int) bool {
	if l.scrollBar == nil || contentWidth <= 0 || height <= 0 {
		return false
	}
	state, ok := l.currentScrollBarState(height, contentWidth)
	if !ok {
		return false
	}

	trackRow := row
	if l.scrollBar.arrows.hasStart() {
		trackRow--
	}
	if trackRow < 0 || trackRow >= state.metrics.trackCells {
		return false
	}
	clickPos := trackRow*subcell + subcell/2
	if clickPos < state.metrics.thumbStart || clickPos >= state.metrics.thumbStart+state.metrics.thumbLen {
		return false
	}

	l.scrollBarInteraction.dragMoved = false
	l.scrollBarInteraction.dragDelta = clickPos - state.metrics.thumbStart
	return true
}

func (l *List) dragScrollBarTo(row int, height int, contentWidth int) bool {
	if l.scrollBarInteraction.dragDelta < 0 || l.scrollBar == nil || contentWidth <= 0 || height <= 0 {
		return false
	}
	state, ok := l.currentScrollBarState(height, contentWidth)
	if !ok {
		return false
	}

	trackRow := row
	if l.scrollBar.arrows.hasStart() {
		trackRow--
	}
	trackRow = min(max(trackRow, 0), state.metrics.trackCells-1)
	clickPos := trackRow*subcell + subcell/2

	maxOffset := max(state.contentLength-state.viewportLength, 0)
	if maxOffset <= 0 {
		return true
	}
	thumbTravel := max(state.metrics.trackLen-state.metrics.thumbLen, 0)
	if thumbTravel <= 0 {
		return true
	}

	targetStart := clickPos - l.scrollBarInteraction.dragDelta
	targetStart = min(max(targetStart, 0), thumbTravel)
	// Convert thumb start in subcells back to content offset.
	targetOffset := (targetStart * maxOffset) / thumbTravel
	delta := targetOffset - state.position
	if delta != 0 {
		l.scroll.pending += delta
		l.scrollBarInteraction.dragMoved = true
	}
	return true
}

func (l *List) shouldDrawScrollBar(width int, height int) bool {
	if width <= 1 || l.scrollBarVisibility == ScrollBarVisibilityNever {
		return false
	}
	switch l.scrollBarVisibility {
	case ScrollBarVisibilityAlways:
		return true
	case ScrollBarVisibilityAutomatic:
		state := l.scrollBarInteraction.state
		if state.contentWidth == width &&
			state.viewportHeight == height &&
			state.contentLength > 0 &&
			state.viewportLength > 0 &&
			state.metrics.trackCells > 0 {
			return state.contentLength > state.viewportLength
		}
		return l.totalContentHeight(width) > height
	default:
		return false
	}
}

func (l *List) mouseScrollStep() int {
	step := 3
	if l.scrollBar != nil && l.scrollBar.scrollStep > 0 {
		step = l.scrollBar.scrollStep
	}
	return step
}

func (l *List) handleScrollBarMouse(action MouseAction, row int, height int, contentWidth int) bool {
	if l.scrollBar == nil || contentWidth <= 0 || height <= 0 {
		return false
	}
	state, ok := l.currentScrollBarState(height, contentWidth)
	if !ok {
		return false
	}

	row = max(row, 0)
	startArrow := l.scrollBar.arrows.hasStart()
	endArrow := l.scrollBar.arrows.hasEnd()
	trackRow := row
	if startArrow {
		if row == 0 {
			if action == MouseLeftClick {
				l.scroll.pending -= l.mouseScrollStep()
			}
			return true
		}
		trackRow--
	}
	if endArrow {
		endRow := state.metrics.trackCells
		if startArrow {
			endRow++
		}
		if row == endRow {
			if action == MouseLeftClick {
				l.scroll.pending += l.mouseScrollStep()
			}
			return true
		}
	}
	if trackRow < 0 || trackRow >= state.metrics.trackCells || action != MouseLeftClick {
		return false
	}

	clickPos := trackRow*subcell + subcell/2
	maxOffset := max(state.contentLength-state.viewportLength, 0)
	if maxOffset <= 0 {
		return true
	}

	switch l.scrollBar.trackClickBehavior {
	case TrackClickBehaviorJumpToClick:
		thumbTravel := max(state.metrics.trackLen-state.metrics.thumbLen, 0)
		if thumbTravel == 0 {
			l.scroll.pending -= state.position
			return true
		}
		targetStart := clickPos - state.metrics.thumbLen/2
		targetStart = min(max(targetStart, 0), thumbTravel)
		targetOffset := (targetStart * maxOffset) / thumbTravel
		l.scroll.pending += targetOffset - state.position
	default:
		if clickPos < state.metrics.thumbStart {
			l.scroll.pending -= state.viewportLength
		} else if clickPos >= state.metrics.thumbStart+state.metrics.thumbLen {
			l.scroll.pending += state.viewportLength
		}
	}
	return true
}

func (l *List) currentScrollBarState(height int, contentWidth int) (listScrollBarState, bool) {
	state := l.scrollBarInteraction.state
	// Reuse cached geometry while viewport/content width is unchanged.
	if state.viewportHeight == height &&
		state.contentWidth == contentWidth &&
		state.contentLength > 0 &&
		state.viewportLength > 0 &&
		state.metrics.trackCells > 0 {
		return state, true
	}
	state, ok := l.computeScrollBarState(contentWidth, height, l.lastDraw)
	if ok {
		l.scrollBarInteraction.state = state
	}
	return state, ok
}

func (l *List) scrollBarLayout(innerX int, innerWidth int) (contentWidth int, scrollBarX int) {
	contentWidth = innerWidth - 1
	scrollBarX = innerX + contentWidth
	// Reuse right padding for the scrollBar when available so we don't reduce content width by an extra column.
	if l.paddingRight > 0 {
		contentWidth = innerWidth
		scrollBarX = innerX + innerWidth + l.paddingRight - 1
	}
	return contentWidth, scrollBarX
}

func (l *List) computeScrollBarState(contentWidth int, viewportHeight int, children []listDrawnItem) (listScrollBarState, bool) {
	state := listScrollBarState{
		contentWidth:   contentWidth,
		viewportHeight: viewportHeight,
	}
	if l.scrollBar == nil || contentWidth <= 0 || viewportHeight <= 0 {
		return state, false
	}
	position, contentLength, viewportLength := l.scrollBarMetrics(contentWidth, viewportHeight, children)
	if contentLength <= 0 || viewportLength <= 0 {
		return state, false
	}
	maxOffset := max(contentLength-viewportLength, 0)
	// Include pending delta so interactions stay in sync with the next drawn frame.
	position = min(max(position+l.scroll.pending, 0), maxOffset)

	trackCells := l.scrollBar.trackLengthExcludingArrowHeads(viewportHeight)
	metrics := computeScrollMetrics(trackCells, contentLength, viewportLength, position)
	if metrics.trackCells <= 0 {
		return state, false
	}

	state.position = position
	state.contentLength = contentLength
	state.viewportLength = viewportLength
	state.metrics = metrics
	return state, true
}

func (l *List) indexAtPoint(x, y int) int {
	if len(l.lastDraw) == 0 {
		return -1
	}
	if x < l.lastRect.x || x >= l.lastRect.x+l.lastRect.width || y < l.lastRect.y || y >= l.lastRect.y+l.lastRect.height {
		return -1
	}

	row := y - l.lastRect.y
	for _, child := range l.lastDraw {
		span := child.height
		if l.gap > 0 {
			span += l.gap
		}
		if row >= child.row && row < child.row+span {
			return child.index
		}
	}
	return -1
}

var _ Primitive = &List{}

type clippedScreen struct {
	tcell.Screen
	x      int
	y      int
	width  int
	height int
}

func newClippedScreen(screen tcell.Screen, x, y, width, height int) *clippedScreen {
	return &clippedScreen{
		Screen: screen,
		x:      x,
		y:      y,
		width:  width,
		height: height,
	}
}

func (s *clippedScreen) inBounds(x, y int) bool {
	return x >= s.x && x < s.x+s.width && y >= s.y && y < s.y+s.height
}

func (s *clippedScreen) SetContent(x int, y int, primary rune, combining []rune, style tcell.Style) {
	if !s.inBounds(x, y) {
		return
	}
	s.Screen.SetContent(x, y, primary, combining, style)
}

func (s *clippedScreen) Put(x int, y int, str string, style tcell.Style) (string, int) {
	if !s.inBounds(x, y) {
		return str, 0
	}
	return s.Screen.Put(x, y, str, style)
}

func (s *clippedScreen) PutStr(x int, y int, str string) {
	s.PutStrStyled(x, y, str, tcell.StyleDefault)
}

func (s *clippedScreen) PutStrStyled(x int, y int, str string, style tcell.Style) {
	if y < s.y || y >= s.y+s.height {
		return
	}

	gr := uniseg.NewGraphemes(str)
	for gr.Next() {
		cluster := gr.Str()
		width := max(uniseg.StringWidth(cluster), 1)
		if x >= s.x+s.width {
			return
		}
		if x >= s.x && x+width <= s.x+s.width {
			s.Screen.Put(x, y, cluster, style)
		}
		x += width
	}
}

func (s *clippedScreen) ShowCursor(x int, y int) {
	if !s.inBounds(x, y) {
		s.Screen.ShowCursor(-1, -1)
		return
	}
	s.Screen.ShowCursor(x, y)
}

func (l *List) trimToFullItems(children []listDrawnItem, height int) []listDrawnItem {
	if len(children) == 0 {
		return children
	}

	// Drop any items that start above the viewport.
	start := 0
	for start < len(children) && children[start].row < 0 {
		start++
	}
	if start > 0 {
		children = children[start:]
	}
	if len(children) == 0 {
		return children
	}

	// Realign the first item to row 0 so we can fill below it.
	shift := -children[0].row
	if shift != 0 {
		for i := range children {
			children[i].row += shift
		}
	}

	// Trim trailing items that don't fully fit.
	end := len(children)
	for end > 0 && children[end-1].row+children[end-1].height > height {
		end--
	}
	children = children[:end]

	return children
}
