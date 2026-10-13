// Package flex lays out widgets along one axis. It implements the row and column packages.
package flex

import (
	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/layout"
)

// Widget lays out its children along one axis. Children that do not fill it are measured with their Layout.
type Widget struct {
	horizontal    bool
	width, height layout.Length
	spacing       int
	alignment     tview.Alignment
	children      []tview.Widget
}

var _ tview.Widget = Widget{}

// New returns children laid out left to right if horizontal and top to bottom otherwise, skipping nil ones. It fills its parent in both directions by default.
func New(horizontal bool, children ...tview.Widget) Widget {
	w := Widget{horizontal: horizontal, width: layout.Fill, height: layout.Fill}
	for _, child := range children {
		w = w.Push(child)
	}
	return w
}

// Push adds child after the others, unless it is nil.
func (w Widget) Push(child tview.Widget) Widget {
	if child != nil {
		w.children = append(w.children[:len(w.children):len(w.children)], child)
	}
	return w
}

// Width sets the width of the layout.
func (w Widget) Width(width layout.Length) Widget {
	w.width = width
	return w
}

// Height sets the height of the layout.
func (w Widget) Height(height layout.Length) Widget {
	w.height = height
	return w
}

// Spacing sets the number of cells between children.
func (w Widget) Spacing(spacing int) Widget {
	w.spacing = spacing
	return w
}

// Align sets where the children that do not fill the layout across its axis go across it: at its start unless set.
func (w Widget) Align(alignment tview.Alignment) Widget {
	w.alignment = alignment
	return w
}

// Size returns the width and height of the layout.
func (w Widget) Size() (width, height layout.Length) {
	return w.width, w.height
}

// Layout returns the size of the layout within limits, laying out its children where it shrinks to them.
func (w Widget) Layout(limits layout.Limits) layout.Size {
	return layout.Sized(limits, w.width, w.height, func(limits layout.Limits) layout.Size {
		_, main, cross := w.resolve(limits)
		width, height := pack(w.horizontal, main, cross)
		return layout.Size{Width: width, Height: height}
	})
}

// pack returns main and cross as a width and height, or a width and height as main and cross.
func pack[T any](horizontal bool, main, cross T) (T, T) {
	if horizontal {
		return main, cross
	}

	return cross, main
}

// resolve returns the size of each child within limits and the space they take along the axis and across it.
// Children that do not fill the axis are laid out first, and the others share what is left.
// Where the cross axis shrinks or scrolls, the children that fill it are laid out last, within what the others take across it.
func (w Widget) resolve(limits layout.Limits) (sizes []layout.Size, main, cross int) {
	mainCompress, crossCompress := pack(w.horizontal, limits.Compression.Width, limits.Compression.Height)
	mainInfinite, crossInfinite := pack(w.horizontal, limits.Infinite.Width, limits.Infinite.Height)
	crossDynamic := crossCompress || crossInfinite
	available, maxCross := pack(w.horizontal, limits.Max.Width, limits.Max.Height)
	spacing := w.spacing * max(len(w.children)-1, 0)
	available -= spacing

	sizes = make([]layout.Size, len(w.children))
	lay := func(i, minMain, maxMain, maxCross int, compress, crossInfinite bool) int {
		var child layout.Limits
		child.Min.Width, child.Min.Height = pack(w.horizontal, minMain, 0)
		child.Max.Width, child.Max.Height = pack(w.horizontal, max(maxMain, 0), maxCross)
		child.Compression.Width, child.Compression.Height = pack(w.horizontal, compress, false)
		child.Infinite.Width, child.Infinite.Height = pack(w.horizontal, mainInfinite, crossInfinite)
		sizes[i] = w.children[i].Layout(child)
		childMain, childCross := pack(w.horizontal, sizes[i].Width, sizes[i].Height)
		cross = max(cross, childCross)
		return childMain
	}

	const (
		static = iota
		crossFluid
		crossFluidDeferred
		mainFluid
	)
	categories, portions := make([]int, len(w.children)), 0
	for i, child := range w.children {
		length, across := w.axes(child)
		switch {
		case !mainCompress && length.Portion() > 0:
			categories[i] = mainFluid
			portions += length.Portion()
		case crossDynamic && across.Portion() > 0 && length.IsFixed():
			categories[i] = crossFluidDeferred
			available -= length.Cells()
		case crossDynamic && across.Portion() > 0:
			categories[i] = crossFluid
		default:
			available -= lay(i, 0, available, maxCross, mainCompress, crossInfinite)
		}
	}

	for i := range w.children {
		if categories[i] == crossFluid {
			available -= lay(i, 0, available, cross, mainCompress, false)
		}
	}

	available = max(available, 0)
	for i, child := range w.children {
		length, across := w.axes(child)
		if categories[i] != mainFluid {
			continue
		}

		// Portions are handed out from what is left so rounding never loses a cell.
		share := available * length.Portion() / portions
		available -= share
		portions -= length.Portion()
		if crossDynamic && across.Portion() > 0 {
			lay(i, share, share, cross, false, false)
		} else {
			lay(i, share, share, maxCross, false, crossInfinite)
		}
	}

	for i, child := range w.children {
		if length, _ := w.axes(child); categories[i] == crossFluidDeferred {
			lay(i, 0, length.Cells(), cross, false, false)
		}
	}

	for _, size := range sizes {
		childMain, _ := pack(w.horizontal, size.Width, size.Height)
		main += childMain
	}

	return sizes, main + spacing, cross
}

// axes returns the length of child along the layout's axis and across it.
func (w Widget) axes(child tview.Widget) (along, across layout.Length) {
	width, height := child.Size()
	return pack(w.horizontal, width, height)
}

func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	for i, childArea := range w.areas(area) {
		w.children[i].Draw(screen, childArea)
	}
}

// Handle passes msg through each child in turn, with the child's area. It stops when a child drops the message.
func (w Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	for i, childArea := range w.areas(area) {
		if msg = w.children[i].Handle(msg, childArea); msg == nil {
			return nil
		}
	}
	return msg
}

// areas returns the area of each child within area, one after the other from its start.
func (w Widget) areas(area tview.Rectangle) []tview.Rectangle {
	sizes, _, _ := w.resolve(layout.Limits{Max: layout.Size{Width: area.Width, Height: area.Height}})
	_, across := pack(w.horizontal, area.Width, area.Height)
	areas := make([]tview.Rectangle, len(sizes))
	pos := 0
	for i, size := range sizes {
		main, cross := pack(w.horizontal, size.Width, size.Height)
		offset := 0
		switch w.alignment {
		case tview.AlignmentCenter:
			offset = (across - cross) / 2
		case tview.AlignmentRight:
			offset = across - cross
		}

		x, y := pack(w.horizontal, pos, offset)
		areas[i] = tview.Rectangle{X: area.X + x, Y: area.Y + y, Width: size.Width, Height: size.Height}
		pos += main + w.spacing
	}

	return areas
}
