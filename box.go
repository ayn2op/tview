package tview

import (
	"github.com/gdamore/tcell/v3"
)

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

	// Whether or not this box has focus. This is typically ignored for
	// container primitives (e.g. Flex, Grid, Pages), as they will delegate
	// focus to their children.
	hasFocus bool

	// Optional callback functions invoked when the primitive receives or loses
	// focus.
	focus, blur func()

	// An optional capture function which receives a key event and returns the
	// event to be forwarded to the primitive's default input handler (nil if
	// nothing should be forwarded).
	inputCapture func(event *tcell.EventKey) *tcell.EventKey

	// An optional function which is called before the box is drawn.
	draw func(screen tcell.Screen, x, y, width, height int) (int, int, int, int)

	// An optional capture function which receives a mouse event and returns the
	// event to be forwarded to the primitive's default mouse event handler (at
	// least one nil if nothing should be forwarded).
	mouseCapture func(action MouseAction, event *tcell.EventMouse) (MouseAction, *tcell.EventMouse)
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

		titleStyle:     tcell.StyleDefault.Foreground(Styles.TitleColor),
		titleAlignment: AlignmentCenter,
	}
	return b
}

// SetBorderPadding sets the size of the borders around the box content.
func (b *Box) SetBorderPadding(top, bottom, left, right int) *Box {
	b.paddingTop, b.paddingBottom, b.paddingLeft, b.paddingRight = top, bottom, left, right
	return b
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

	if b.borders.Has(BordersBottom) {
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
	b.x = x
	b.y = y
	b.width = width
	b.height = height
	b.innerX = -1 // Mark inner rect as uninitialized.
}

// SetDrawFunc sets a callback function which is invoked after the box primitive
// has been drawn. This allows you to add a more individual style to the box
// (and all primitives which extend it).
//
// The function is provided with the box's dimensions (set via SetRect()). It
// must return the box's inner dimensions (x, y, width, height) which will be
// returned by GetInnerRect(), used by descendent primitives to draw their own
// content.
func (b *Box) SetDrawFunc(handler func(screen tcell.Screen, x, y, width, height int) (int, int, int, int)) *Box {
	b.draw = handler
	return b
}

// GetDrawFunc returns the callback function which was installed with
// SetDrawFunc() or nil if no such function has been installed.
func (b *Box) GetDrawFunc() func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
	return b.draw
}

// WrapInputHandler wraps an input handler (see [Box.InputHandler]) with the
// functionality to capture input (see [Box.SetInputCapture]) before passing it
// on to the provided (default) input handler.
//
// This is only meant to be used by subclassing primitives.
func (b *Box) WrapInputHandler(inputHandler func(*tcell.EventKey, func(p Primitive))) func(*tcell.EventKey, func(p Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p Primitive)) {
		if b.inputCapture != nil {
			event = b.inputCapture(event)
		}
		if event != nil && inputHandler != nil {
			inputHandler(event, setFocus)
		}
	}
}

// InputHandler returns nil. Box has no default input handling.
func (b *Box) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return b.WrapInputHandler(nil)
}

// WrapPasteHandler wraps a paste handler (see [Box.PasteHandler]).
func (b *Box) WrapPasteHandler(pasteHandler func(string, func(p Primitive))) func(string, func(p Primitive)) {
	return func(text string, setFocus func(p Primitive)) {
		if pasteHandler != nil {
			pasteHandler(text, setFocus)
		}
	}
}

// PasteHandler returns nil. Box has no default paste handling.
func (b *Box) PasteHandler() func(pastedText string, setFocus func(p Primitive)) {
	return b.WrapPasteHandler(nil)
}

// SetInputCapture installs a function which captures key events before they are
// forwarded to the primitive's default key event handler. This function can
// then choose to forward that key event (or a different one) to the default
// handler by returning it. If nil is returned, the default handler will not
// be called.
//
// Providing a nil handler will remove a previously existing handler.
//
// This function can also be used on container primitives (like Flex, Grid, or
// Form) as keyboard events will be handed down until they are handled.
//
// Pasted key events are not forwarded to the input capture function if pasting
// is enabled (see [Application.EnablePaste]).
func (b *Box) SetInputCapture(capture func(event *tcell.EventKey) *tcell.EventKey) *Box {
	b.inputCapture = capture
	return b
}

