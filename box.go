package tview

import "github.com/gdamore/tcell/v3"

// Box implements the Model interface with an empty background and optional
// elements such as a border and a title. Box itself does not hold any content
// but serves as the superclass of all other models. Subclasses add their
// own content, typically (but not necessarily) keeping their content within the
// box's rectangle.
//
// Box provides a number of utility functions available to all models.
type Box struct {
	// The position of the rect.
	x, y, width, height int

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

	hasFocus bool
}

// NewBox returns a Box without a border.
func NewBox() *Box {
	return &Box{
		width:           15,
		height:          10,
		backgroundColor: Styles.PrimitiveBackgroundColor,

		borderStyle: tcell.StyleDefault.Foreground(Styles.BorderColor).Background(Styles.PrimitiveBackgroundColor),
		borderSet:   BorderSetPlain(),

		titleStyle:      tcell.StyleDefault.Foreground(Styles.TitleColor),
		titleAlignment:  AlignmentCenter,
		footerStyle:     tcell.StyleDefault.Foreground(Styles.TitleColor),
		footerAlignment: AlignmentCenter,
	}
}

// BorderPadding returns the configured border padding.
func (b *Box) BorderPadding() (top, bottom, left, right int) {
	return b.paddingTop, b.paddingBottom, b.paddingLeft, b.paddingRight
}

// SetBorderPadding sets the size of the borders around the box content.
func (b *Box) SetBorderPadding(top, bottom, left, right int) *Box {
	if b.paddingTop != top || b.paddingBottom != bottom || b.paddingLeft != left || b.paddingRight != right {
		b.paddingTop, b.paddingBottom, b.paddingLeft, b.paddingRight = top, bottom, left, right
	}
	return b
}

// Rect returns the current position of the rectangle, x, y, width, and
// height.
func (b *Box) Rect() (int, int, int, int) {
	return b.x, b.y, b.width, b.height
}

// SetRect sets the model's position. Layouts and Application may override it.
func (b *Box) SetRect(x, y, width, height int) {
	if b.x != x || b.y != y || b.width != width || b.height != height {
		b.x = x
		b.y = y
		b.width = width
		b.height = height
	}
}

// InnerRect returns the position of the inner rectangle (x, y, width,
// height), without the border and without any padding. Width and height values
// will clamp to 0 and thus never be negative.
func (b *Box) InnerRect() (int, int, int, int) {
	x, y, width, height := b.Rect()

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

var _ Model = (*Box)(nil)

func (b *Box) Update(msg Msg) Cmd {
	switch msg.(type) {
	case FocusMsg:
		b.hasFocus = true
	case BlurMsg:
		b.hasFocus = false
	}
	return nil
}

func (b *Box) View(screen tcell.Screen) {
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
		start, end, _ := PrintStyled(screen, b.title, b.x+1, b.y, 0, b.width-2, b.titleAlignment, b.titleStyle, true)
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
		start, end, _ := PrintStyled(screen, b.footer, b.x+1, b.y+b.height-1, 0, b.width-2, b.footerAlignment, b.footerStyle, true)
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

}

// InRect returns true if the given coordinate is within the bounds of the box's
// rectangle.
func (b *Box) InRect(x, y int) bool {
	return ModelInRect(b, x, y)
}

// InInnerRect returns true if the given coordinate is within the bounds of the
// box's inner rectangle (within the border and padding).
func (b *Box) InInnerRect(x, y int) bool {
	rectX, rectY, width, height := b.InnerRect()
	return x >= rectX && x < rectX+width && y >= rectY && y < rectY+height
}

// SetDontClear sets whether drawing should skip clearing the background.
func (b *Box) SetDontClear(dontClear bool) *Box {
	b.dontClear = dontClear
	return b
}

// BackgroundColor returns the box's background color.
func (b *Box) BackgroundColor() tcell.Color {
	return b.backgroundColor
}

// SetBackgroundColor sets the box's background color.
func (b *Box) SetBackgroundColor(color tcell.Color) *Box {
	if b.backgroundColor != color {
		b.backgroundColor = color
		b.borderStyle = b.borderStyle.Background(color)
	}
	return b
}

// Borders returns the borders.
func (b *Box) Borders() Borders {
	return b.borders
}

// SetBorders sets which borders to draw.
func (b *Box) SetBorders(flag Borders) *Box {
	if b.borders != flag {
		b.borders = flag
	}
	return b
}

// BorderSet returns the border set.
func (b *Box) BorderSet() BorderSet {
	return b.borderSet
}

// SetBorderSet sets the border set.
func (b *Box) SetBorderSet(borderSet BorderSet) *Box {
	b.borderSet = borderSet
	return b
}

// SetBorderStyle sets the box's border style.
func (b *Box) SetBorderStyle(style tcell.Style) *Box {
	b.borderStyle = style
	return b
}

// Title returns the box's current title.
func (b *Box) Title() string {
	return b.title
}

// SetTitle sets the box's title.
func (b *Box) SetTitle(title string) *Box {
	if b.title != title {
		b.title = title
	}
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

// Footer returns the box's current footer.
func (b *Box) Footer() string {
	return b.footer
}

// SetFooter sets the box's footer.
func (b *Box) SetFooter(footer string) *Box {
	if b.footer != footer {
		b.footer = footer
	}
	return b
}

// SetFooterStyle sets the style of the footer.
func (b *Box) SetFooterStyle(style tcell.Style) *Box {
	b.footerStyle = style
	return b
}

// SetFooterAlignment sets the alignment of the footer.
func (b *Box) SetFooterAlignment(alignment Alignment) *Box {
	b.footerAlignment = alignment
	return b
}

// HasFocus returns whether or not this model has focus.
func (b *Box) HasFocus() bool {
	return b.hasFocus
}
