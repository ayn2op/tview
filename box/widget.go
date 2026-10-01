// Package box draws a border with a title and footer around an element.
package box

import (
	"github.com/ayn2op/tview"
	"github.com/gdamore/tcell/v3"
	"github.com/rivo/uniseg"
)

// Widget draws a background, border, title, and footer, with its child inside them.
type Widget struct {
	child         tview.Element
	width, height tview.Length

	background                      tcell.Color
	borders                         tview.Borders
	borderSet                       tview.BorderSet
	borderStyle                     tcell.Style
	paddingTop, paddingBottom       int
	paddingLeft, paddingRight       int
	title, footer                   string
	titleStyle, footerStyle         tcell.Style
	titleAlignment, footerAlignment tview.Alignment
}

var _ tview.Element = Widget{}

// New returns a box around child, without a border, that fills its parent.
func New(child tview.Element) Widget {
	return Widget{
		child:           child,
		width:           tview.Fill,
		height:          tview.Fill,
		borderSet:       tview.BorderSetPlain(),
		titleAlignment:  tview.AlignmentCenter,
		footerAlignment: tview.AlignmentCenter,
	}
}

// Width sets the width of the box.
func (w Widget) Width(width tview.Length) Widget {
	w.width = width
	return w
}

// Height sets the height of the box.
func (w Widget) Height(height tview.Length) Widget {
	w.height = height
	return w
}

// Background sets the background color.
func (w Widget) Background(color tcell.Color) Widget {
	w.background = color
	return w
}

// Borders sets which sides have a border.
func (w Widget) Borders(borders tview.Borders) Widget {
	w.borders = borders
	return w
}

// BorderSet sets the characters the border is drawn with.
func (w Widget) BorderSet(set tview.BorderSet) Widget {
	w.borderSet = set
	return w
}

// BorderStyle sets the style of the border.
func (w Widget) BorderStyle(style tcell.Style) Widget {
	w.borderStyle = style
	return w
}

// Padding sets the empty cells between the border and the child.
func (w Widget) Padding(top, bottom, left, right int) Widget {
	w.paddingTop, w.paddingBottom, w.paddingLeft, w.paddingRight = top, bottom, left, right
	return w
}

// Title sets the text in the top border.
func (w Widget) Title(title string) Widget {
	w.title = title
	return w
}

// TitleStyle sets the style of the title.
func (w Widget) TitleStyle(style tcell.Style) Widget {
	w.titleStyle = style
	return w
}

// TitleAlignment sets the alignment of the title.
func (w Widget) TitleAlignment(alignment tview.Alignment) Widget {
	w.titleAlignment = alignment
	return w
}

// Footer sets the text in the bottom border.
func (w Widget) Footer(footer string) Widget {
	w.footer = footer
	return w
}

// FooterStyle sets the style of the footer.
func (w Widget) FooterStyle(style tcell.Style) Widget {
	w.footerStyle = style
	return w
}

// FooterAlignment sets the alignment of the footer.
func (w Widget) FooterAlignment(alignment tview.Alignment) Widget {
	w.footerAlignment = alignment
	return w
}

// Size returns the width and height of the box.
func (w Widget) Size() (width, height tview.Length) {
	return w.width, w.height
}

// InnerArea returns the part of area inside the border and padding, where the child is drawn.
func (w Widget) InnerArea(area tview.Rectangle) tview.Rectangle {
	if w.title != "" || w.borders.Has(tview.BordersTop) {
		area.Y++
		area.Height--
	}
	if w.footer != "" || w.borders.Has(tview.BordersBottom) {
		area.Height--
	}
	if w.borders.Has(tview.BordersLeft) {
		area.X++
		area.Width--
	}
	if w.borders.Has(tview.BordersRight) {
		area.Width--
	}
	area.X += w.paddingLeft
	area.Y += w.paddingTop
	area.Width = max(area.Width-w.paddingLeft-w.paddingRight, 0)
	area.Height = max(area.Height-w.paddingTop-w.paddingBottom, 0)
	return area
}

// Draw draws the background, border, title, and footer in area and the child inside them.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	if area.Width <= 0 || area.Height <= 0 {
		return
	}
	left, top := area.X, area.Y
	right, bottom := area.X+area.Width-1, area.Y+area.Height-1

	fill := func(x0, y0, x1, y1 int, str string, style tcell.Style) {
		for y := y0; y <= y1; y++ {
			for x := x0; x <= x1; x++ {
				screen.Put(x, y, str, style)
			}
		}
	}
	screen.FillArea(area.X, area.Y, area.Width, area.Height, ' ', tcell.StyleDefault.Background(w.background))

	if w.borders != tview.BordersNone && area.Width >= 2 && area.Height >= 2 {
		set := w.borderSet
		for _, part := range []struct {
			sides          tview.Borders
			x0, y0, x1, y1 int
			glyph          string
		}{
			{tview.BordersTop, left + 1, top, right - 1, top, set.Top},
			{tview.BordersBottom, left + 1, bottom, right - 1, bottom, set.Bottom},
			{tview.BordersLeft, left, top + 1, left, bottom - 1, set.Left},
			{tview.BordersRight, right, top + 1, right, bottom - 1, set.Right},
			{tview.BordersTop | tview.BordersLeft, left, top, left, top, set.TopLeft},
			{tview.BordersTop | tview.BordersRight, right, top, right, top, set.TopRight},
			{tview.BordersBottom | tview.BordersLeft, left, bottom, left, bottom, set.BottomLeft},
			{tview.BordersBottom | tview.BordersRight, right, bottom, right, bottom, set.BottomRight},
		} {
			if w.borders.Has(part.sides) {
				fill(part.x0, part.y0, part.x1, part.y1, part.glyph, w.borderStyle)
			}
		}
	}

	label(screen, w.title, area, top, w.titleAlignment, w.titleStyle)
	label(screen, w.footer, area, bottom, w.footerAlignment, w.footerStyle)
	w.child.Draw(screen, w.InnerArea(area))
}

// label draws text on row y inside the corners of area, ending it with an ellipsis if it is cut off.
func label(screen tview.Screen, text string, area tview.Rectangle, y int, alignment tview.Alignment, style tcell.Style) {
	if text == "" || area.Width < 4 {
		return
	}
	if tview.Print(screen, text, area.X+1, y, area.Width-2, alignment, style) < uniseg.StringWidth(text) {
		x := area.X + area.Width - 2
		if alignment == tview.AlignmentRight {
			x = area.X + 1
		}
		screen.Put(x, y, tview.SemigraphicsHorizontalEllipsis, style)
	}
}

// Handle passes msg to the child with the area inside the frame.
func (w Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	return w.child.Handle(msg, w.InnerArea(area))
}
