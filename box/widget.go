// Package box draws a border with a title and footer around a widget.
package box

import (
	"slices"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/layout"
	"github.com/gdamore/tcell/v3"
)

// Widget draws a background, border, title, and footer, with its child inside them.
type Widget struct {
	child         tview.Widget
	width, height layout.Length

	background                      tcell.Color
	borders                         tview.Borders
	borderSet                       tview.BorderSet
	borderStyle                     tview.Style
	paddingTop, paddingBottom       int
	paddingLeft, paddingRight       int
	title, footer                   tview.Widget
	titleAlignment, footerAlignment tview.HorizontalAlignment
}

var _ tview.Widget = Widget{}

// New returns a box around child, without a border, that fills its parent.
func New(child tview.Widget) Widget {
	return Widget{
		child:           child,
		width:           layout.Fill,
		height:          layout.Fill,
		borderSet:       tview.BorderSetPlain(),
		titleAlignment:  tview.HorizontalAlignmentCenter,
		footerAlignment: tview.HorizontalAlignmentCenter,
	}
}

// Width sets the width of the box.
func (w Widget) Width(width layout.Length) Widget {
	w.width = width
	return w
}

// Height sets the height of the box.
func (w Widget) Height(height layout.Length) Widget {
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
func (w Widget) BorderStyle(style tview.Style) Widget {
	w.borderStyle = style
	return w
}

// Padding sets the empty cells between the border and the child.
func (w Widget) Padding(top, bottom, left, right int) Widget {
	w.paddingTop, w.paddingBottom, w.paddingLeft, w.paddingRight = top, bottom, left, right
	return w
}

// Title sets the widget drawn in the top border, one row tall and as wide as its content, such as text or tabs, or nil for none.
func (w Widget) Title(title tview.Widget) Widget {
	w.title = title
	return w
}

// TitleAlignment sets the alignment of the title.
func (w Widget) TitleAlignment(align tview.HorizontalAlignment) Widget {
	w.titleAlignment = align
	return w
}

// Footer sets the widget drawn in the bottom border, one row tall and as wide as its content, or nil for none.
func (w Widget) Footer(footer tview.Widget) Widget {
	w.footer = footer
	return w
}

// FooterAlignment sets the alignment of the footer.
func (w Widget) FooterAlignment(align tview.HorizontalAlignment) Widget {
	w.footerAlignment = align
	return w
}

// Size returns the width and height of the box.
func (w Widget) Size() (width, height layout.Length) {
	return w.width, w.height
}

// Layout returns the size of the box within limits: that of its child inside the frame, plus the frame.
func (w Widget) Layout(limits layout.Limits) layout.Size {
	limits = limits.Width(w.width).Height(w.height)
	inner := w.InnerArea(tview.Rectangle{Width: layout.Infinity, Height: layout.Infinity})
	frame := layout.Size{Width: layout.Infinity - inner.Width, Height: layout.Infinity - inner.Height}
	content := w.child.Layout(limits.Shrink(frame).Loose())
	size := limits.Shrink(frame).Resolve(w.width, w.height, content)
	return layout.Size{Width: size.Width + frame.Width, Height: size.Height + frame.Height}
}

// InnerArea returns the part of area inside the border and padding, where the child is drawn.
func (w Widget) InnerArea(area tview.Rectangle) tview.Rectangle {
	if w.title != nil || w.borders.Has(tview.BordersTop) {
		area.Y++
		area.Height--
	}
	if w.footer != nil || w.borders.Has(tview.BordersBottom) {
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

	// A cell keeps what it holds from the last frame unless it changes, which spares the terminal screen measuring it again, so cells are not filled only to be drawn over.
	labels := w.labels(area)
	fill := func(x0, y0, x1, y1 int, str string, style tview.Style) {
		for y := y0; y <= y1; y++ {
			for x := x0; x <= x1; x++ {
				if !slices.ContainsFunc(labels, func(l label) bool { return l.area.Contains(x, y) }) {
					screen.Put(x, y, str, style)
				}
			}
		}
	}

	inset := func(side tview.Borders) int {
		if w.borders.Has(side) {
			return 1
		}
		return 0
	}
	x0, x1 := left+inset(tview.BordersLeft), right-inset(tview.BordersRight)
	y0, y1 := top+inset(tview.BordersTop), bottom-inset(tview.BordersBottom)
	screen.FillArea(x0, y0, x1-x0+1, y1-y0+1, ' ', tcell.StyleDefault.Background(w.background))

	// A box one row tall or one column wide still draws its borders, so a top border alone makes a line, its title on it.
	if w.borders != tview.BordersNone {
		set := w.borderSet
		// A side reaches into a corner unless the side it meets there is drawn too, which makes the corner.
		for _, part := range []struct {
			sides          tview.Borders
			x0, y0, x1, y1 int
			glyph          string
		}{
			{tview.BordersTop, x0, top, x1, top, set.Top},
			{tview.BordersBottom, x0, bottom, x1, bottom, set.Bottom},
			{tview.BordersLeft, left, y0, left, y1, set.Left},
			{tview.BordersRight, right, y0, right, y1, set.Right},
			{tview.BordersTop | tview.BordersLeft, left, top, left, top, set.TopLeft},
			{tview.BordersTop | tview.BordersRight, right, top, right, top, set.TopRight},
			{tview.BordersBottom | tview.BordersLeft, left, bottom, left, bottom, set.BottomLeft},
			{tview.BordersBottom | tview.BordersRight, right, bottom, right, bottom, set.BottomRight},
		} {
			if w.borders&part.sides == part.sides {
				fill(part.x0, part.y0, part.x1, part.y1, part.glyph, w.borderStyle)
			}
		}
	}

	for _, l := range labels {
		l.widget.Draw(screen, l.area)
	}
	w.child.Draw(screen, w.InnerArea(area))
}

// label is a title or footer and where it is drawn in the border.
type label struct {
	widget tview.Widget
	area   tview.Rectangle
}

// labels returns where the title and footer are drawn on the top and bottom rows of area, inside the corners, as wide as their content and placed by their alignments.
func (w Widget) labels(area tview.Rectangle) []label {
	if area.Width < 4 {
		return nil
	}

	var labels []label
	for _, l := range []struct {
		widget tview.Widget
		y      int
		align  tview.HorizontalAlignment
	}{
		{w.title, area.Y, w.titleAlignment},
		{w.footer, area.Y + area.Height - 1, w.footerAlignment},
	} {
		if l.widget == nil {
			continue
		}

		room := area.Width - 2
		width := l.widget.Layout(layout.Limits{Max: layout.Size{Width: room, Height: 1}}).Width
		x := area.X + 1
		switch l.align {
		case tview.HorizontalAlignmentCenter:
			x += room/2 - width/2
		case tview.HorizontalAlignmentRight:
			x += room - width
		}

		labels = append(labels, label{widget: l.widget, area: tview.Rectangle{X: x, Y: l.y, Width: width, Height: 1}})
	}

	return labels
}

// Handle passes a mouse message on the title or footer to it, and other messages to the child with the area inside the frame.
func (w Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	if mouse, ok := msg.(tview.MouseMsg); ok {
		for _, l := range w.labels(area) {
			if l.area.Contains(mouse.Position()) {
				return l.widget.Handle(msg, l.area)
			}
		}
	}

	return w.child.Handle(msg, w.InnerArea(area))
}
