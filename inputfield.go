package tview

import (
	"github.com/gdamore/tcell/v3"
	"github.com/rivo/uniseg"
)

// InputField is a one-line box into which the user can enter text. Use
// [InputField.SetAcceptanceFunc] to accept or reject input,
// [InputField.SetChangedFunc] to listen for changes, and
// [InputField.SetMaskCharacter] to hide input from onlookers (e.g. for password
// input).
//
// Navigation and editing is the same as for a [TextArea], with the following
// exceptions:
//
//   - Tab, BackTab, Enter, Escape: Finish editing.
//
// Note that while pressing Tab or Enter is intercepted by the input field, it
// is possible to paste such characters into the input field, possibly resulting
// in multi-line input. You can use [InputField.SetAcceptanceFunc] to prevent
// this.
//
// See https://github.com/ayn2op/tview/wiki/InputField for an example.
type InputField struct {
	*Box

	// The text area providing the core functionality of the input field.
	textArea *TextArea

	// The screen width of the input area. A value of 0 means extend as much as
	// possible.
	fieldWidth int

	// An optional function which is called when the input has changed.
	changed func(text string)

	// An optional function which is called when the user indicated that they
	// are done entering text. The key which was pressed is provided (tab,
	// shift-tab, enter, or escape).
	done func(tcell.Key)

	// A callback function set by the Form class and called when the user leaves
	// this form item.
	finished func(tcell.Key)
}

// NewInputField returns a new input field.
func NewInputField() *InputField {
	i := &InputField{
		Box:      NewBox(),
		textArea: NewTextArea().SetWrap(false),
	}
	i.textArea.SetChangedFunc(func() {
		if i.changed != nil {
			i.changed(i.textArea.GetText())
		}
	})
	i.textArea.textStyle = tcell.StyleDefault.Background(Styles.ContrastBackgroundColor).Foreground(Styles.PrimaryTextColor)
	return i
}

// SetText sets the current text of the input field. This can be undone by the
// user. Calling this function will also trigger a "changed" event.
func (i *InputField) SetText(text string) *InputField {
	if i.textArea.GetText() != text {
		i.textArea.Replace(0, i.textArea.GetTextLength(), text)
	}
	return i
}

// GetText returns the current text of the input field.
func (i *InputField) GetText() string {
	return i.textArea.GetText()
}

// SetLabel sets the text to be displayed before the input area.
func (i *InputField) SetLabel(label string) *InputField {
	if i.textArea.GetLabel() != label {
		i.textArea.SetLabel(label)
	}
	return i
}

// GetLabel returns the text to be displayed before the input area.
func (i *InputField) GetLabel() string {
	return i.textArea.GetLabel()
}

// SetLabelWidth sets the screen width of the label.
// A value of 0 represents the width of the label string.
func (i *InputField) SetLabelWidth(width int) *InputField {
	if i.textArea.GetLabelWidth() != width {
		i.textArea.SetLabelWidth(width)
	}
	return i
}

// SetPlaceholder sets the styled text to be displayed when the input text is
// empty.
func (i *InputField) SetPlaceholder(line Line) *InputField {
	i.textArea.SetPlaceholder(line)
	return i
}

// SetLabelColor sets the text color of the label.
func (i *InputField) SetLabelColor(color tcell.Color) *InputField {
	style := i.textArea.GetLabelStyle().Foreground(color)
	if i.textArea.GetLabelStyle() != style {
		i.textArea.SetLabelStyle(style)
	}
	return i
}

// SetLabelStyle sets the style of the label.
func (i *InputField) SetLabelStyle(style tcell.Style) *InputField {
	if i.textArea.GetLabelStyle() != style {
		i.textArea.SetLabelStyle(style)
	}
	return i
}

// GetLabelStyle returns the style of the label.
func (i *InputField) GetLabelStyle() tcell.Style {
	return i.textArea.GetLabelStyle()
}

// SetFieldStyle sets the style of the input area (when no placeholder is
// shown).
func (i *InputField) SetFieldStyle(style tcell.Style) *InputField {
	if i.textArea.GetTextStyle() != style {
		i.textArea.SetTextStyle(style)
	}
	return i
}

// GetFieldStyle returns the style of the input area (when no placeholder is
// shown).
func (i *InputField) GetFieldStyle() tcell.Style {
	return i.textArea.GetTextStyle()
}

// SetFormAttributes sets attributes shared by all form items.
func (i *InputField) SetFormAttributes(labelWidth int, labelColor, bgColor, fieldTextColor, fieldBgColor tcell.Color) FormItem {
	i.textArea.SetFormAttributes(labelWidth, labelColor, bgColor, fieldTextColor, fieldBgColor)
	return i
}

// SetFieldWidth sets the screen width of the input area. A value of 0 means
// extend as much as possible.
func (i *InputField) SetFieldWidth(width int) *InputField {
	if i.fieldWidth != width {
		i.fieldWidth = width
	}
	return i
}

