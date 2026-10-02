// Package tabs provides a row of tab labels above the content of the active tab.
package tabs

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/layout"
	"github.com/gdamore/tcell/v3"
	"github.com/rivo/uniseg"
)

// Widget draws the tab labels on the first row above the content of the active tab. Labels too wide for the row scroll to center the active one. The model owns the tabs and switches between them.
type Widget struct {
	labels                         []string
	active                         int
	content                        tview.Element
	style, activeStyle, arrowStyle tcell.Style
	alignment                      tview.Alignment
	separator                      string
	paddingLeft, paddingRight      string
	arrowStart, arrowEnd           string
	clickableArrows, wrap          bool
	keybind                        func(tview.KeyMsg) Action
	onSelect                       func(index int) tview.Msg
}

var _ tview.Element = Widget{}

// New returns centered tabs with labels separated by a space, the active one reversed, and no arrows.
func New(labels ...string) Widget {
	return Widget{
		labels:          labels,
		activeStyle:     tcell.StyleDefault.Reverse(true),
		arrowStyle:      tcell.StyleDefault.Dim(true),
		alignment:       tview.AlignmentCenter,
		separator:       " ",
		clickableArrows: true,
		keybind:         DefaultKeybind,
	}
}

// Active sets the index of the active tab.
func (w Widget) Active(index int) Widget {
	w.active = index
	return w
}

// Content sets the element of the active tab.
func (w Widget) Content(content tview.Element) Widget {
	w.content = content
	return w
}

// Style sets the style of the row and of the labels of inactive tabs.
func (w Widget) Style(style tcell.Style) Widget {
	w.style = style
	return w
}

// ActiveStyle sets the style of the label of the active tab.
func (w Widget) ActiveStyle(style tcell.Style) Widget {
	w.activeStyle = style
	return w
}

// Alignment sets where labels that fit are placed in the row.
func (w Widget) Alignment(alignment tview.Alignment) Widget {
	w.alignment = alignment
	return w
}

// Separator sets the text drawn between labels.
func (w Widget) Separator(separator string) Widget {
	w.separator = separator
	return w
}

// Padding sets the text drawn on either side of each label, in the style of the label.
func (w Widget) Padding(left, right string) Widget {
	w.paddingLeft, w.paddingRight = left, right
	return w
}

// Arrows sets the arrows drawn at the start and end of the row while labels are hidden past them, such as "◀" and "▶". An empty string draws no arrow at that end.
func (w Widget) Arrows(start, end string) Widget {
	w.arrowStart, w.arrowEnd = start, end
	return w
}

// ClickableArrows sets whether clicking an arrow selects the neighbor of the active tab, true unless set.
func (w Widget) ClickableArrows(clickable bool) Widget {
	w.clickableArrows = clickable
	return w
}

// ArrowStyle sets the style of the arrows.
func (w Widget) ArrowStyle(style tcell.Style) Widget {
	w.arrowStyle = style
	return w
}

// Wrap sets whether moving past the last tab selects the first and past the first selects the last.
func (w Widget) Wrap(wrap bool) Widget {
	w.wrap = wrap
	return w
}

// Keybind sets the function that turns keys into Actions, or ActionNone for keys it does not bind, DefaultKeybind unless set.
func (w Widget) Keybind(f func(tview.KeyMsg) Action) Widget {
	w.keybind = f
	return w
}

// OnSelect sets the function that turns the index of a tab to switch to into a message.
func (w Widget) OnSelect(onSelect func(index int) tview.Msg) Widget {
	w.onSelect = onSelect
	return w
}

// Size returns Fill, as the tabs takes its whole area.
func (Widget) Size() (width, height layout.Length) {
	return layout.Fill, layout.Fill
}

// Layout returns the size of limits, as the tabs takes its whole area.
func (Widget) Layout(limits layout.Limits) layout.Size {
	return layout.Atomic(limits, layout.Fill, layout.Fill)
}

// Draw draws the labels on the first row of area, highlighting the active one, and the content below them.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	l := w.layout(area)
	for x := area.X; x < area.X+area.Width; x++ {
		screen.Put(x, area.Y, " ", w.style)
	}
	for i, label := range w.labels {
		style := w.style
		if i == w.active {
			style = w.activeStyle
		}
		l.print(screen, w.paddingLeft+label+w.paddingRight, l.xs[i], l.widths[i], area.Y, style)
		if i < len(w.labels)-1 {
			l.print(screen, w.separator, l.xs[i+1]-l.gap, l.gap, area.Y, w.style)
		}
	}
	if l.startArrow {
		tview.Print(screen, w.arrowStart, area.X, area.Y, l.left-area.X, tview.AlignmentLeft, w.arrowStyle)
	}
	if l.endArrow {
		tview.Print(screen, w.arrowEnd, l.right, area.Y, area.X+area.Width-l.right, tview.AlignmentLeft, w.arrowStyle)
	}
	if w.content != nil {
		w.content.Draw(screen, contentArea(area))
	}
}

