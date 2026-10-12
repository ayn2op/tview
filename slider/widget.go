// Package slider provides a declarative slider widget.
package slider

import (
	"math"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/layout"
	"github.com/gdamore/tcell/v3"
)

// The symbols a slider is drawn with: the rail before the handle, the handle, and the rail after it.
const (
	symbolFilled = "━"
	symbolHandle = "●"
	symbolEmpty  = "─"
)

// Number is what the value of a slider can be.
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64
}

// Widget is a value shown as a handle on a rail. It is a value built in View; the model owns the value and changes it in Update.
type Widget[T Number] struct {
	min, max, value T
	step, shiftStep T
	width           layout.Length
	style           tview.Style
	keybind         func(tview.KeyMsg) Action
	onChange        func(value T) tview.Msg
	rail            func(cell, handle int) (string, tview.Style)
	focused         bool
}

var _ tview.Widget = Widget[int]{}

// New returns a slider of value, kept within min and max, that fills the width of its parent. Changing the value returns the message of onChange.
func New[T Number](min, max, value T, onChange func(value T) tview.Msg) Widget[T] {
	return Widget[T]{min: min, max: max, value: T(clamp(float64(value), float64(min), float64(max))), step: 1, width: layout.Fill, keybind: DefaultKeybind, onChange: onChange}
}

// Step sets how much a key changes the value by and what a click rounds it to, 1 unless set.
func (w Widget[T]) Step(step T) Widget[T] {
	w.step = step
	return w
}

// ShiftStep sets the step used in place of Step while shift is held, which is Step unless set.
func (w Widget[T]) ShiftStep(step T) Widget[T] {
	w.shiftStep = step
	return w
}

// Width sets the width of the slider.
func (w Widget[T]) Width(width layout.Length) Widget[T] {
	w.width = width
	return w
}

// Style sets the style of the handle and of the rail before it. The rail after it is dim.
func (w Widget[T]) Style(style tview.Style) Widget[T] {
	w.style = style
	return w
}

// Rail sets what each cell is drawn as, given which of them the handle is on, in place of a line.
func (w Widget[T]) Rail(rail func(cell, handle int) (string, tview.Style)) Widget[T] {
	w.rail = rail
	return w
}

// Keybind sets the function that turns keys into Actions, or ActionNone for keys it does not bind, DefaultKeybind unless set.
func (w Widget[T]) Keybind(f func(tview.KeyMsg) Action) Widget[T] {
	w.keybind = f
	return w
}

// Focused sets whether the slider has the focus, so keys change the value.
func (w Widget[T]) Focused(focused bool) Widget[T] {
	w.focused = focused
	return w
}

// Size returns the width of the slider and a height of one row.
func (w Widget[T]) Size() (width, height layout.Length) { return w.width, layout.Fixed(1) }

// Layout returns the size of the slider within limits.
func (w Widget[T]) Layout(limits layout.Limits) layout.Size {
	return layout.Atomic(limits, w.width, layout.Fixed(1))
}

// Draw draws the rail along the first row of area, with the handle as far along it as the value is within the range.
func (w Widget[T]) Draw(screen tview.Screen, area tview.Rectangle) {
	var handle int
	if w.max > w.min {
		handle = int(float64(w.value-w.min) * float64(area.Width-1) / float64(w.max-w.min))
	}

	for cell := range area.Width {
		symbol, style := symbolEmpty, w.style.Dim(true)
		switch {
		case w.rail != nil:
			symbol, style = w.rail(cell, handle)
		case cell < handle:
			symbol, style = symbolFilled, w.style
		case cell == handle:
			symbol, style = symbolHandle, w.style
		}

		screen.Put(area.X+cell, area.Y, symbol, style)
	}
}

// Handle returns the message of onChange, if it is not nil, for a key that changes the value while focused, and for a left press or drag within area. Other messages pass through unchanged.
func (w Widget[T]) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg {
	if w.onChange == nil {
		return msg
	}

	// The value is worked out as a float64, for one of any type not to overflow.
	lo, hi, value, step := float64(w.min), float64(w.max), float64(w.value), float64(w.step)
	if m, ok := msg.(interface{ Modifiers() tcell.ModMask }); ok && w.shiftStep != 0 && m.Modifiers()&tcell.ModShift != 0 {
		step = float64(w.shiftStep)
	}

	switch m := msg.(type) {
	case tview.KeyMsg:
		if !w.focused {
			return msg
		}

		switch w.keybind(m) {
		case ActionDecrease:
			value -= step
		case ActionIncrease:
			value += step
		}
	case tview.MouseMsg:
		x, y := m.Position()
		pressed := m.Action == tview.MouseLeftDown || m.Action == tview.MouseMove && m.Buttons()&tcell.ButtonPrimary != 0
		if !pressed || !area.Contains(x, y) {
			return msg
		}

		// The value under the pointer, rounded to the nearest step.
		value = float64(x-area.X) / float64(max(area.Width-1, 1)) * (hi - lo)
		value = lo + math.Round(value/step)*step
	}

	if value := T(clamp(value, lo, hi)); value != w.value {
		return w.onChange(value)
	}

	return msg
}

func clamp(value, lo, hi float64) float64 {
	return min(max(value, lo), hi)
}
