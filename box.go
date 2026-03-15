package tview

import "github.com/gdamore/tcell/v3"

// Box implements the Primitive interface with an empty background and optional
// elements such as a border and a title. Box itself does not hold any content
// but serves as the superclass of all other primitives. Subclasses add their
// own content, typically (but not necessarily) keeping their content within the
// box's rectangle.
//
// Box provides a number of utility functions available to all primitives.
//
// See https://github.com/ayn2op/tview/wiki/Box for an example.
type Box struct {
	// The position of the rect.
	x, y, width, height int

	// The inner rect reserved for the box's content. If innerX is negative,
	// the rect is undefined and must be calculated.
	innerX, innerY, innerWidth, innerHeight int

	// Border padding.
	paddingTop, paddingBottom, paddingLeft, paddingRight int

	// The box's background color.
	backgroundColor tcell.Color

	// If set to true, the background of this box is not cleared while drawing.
	dontClear bool

	// Border
	borders     Borders
	borderSet   BorderSet
	borderStyle tcell.Style

	// Title
	title          string
	titleStyle     tcell.Style
	titleAlignment Alignment

	// Footer
	footer          string
	footerStyle     tcell.Style
	footerAlignment Alignment

	// Whether or not this box has focus. This is typically ignored for
	// container primitives (e.g. Flex, Grid, Layers), as they will delegate
	// focus to their children.
	hasFocus bool

	// Optional callback functions invoked when the primitive receives or loses
	// focus.
	focus, blur func()
}

// NewBox returns a Box without a border.
func NewBox() *Box {
	b := &Box{
		width:           15,
		height:          10,
		innerX:          -1, // Mark as uninitialized.
		backgroundColor: Styles.PrimitiveBackgroundColor,

		borderStyle: tcell.StyleDefault.Foreground(Styles.BorderColor).Background(Styles.PrimitiveBackgroundColor),
		borderSet:   BorderSetPlain(),

		titleStyle:      tcell.StyleDefault.Foreground(Styles.TitleColor),
		titleAlignment:  AlignmentCenter,
		footerStyle:     tcell.StyleDefault.Foreground(Styles.TitleColor),
		footerAlignment: AlignmentCenter,
	}
	return b
}

// SetBorderPadding sets the size of the borders around the box content.
func (b *Box) SetBorderPadding(top, bottom, left, right int) *Box {
	if b.paddingTop != top || b.paddingBottom != bottom || b.paddingLeft != left || b.paddingRight != right {
		b.paddingTop, b.paddingBottom, b.paddingLeft, b.paddingRight = top, bottom, left, right
		b.innerX = -1 // Mark inner rect as uninitialized.
	}
	return b
}

// GetBorderPadding returns the configured border padding.
func (b *Box) GetBorderPadding() (top, bottom, left, right int) {
	return b.paddingTop, b.paddingBottom, b.paddingLeft, b.paddingRight
}

// GetRect returns the current position of the rectangle, x, y, width, and
// height.
func (b *Box) GetRect() (int, int, int, int) {
	return b.x, b.y, b.width, b.height
}

// GetInnerRect returns the position of the inner rectangle (x, y, width,
// height), without the border and without any padding. Width and height values
// will clamp to 0 and thus never be negative.
func (b *Box) GetInnerRect() (int, int, int, int) {
	if b.innerX >= 0 {
		return b.innerX, b.innerY, b.innerWidth, b.innerHeight
	}

	x, y, width, height := b.GetRect()

	if b.title != "" || b.borders.Has(BordersTop) {
		y++
		height--
	}

	if b.footer != "" || b.borders.Has(BordersBottom) {
		height--
	}

	if b.borders.Has(BordersLeft) {
		x++
		width--
	}

	if b.borders.Has(BordersRight) {
		width--
	}

	x += b.paddingLeft
	y += b.paddingTop
	width -= (b.paddingLeft + b.paddingRight)
	height -= (b.paddingTop + b.paddingBottom)
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	return x, y, width, height
}

// SetRect sets a new position of the primitive. Note that this has no effect
// if this primitive is part of a layout (e.g. Flex, Grid) or if it was added
// like this:
//
//	application.SetRoot(p, true)
func (b *Box) SetRect(x, y, width, height int) {
	if b.x != x || b.y != y || b.width != width || b.height != height {
		b.x = x
		b.y = y
		b.width = width
		b.height = height
		b.innerX = -1 // Mark inner rect as uninitialized.
	}
}

// HandleEvent handles input events for this primitive.
func (b *Box) HandleEvent(event tcell.Event) Command {
	switch event := event.(type) {
	case *MouseEvent:
		if event.Action == MouseLeftDown && b.InRect(event.Position()) {
			return SetFocus(b)
		}
	}
	return nil
}