// Handle turns tab Actions, clicks on the labels, and scrolling over the row of labels into the OnSelect message, and passes other messages to the content below the labels.
func (w Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	switch msg := msg.(type) {
	case tview.KeyMsg:
		if action := w.keybind(msg); action != ActionNone {
			delta := 1
			if action == ActionPrevious {
				delta = -1
			}
			if tab, ok := w.step(delta); ok {
				return w.selectTab(tab)
			}
		}
	case tview.MouseMsg:
		x, y := msg.Position()
		if y != area.Y || !area.Contains(x, y) {
			break
		}
		switch msg.Action {
		case tview.MouseLeftClick:
			if tab, ok := w.tabAt(area, x); ok {
				return w.selectTab(tab)
			}
		case tview.MouseScrollUp, tview.MouseScrollLeft:
			if tab, ok := w.step(-1); ok {
				return w.selectTab(tab)
			}
		case tview.MouseScrollDown, tview.MouseScrollRight:
			if tab, ok := w.step(1); ok {
				return w.selectTab(tab)
			}
		}
		return nil
	}
	if w.content == nil {
		return msg
	}
	return w.content.Handle(msg, contentArea(area))
}

// step returns the tab delta tabs from the active one, wrapping around the ends when Wrap is set, and whether there is one.
func (w Widget) step(delta int) (int, bool) {
	next := w.active + delta
	if w.wrap && len(w.labels) > 0 {
		return (next + len(w.labels)) % len(w.labels), true
	}
	return next, next >= 0 && next < len(w.labels)
}

// selectTab returns the OnSelect message for index, or nil if index is already active.
func (w Widget) selectTab(index int) tview.Msg {
	if index == w.active || w.onSelect == nil {
		return nil
	}
	return w.onSelect(index)
}

// contentArea returns the area of the content, below the labels.
func contentArea(area tview.Rectangle) tview.Rectangle {
	area.Y++
	area.Height = max(area.Height-1, 0)
	return area
}

// layout is where the labels are drawn on the first row of an area.
type labelLayout struct {
	// xs and widths are where each padded label starts and how wide it is, with gap cells between labels for the separator.
	xs, widths []int
	gap        int
	// left and right bound the columns the labels are drawn in, the rest of the row is kept for the arrows.
	left, right int
	// startArrow and endArrow are set while labels are hidden past an arrow.
	startArrow, endArrow bool
}

// layout places the labels in area by the alignment, or when they do not fit, centers the active one without scrolling past the first or last label.
func (w Widget) layout(area tview.Rectangle) labelLayout {
	l := labelLayout{gap: uniseg.StringWidth(w.separator), left: area.X, right: area.X + area.Width}
	stripWidth := -l.gap // no separator after the last label
	for _, label := range w.labels {
		width := uniseg.StringWidth(w.paddingLeft + label + w.paddingRight)
		l.xs, l.widths = append(l.xs, stripWidth+l.gap), append(l.widths, width)
		stripWidth += width + l.gap
	}
	offset := l.left
	switch {
	case stripWidth > area.Width:
		l.left += uniseg.StringWidth(w.arrowStart)
		l.right -= uniseg.StringWidth(w.arrowEnd)
		width := l.right - l.left
		center := l.xs[w.active] + l.widths[w.active]/2
		offset = l.left - min(max(center-width/2, 0), stripWidth-width)
	case w.alignment == tview.AlignmentCenter:
		offset += (area.Width - stripWidth) / 2
	case w.alignment == tview.AlignmentRight:
		offset += area.Width - stripWidth
	}
	for i := range l.xs {
		l.xs[i] += offset
	}
	if last := len(l.xs) - 1; last >= 0 {
		l.startArrow = l.xs[0] < l.left
		l.endArrow = l.xs[last]+l.widths[last] > l.right
	}
	return l
}

// print draws text, width cells wide from x, on row y, cut to the columns the labels are drawn in.
func (l labelLayout) print(screen tview.Screen, text string, x, width, y int, style tcell.Style) {
	// Right alignment cuts the start of text that begins before the columns, center alignment both ends.
	end := x + width
	alignment := tview.AlignmentLeft
	switch {
	case x < l.left && end > l.right:
		alignment = tview.AlignmentCenter
	case x < l.left:
		alignment = tview.AlignmentRight
	}
	from, to := max(x, l.left), min(end, l.right)
	tview.Print(screen, text, from, y, to-from, alignment, style)
}

// tabAt returns the tab whose label is at column x of the first row of area, or the neighbor of the active tab past an arrow.
func (w Widget) tabAt(area tview.Rectangle, x int) (int, bool) {
	l := w.layout(area)
	switch {
	case x < l.left:
		return w.active - 1, l.startArrow && w.clickableArrows
	case x >= l.right:
		return w.active + 1, l.endArrow && w.clickableArrows
	}
	for i, start := range l.xs {
		if x >= start && x < start+l.widths[i] {
			return i, true
		}
	}
	return 0, false
}
