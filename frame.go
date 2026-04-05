package tview

import (
	"github.com/gdamore/tcell/v3"
)

// frameText holds information about a line of text shown in the frame.
type frameText struct {
	Text      string // The text to be displayed.
	Header    bool   // true = place in header, false = place in footer.
	Alignment Alignment
	Color     tcell.Color // The text color.
}

// Frame is a wrapper which adds space around another model. In addition,
// the top area (header) and the bottom area (footer) may also contain text.
//
// See https://github.com/ayn2op/tview/wiki/Frame for an example.
type Frame struct {
	*Box

	// The contained model. May be nil.
	primitive Model

	// The lines of text to be displayed.
	text []*frameText

	// Border spacing.
	top, bottom, header, footer, left, right int

	// Keep a reference in case we need it when we change the model.
	setFocus func(m Model)
}

// NewFrame returns a new frame around the given model. The model's
// size will be changed to fit within this frame. The model may be nil, in
// which case no other model is embedded in the frame.
func NewFrame(primitive Model) *Frame {
	box := NewBox()

	f := &Frame{
		Box:       box,
		primitive: primitive,
		top:       1,
		bottom:    1,
		header:    1,
		footer:    1,
		left:      1,
		right:     1,
	}

	return f
}

// SetPrimitive replaces the contained model with the given one. To remove
// a model, set it to nil.
func (f *Frame) SetPrimitive(m Model) *Frame {
	if f.primitive == m {
		return f
	}
	var hasFocus bool
	if f.primitive != nil {
		hasFocus = f.primitive.HasFocus()
	}
	f.primitive = m
	if hasFocus && f.setFocus != nil {
		f.setFocus(m) // Restore focus.
	}
	return f
}

// GetPrimitive returns the model contained in this frame.
func (f *Frame) GetPrimitive() Model {
	return f.primitive
}

// AddText adds text to the frame. Set "header" to true if the text is to appear
// in the header, above the contained model. Set it to false for it to
// appear in the footer, below the contained model. Rows in the header are printed top to bottom, rows in
// the footer are printed bottom to top. Note that long text can overlap as
// different alignments will be placed on the same row.
func (f *Frame) AddText(text string, header bool, alignment Alignment, color tcell.Color) *Frame {
	f.text = append(f.text, &frameText{
		Text:      text,
		Header:    header,
		Alignment: alignment,
		Color:     color,
	})
	return f
}

// Clear removes all text from the frame.
func (f *Frame) Clear() *Frame {
	if len(f.text) > 0 {
		f.text = nil
	}
	return f
}

// SetBorders sets the width of the frame borders as well as "header" and
// "footer", the vertical space between the header and footer text and the
// contained model (does not apply if there is no text).
func (f *Frame) SetBorders(top, bottom, header, footer, left, right int) *Frame {
	if f.top != top || f.bottom != bottom || f.header != header || f.footer != footer || f.left != left || f.right != right {
		f.top, f.bottom, f.header, f.footer, f.left, f.right = top, bottom, header, footer, left, right
	}
	return f
}

// Draw draws this model onto the screen.
func (f *Frame) Draw(screen tcell.Screen) {
	f.DrawForSubclass(screen, f)

	// Calculate start positions.
	x, top, width, height := f.InnerRect()
	bottom := top + height - 1
	x += f.left
	top += f.top
	bottom -= f.bottom
	width -= f.left + f.right
	if width <= 0 || top >= bottom {
		return // No space left.
	}

	// Draw text.
	var rows [6]int // top-left, top-center, top-right, bottom-left, bottom-center, bottom-right.
	topMax := top
	bottomMin := bottom
	for _, text := range f.text {
		// Where do we place this text?
		var y int
		if text.Header {
			y = top + rows[text.Alignment]
			rows[text.Alignment]++
			if y >= bottomMin {
				continue
			}
			if y+1 > topMax {
				topMax = y + 1
			}
		} else {
			y = bottom - rows[3+text.Alignment]
			rows[3+text.Alignment]++
			if y <= topMax {
				continue
			}
			if y-1 < bottomMin {
				bottomMin = y - 1
			}
		}

		// Draw text.
		Print(screen, text.Text, x, y, width, text.Alignment, text.Color)
	}

	// Set the size of the contained model.
	if f.primitive != nil {
		if topMax > top {
			top = topMax + f.header
		}
		if bottomMin < bottom {
			bottom = bottomMin - f.footer
		}
		if top > bottom {
			return // No space for the model.
		}
		f.primitive.SetRect(x, top, width, bottom+1-top)

		// Finally, draw the contained model.
		f.primitive.Draw(screen)
	}
}

// Focus is called when this model receives focus.
func (f *Frame) Focus(delegate func(m Model)) {
	f.setFocus = delegate
	if f.primitive != nil {
		delegate(f.primitive)
	} else {
		f.Box.Focus(delegate)
	}
}

// HasFocus returns whether or not this model has focus.
func (f *Frame) HasFocus() bool {
	if f.primitive == nil {
		return f.Box.HasFocus()
	}
	return f.primitive.HasFocus()
}

// Update handles input events for this model.
func (f *Frame) Update(msg Msg) Cmd {
	switch msg := msg.(type) {
	case *MouseMsg:
		if !f.InRect(msg.Position()) {
			return nil
		}

		// Pass mouse events on to contained model.
		if f.primitive != nil {
			childCmds := f.primitive.Update(msg)
			if childCmds != nil {
				return childCmds
			}
		}

		// Clicking on the frame parts.
		if msg.Action == MouseLeftDown {
			return SetFocus(f)
		}
	case *KeyMsg, *PasteMsg:
		if f.primitive == nil {
			return nil
		}
		return f.primitive.Update(msg)
	}
	return nil
}