// InRect returns true if the given coordinate is within the bounds of the box's
// rectangle.
func (b *Box) InRect(x, y int) bool {
	rectX, rectY, width, height := b.GetRect()
	return x >= rectX && x < rectX+width && y >= rectY && y < rectY+height
}

// InInnerRect returns true if the given coordinate is within the bounds of the
// box's inner rectangle (within the border and padding).
func (b *Box) InInnerRect(x, y int) bool {
	rectX, rectY, width, height := b.GetInnerRect()
	return x >= rectX && x < rectX+width && y >= rectY && y < rectY+height
}

// SetDontClear sets whether drawing should skip clearing the background.
func (b *Box) SetDontClear(dontClear bool) *Box {
	if b.dontClear != dontClear {
		b.dontClear = dontClear
	}
	return b
}

// SetBackgroundColor sets the box's background color.
func (b *Box) SetBackgroundColor(color tcell.Color) *Box {
	if b.backgroundColor != color {
		b.backgroundColor = color
		b.borderStyle = b.borderStyle.Background(color)
	}
	return b
}

// GetBorders returns the borders.
func (b *Box) GetBorders() Borders {
	return b.borders
}

// SetBorders sets which borders to draw.
func (b *Box) SetBorders(flag Borders) *Box {
	if b.borders != flag {
		b.borders = flag
		b.innerX = -1 // Mark inner rect as uninitialized.
	}
	return b
}

// SetBorderSet sets the border set.
func (b *Box) SetBorderSet(borderSet BorderSet) *Box {
	if b.borderSet != borderSet {
		b.borderSet = borderSet
	}
	return b
}

// GetBorderSet returns the border set.
func (b *Box) GetBorderSet() BorderSet {
	return b.borderSet
}

// SetBorderStyle sets the box's border style.
func (b *Box) SetBorderStyle(style tcell.Style) *Box {
	if b.borderStyle != style {
		b.borderStyle = style
	}
	return b
}

// GetBackgroundColor returns the box's background color.
func (b *Box) GetBackgroundColor() tcell.Color {
	return b.backgroundColor
}

// GetTitle returns the box's current title.
func (b *Box) GetTitle() string {
	return b.title
}

// SetTitle sets the box's title.
func (b *Box) SetTitle(title string) *Box {
	if b.title != title {
		b.title = title
		b.innerX = -1 // Mark inner rect as uninitialized.
	}
	return b
}

// SetTitleStyle sets the style of the title.
func (b *Box) SetTitleStyle(style tcell.Style) *Box {
	if b.titleStyle != style {
		b.titleStyle = style
	}
	return b
}

// SetTitleAlignment sets the alignment of the title.
func (b *Box) SetTitleAlignment(alignment Alignment) *Box {
	if b.titleAlignment != alignment {
		b.titleAlignment = alignment
	}
	return b
}

// GetFooter returns the box's current footer.
func (b *Box) GetFooter() string {
	return b.footer
}

// SetFooter sets the box's footer.
func (b *Box) SetFooter(footer string) *Box {
	if b.footer != footer {
		b.footer = footer
		b.innerX = -1 // Mark inner rect as uninitialized.
	}
	return b
}

// SetFooterStyle sets the style of the footer.
func (b *Box) SetFooterStyle(style tcell.Style) *Box {
	if b.footerStyle != style {
		b.footerStyle = style
	}
	return b
}

// SetFooterAlignment sets the alignment of the footer.
func (b *Box) SetFooterAlignment(alignment Alignment) *Box {
	if b.footerAlignment != alignment {
		b.footerAlignment = alignment
	}
	return b
}

// Draw draws this primitive onto the screen.
func (b *Box) Draw(screen tcell.Screen) {
	b.DrawForSubclass(screen, b)
}