// GetFieldWidth returns this model's field width.
func (i *InputField) GetFieldWidth() int {
	return i.fieldWidth
}

// GetFieldHeight returns this model's field height.
func (i *InputField) GetFieldHeight() int {
	return 1
}

// SetDisabled sets whether or not the item is disabled / read-only.
func (i *InputField) SetDisabled(disabled bool) FormItem {
	if i.textArea.GetDisabled() != disabled {
		i.textArea.SetDisabled(disabled)
	}
	if i.finished != nil {
		i.finished(-1)
	}
	return i
}

// GetDisabled returns whether or not the item is disabled / read-only.
func (i *InputField) GetDisabled() bool {
	return i.textArea.GetDisabled()
}

// SetMaskCharacter sets a character that masks user input on a screen. A value
// of 0 disables masking.
func (i *InputField) SetMaskCharacter(mask rune) *InputField {
	if mask == 0 {
		i.textArea.setTransform(nil)
		return i
	}
	maskStr := string(mask)
	maskWidth := uniseg.StringWidth(maskStr)
	i.textArea.setTransform(func(cluster, rest string, boundaries int) (newCluster string, newBoundaries int) {
		return maskStr, maskWidth << uniseg.ShiftWidth
	})
	return i
}

// SetChangedFunc sets a handler which is called whenever the text of the input
// field has changed. It receives the current text (after the change).
func (i *InputField) SetChangedFunc(handler func(text string)) *InputField {
	i.changed = handler
	return i
}

// SetDoneFunc sets a handler which is called when the user is done entering
// text. The callback function is provided with the key that was pressed, which
// is one of the following:
//
//   - KeyEnter: Done entering text.
//   - KeyEscape: Abort text input.
//   - KeyTab: Move to the next field.
//   - KeyBacktab: Move to the previous field.
func (i *InputField) SetDoneFunc(handler func(key tcell.Key)) *InputField {
	i.done = handler
	return i
}

// SetFinishedFunc sets a callback invoked when the user leaves this form item.
func (i *InputField) SetFinishedFunc(handler func(key tcell.Key)) FormItem {
	i.finished = handler
	return i
}

// Focus is called when this model receives focus.
func (i *InputField) Focus(delegate func(m Model)) {
	// If we're part of a form and this item is disabled, there's nothing the
	// user can do here so we're finished.
	if i.finished != nil && i.textArea.GetDisabled() {
		i.finished(-1)
		return
	}

	i.Box.Focus(delegate)
}

// HasFocus returns whether or not this model has focus.
func (i *InputField) HasFocus() bool {
	return i.textArea.HasFocus() || i.Box.HasFocus()
}

// Blur is called when this model loses focus.
func (i *InputField) Blur() {
	i.textArea.Blur()
	i.Box.Blur()
}

// Draw draws this model onto the screen.
func (i *InputField) Draw(screen tcell.Screen) {
	i.DrawForSubclass(screen, i)

	// Prepare
	x, y, width, height := i.InnerRect()
	if height < 1 || width < 1 {
		return
	}

	// Resize text area.
	labelWidth := i.textArea.GetLabelWidth()
	if labelWidth == 0 {
		labelWidth = TaggedStringWidth(i.textArea.GetLabel())
	}
	fieldWidth := i.fieldWidth
	if fieldWidth == 0 {
		fieldWidth = width - labelWidth
	}
	i.textArea.SetRect(x, y, labelWidth+fieldWidth, 1)
	i.textArea.setMinCursorPadding(fieldWidth-1, 1)

	// Draw text area.
	i.textArea.hasFocus = i.HasFocus() // Force cursor positioning.
	i.textArea.Draw(screen)
}

// HandleEvent handles input events for this model.
func (i *InputField) HandleEvent(event Event) Command {
	if i.textArea.GetDisabled() {
		return nil
	}

	switch event := event.(type) {
	case *KeyEvent:
		// Finish up.
		finish := func(key tcell.Key) {
			if i.done != nil {
				i.done(key)
			}
			if i.finished != nil {
				i.finished(key)
			}
		}

		// Process special key events for the input field.
		switch key := event.Key(); key {
		case tcell.KeyEnter, tcell.KeyEscape, tcell.KeyTab, tcell.KeyBacktab:
			finish(key)
			return nil
		default:
			// Forward other key events to the text area.
			return i.textArea.HandleEvent(event)
		}
	case *MouseEvent:
		// Is mouse event within the input field?
		x, y := event.Position()
		if !i.InRect(x, y) {
			return nil
		}

		// Forward mouse event to the text area.
		cmd := i.textArea.HandleEvent(event)

		// Focus in any case.
		if event.Action == MouseLeftDown && cmd == nil {
			cmd = SetFocus(i)
		}
		return cmd
	case *PasteEvent:
		// Forward the pasted text to the text area.
		return i.textArea.HandleEvent(event)
	}
	return nil
}
