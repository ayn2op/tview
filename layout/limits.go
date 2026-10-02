package layout

// Size is a width and height in cells.
type Size struct {
	Width, Height int
}

// Axes holds a flag for each axis.
type Axes struct {
	Width, Height bool
}

// Infinity is the bound of Limits along an axis that is infinite.
const Infinity = 1 << 30

// Limits is a set of size constraints for laying out an element.
type Limits struct {
	// Min and Max are the smallest and largest size the element may take.
	Min, Max Size
	// Compression is whether Fill lengths are compressed to the size of the content.
	Compression Axes
	// Infinite is whether the size of the content may exceed Max, as along the axis a parent scrolls. Max is then a hint.
	Infinite Axes
}

// Bounds returns Max, with Infinity along the infinite axes.
func (l Limits) Bounds() Size {
	bounds := l.Max
	if l.Infinite.Width {
		bounds.Width = Infinity
	}
	if l.Infinite.Height {
		bounds.Height = Infinity
	}
	return bounds
}

// Width applies a width to the limits: Shrink compresses and Fixed sets both bounds.
func (l Limits) Width(width Length) Limits {
	l.Min.Width, l.Max.Width, l.Compression.Width, l.Infinite.Width = constrain(l.Min.Width, l.Max.Width, l.Compression.Width, l.Infinite.Width, width)
	return l
}

// Height applies a height to the limits: Shrink compresses and Fixed sets both bounds.
func (l Limits) Height(height Length) Limits {
	l.Min.Height, l.Max.Height, l.Compression.Height, l.Infinite.Height = constrain(l.Min.Height, l.Max.Height, l.Compression.Height, l.Infinite.Height, height)
	return l
}

func constrain(lower, upper int, compression, infinite bool, length Length) (int, int, bool, bool) {
	switch {
	case length.IsShrink():
		compression = true
	case length.IsFixed():
		cells := length.Cells()
		if !infinite {
			cells = min(cells, upper)
		}
		cells = max(cells, lower)
		lower, upper, compression, infinite = cells, cells, false, false
	}
	return lower, upper, compression, infinite
}

// Shrink returns the limits with size taken off both bounds.
func (l Limits) Shrink(size Size) Limits {
	l.Min = Size{Width: max(l.Min.Width-size.Width, 0), Height: max(l.Min.Height-size.Height, 0)}
	l.Max = Size{Width: max(l.Max.Width-size.Width, 0), Height: max(l.Max.Height-size.Height, 0)}
	return l
}

// Loose returns the limits without a minimum size.
func (l Limits) Loose() Limits {
	l.Min = Size{}
	return l
}

// Resolve returns the size that fits the limits for a width and height and the intrinsic size of the content.
func (l Limits) Resolve(width, height Length, intrinsic Size) Size {
	return Size{Width: l.ResolveWidth(width, intrinsic.Width), Height: l.ResolveHeight(height, intrinsic.Height)}
}

// ResolveWidth resolves only the width.
func (l Limits) ResolveWidth(width Length, intrinsic int) int {
	return resolve(l.Min.Width, l.Max.Width, l.Compression.Width, l.Infinite.Width, width, intrinsic)
}

// ResolveHeight resolves only the height.
func (l Limits) ResolveHeight(height Length, intrinsic int) int {
	return resolve(l.Min.Height, l.Max.Height, l.Compression.Height, l.Infinite.Height, height, intrinsic)
}

func resolve(lower, upper int, compression, infinite bool, length Length, intrinsic int) int {
	switch {
	case length.Portion() > 0 && !compression:
		if infinite {
			upper = max(upper, intrinsic)
		}
		return max(upper, lower)
	case length.IsFixed():
		cells := length.Cells()
		if !infinite {
			cells = min(cells, upper)
		}
		return max(cells, lower)
	default:
		if !infinite {
			intrinsic = min(intrinsic, upper)
		}
		return max(intrinsic, lower)
	}
}

// Atomic returns the size that fits limits for an element of a width and height with no content of its own.
func Atomic(limits Limits, width, height Length) Size {
	return limits.Width(width).Height(height).Resolve(width, height, Size{})
}

// Sized returns the size that fits limits for an element of a width and height whose content has the intrinsic size that content returns within the limits. content is only called if its result can change the size.
func Sized(limits Limits, width, height Length, content func(Limits) Size) Size {
	limits = limits.Width(width).Height(height)
	var intrinsic Size
	if c, i := limits.Compression, limits.Infinite; !width.IsFixed() && (c.Width || i.Width) || !height.IsFixed() && (c.Height || i.Height) {
		intrinsic = content(limits)
	}
	return limits.Resolve(width, height, intrinsic)
}