// DrawForSubclass draws this box under the assumption that primitive p is a
// subclass of this box. This is needed e.g. to draw proper box frames which
// depend on the subclass's focus.
//
// Only call this function from your own custom primitives. It is not needed in
// applications that have no custom primitives.
func (b *Box) DrawForSubclass(screen tcell.Screen, p Primitive) {
	// Don't draw anything if there is no space.
	if b.width <= 0 || b.height <= 0 {
		return
	}

	// Fill background.
	background := tcell.StyleDefault.Background(b.backgroundColor)
	if !b.dontClear {
		for y := b.y; y < b.y+b.height; y++ {
			for x := b.x; x < b.x+b.width; x++ {
				screen.Put(x, y, " ", background)
			}
		}
	}

	// Draw border.
	if b.borders != BordersNone && b.width >= 2 && b.height >= 2 {
		if b.borders.Has(BordersTop) {
			for x := b.x + 1; x < b.x+b.width-1; x++ {
				screen.Put(x, b.y, b.borderSet.Top, b.borderStyle)
			}
		}

		if b.borders.Has(BordersBottom) {
			for x := b.x + 1; x < b.x+b.width-1; x++ {
				screen.Put(x, b.y+b.height-1, b.borderSet.Bottom, b.borderStyle)
			}
		}

		if b.borders.Has(BordersLeft) {
			for y := b.y + 1; y < b.y+b.height-1; y++ {
				screen.Put(b.x, y, b.borderSet.Left, b.borderStyle)
			}
		}

		if b.borders.Has(BordersRight) {
			for y := b.y + 1; y < b.y+b.height-1; y++ {
				screen.Put(b.x+b.width-1, y, b.borderSet.Right, b.borderStyle)
			}
		}

		if b.borders.Has(BordersTop | BordersLeft) {
			screen.Put(b.x, b.y, b.borderSet.TopLeft, b.borderStyle)
		}

		if b.borders.Has(BordersTop | BordersRight) {
			screen.Put(b.x+b.width-1, b.y, b.borderSet.TopRight, b.borderStyle)
		}

		if b.borders.Has(BordersBottom | BordersLeft) {
			screen.Put(b.x, b.y+b.height-1, b.borderSet.BottomLeft, b.borderStyle)
		}

		if b.borders.Has(BordersBottom | BordersRight) {
			screen.Put(b.x+b.width-1, b.y+b.height-1, b.borderSet.BottomRight, b.borderStyle)
		}
	}

	// Draw title.
	if b.title != "" && b.width >= 4 {
		start, end, _ := printWithStyle(screen, b.title, b.x+1, b.y, 0, b.width-2, b.titleAlignment, b.titleStyle, true)
		printed := end - start
		if len(b.title)-printed > 0 && printed > 0 {
			xEllipsis := b.x + b.width - 2
			if b.titleAlignment == AlignmentRight {
				xEllipsis = b.x + 1
			}
			_, style, _ := screen.Get(xEllipsis, b.y)
			fg := style.GetForeground()
			Print(screen, string(SemigraphicsHorizontalEllipsis), xEllipsis, b.y, 1, AlignmentLeft, fg)
		}
	}

	// Draw footer.
	if b.footer != "" && b.width >= 4 {
		start, end, _ := printWithStyle(screen, b.footer, b.x+1, b.y+b.height-1, 0, b.width-2, b.footerAlignment, b.footerStyle, true)
		printed := end - start
		if len(b.footer)-printed > 0 && printed > 0 {
			xEllipsis := b.x + b.width - 2
			if b.footerAlignment == AlignmentRight {
				xEllipsis = b.x + 1
			}
			_, style, _ := screen.Get(xEllipsis, b.y+b.height-1)
			fg := style.GetForeground()
			Print(screen, string(SemigraphicsHorizontalEllipsis), xEllipsis, b.y+b.height-1, 1, AlignmentLeft, fg)
		}
	}

	// Remember the inner rect.
	b.innerX = -1
	b.innerX, b.innerY, b.innerWidth, b.innerHeight = b.GetInnerRect()
}

// SetFocusFunc sets a callback function which is invoked when this primitive
// receives focus. Container primitives such as [Flex] or [Grid] will also be
// notified if one of their descendents receive focus directly. Note that this
// may result in a blur notification, immediately followed by a focus
// notification, when the focus is set to a different descendent of the
// container primitive.
//
// At this point, the order in which the focus callbacks are invoked during one
// draw cycle, is not defined. However, the blur callbacks are always invoked
// before the focus callbacks.
//
// Set to nil to remove the callback function.
func (b *Box) SetFocusFunc(callback func()) *Box {
	b.focus = callback
	return b
}

// SetBlurFunc sets a callback function which is invoked when this primitive
// loses focus. Container primitives such as [Flex] or [Grid] will also be
// notified if one of their descendents lose focus. Note that this may result in
// a blur notification, immediately followed by a focus notification, when the
// focus is set to a different descendent of the container primitive.
//
// At this point, the order in which the blur callbacks are invoked during one
// draw cycle, is not defined. However, the blur callbacks are always invoked
// before the focus callbacks.
//
// Set to nil to remove the callback function.
func (b *Box) SetBlurFunc(callback func()) *Box {
	b.blur = callback
	return b
}

// Focus is called when this primitive directly receives focus.
func (b *Box) Focus(delegate func(p Primitive)) {
	if !b.hasFocus {
		b.hasFocus = true
	}
	if b.focus != nil {
		b.focus()
	}
}

// Blur is called when this primitive directly loses focus.
func (b *Box) Blur() {
	if b.hasFocus {
		b.hasFocus = false
	}
	if b.blur != nil {
		b.blur()
	}
}

// HasFocus returns whether or not this primitive has focus.
func (b *Box) HasFocus() bool {
	return b.hasFocus
}