// GetInputCapture returns the function installed with SetInputCapture() or nil
// if no such function has been installed.
func (b *Box) GetInputCapture() func(event *tcell.EventKey) *tcell.EventKey {
	return b.inputCapture
}

// WrapMouseHandler wraps a mouse event handler (see [Box.MouseHandler]) with the
// functionality to capture mouse events (see [Box.SetMouseCapture]) before passing
// them on to the provided (default) event handler.
//
// This is only meant to be used by subclassing primitives.
func (b *Box) WrapMouseHandler(mouseHandler func(MouseAction, *tcell.EventMouse, func(p Primitive)) (bool, Primitive)) func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
	return func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
		if b.mouseCapture != nil {
			action, event = b.mouseCapture(action, event)
		}
		if event == nil {
			if action == MouseConsumed {
				consumed = true
			}
		} else if mouseHandler != nil {
			consumed, capture = mouseHandler(action, event, setFocus)
		}
		return
	}
}

// MouseHandler returns nil. Box has no default mouse handling.
func (b *Box) MouseHandler() func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
	return b.WrapMouseHandler(func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
		if action == MouseLeftDown && b.InRect(event.Position()) {
			setFocus(b)
			consumed = true
		}
		return
	})
}

// SetMouseCapture sets a function which captures mouse events (consisting of
// the original tcell mouse event and the semantic mouse action) before they are
// forwarded to the primitive's default mouse event handler. This function can
// then choose to forward that event (or a different one) by returning it or
// returning a nil mouse event, in which case the default handler will not be
// called.
//
// When a nil event is returned, the returned mouse action value may be set to
// [MouseConsumed] to indicate that the event was consumed and the screen should
// be redrawn. Any other value will not cause a redraw.
//
// Providing a nil handler will remove a previously existing handler.
//
// Note that mouse events are ignored completely if the application has not been
// enabled for mouse events (see [Application.EnableMouse]), which is the
// default.
func (b *Box) SetMouseCapture(capture func(action MouseAction, event *tcell.EventMouse) (MouseAction, *tcell.EventMouse)) *Box {
	b.mouseCapture = capture
	return b
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

// GetMouseCapture returns the function installed with SetMouseCapture() or nil
// if no such function has been installed.
func (b *Box) GetMouseCapture() func(action MouseAction, event *tcell.EventMouse) (MouseAction, *tcell.EventMouse) {
	return b.mouseCapture
}

// SetBackgroundColor sets the box's background color.
func (b *Box) SetBackgroundColor(color tcell.Color) *Box {
	b.backgroundColor = color
	b.borderStyle = b.borderStyle.Background(color)
	return b
}

// GetBorders returns the borders.
func (b *Box) GetBorders() Borders {
	return b.borders
}

// SetBorders sets which borders to draw.
func (b *Box) SetBorders(flag Borders) *Box {
	b.borders = flag
	return b
}

// SetBorderSet sets the box' borderset
func (b *Box) SetBorderSet(borderSet BorderSet) *Box {
	b.borderSet = borderSet
	return b
}

// GetBorderSet returns the box' borderSet
func (b *Box) GetBorderSet() BorderSet {
	return b.borderSet
}

// SetBorderStyle sets the box's border style.
func (b *Box) SetBorderStyle(style tcell.Style) *Box {
	b.borderStyle = style
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
	b.title = title
	return b
}

// SetTitleStyle sets the style of the title.
func (b *Box) SetTitleStyle(style tcell.Style) *Box {
	b.titleStyle = style
	return b
}

// SetTitleAlignment sets the alignment of the title.
func (b *Box) SetTitleAlignment(alignment Alignment) *Box {
	b.titleAlignment = alignment
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

	// Call custom draw function.
	if b.draw != nil {
		b.innerX, b.innerY, b.innerWidth, b.innerHeight = b.draw(screen, b.x, b.y, b.width, b.height)
	} else {
		// Remember the inner rect.
		b.innerX = -1
		b.innerX, b.innerY, b.innerWidth, b.innerHeight = b.GetInnerRect()
	}
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
// focus is set to a different different descendent of the container primitive.
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
	b.hasFocus = true
	if b.focus != nil {
		b.focus()
	}
}

// Blur is called when this primitive directly loses focus.
func (b *Box) Blur() {
	if b.blur != nil {
		b.blur()
	}
	b.hasFocus = false
}

// HasFocus returns whether or not this primitive has focus.
func (b *Box) HasFocus() bool {
	return b.hasFocus
}
