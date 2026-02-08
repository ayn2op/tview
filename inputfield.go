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
	bindDirtyParent(i.textArea, i.Box)
	i.textArea.SetChangedFunc(func() {
		if i.changed != nil {
			i.changed(i.textArea.GetText())
		}
	})
	i.textArea.textStyle = tcell.StyleDefault.Background(Styles.ContrastBackgroundColor).Foreground(Styles.PrimaryTextColor)
	i.textArea.placeholderStyle = tcell.StyleDefault.Background(Styles.ContrastBackgroundColor).Foreground(Styles.ContrastSecondaryTextColor)
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

// SetLabelWidth sets the screen width of the label. A value of 0 will cause the
// primitive to use the width of the label string.
func (i *InputField) SetLabelWidth(width int) *InputField {
	if i.textArea.GetLabelWidth() != width {
		i.textArea.SetLabelWidth(width)
	}
	return i
}

// SetPlaceholder sets the text to be displayed when the input text is empty.
func (i *InputField) SetPlaceholder(text string) *InputField {
	if i.textArea.placeholder != text {
		i.textArea.SetPlaceholder(text)
	}
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

// SetFieldBackgroundColor sets the background color of the input area.
func (i *InputField) SetFieldBackgroundColor(color tcell.Color) *InputField {
	style := i.textArea.GetTextStyle().Background(color)
	if i.textArea.GetTextStyle() != style {
		i.textArea.SetTextStyle(style)
	}
	return i
}

// SetFieldTextColor sets the text color of the input area.
func (i *InputField) SetFieldTextColor(color tcell.Color) *InputField {
	style := i.textArea.GetTextStyle().Foreground(color)
	if i.textArea.GetTextStyle() != style {
		i.textArea.SetTextStyle(style)
	}
	return i
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

// SetPlaceholderTextColor sets the text color of placeholder text.
func (i *InputField) SetPlaceholderTextColor(color tcell.Color) *InputField {
	style := i.textArea.GetPlaceholderStyle().Foreground(color)
	if i.textArea.GetPlaceholderStyle() != style {
		i.textArea.SetPlaceholderStyle(style)
	}
	return i
}

// SetPlaceholderStyle sets the style of the input area (when a placeholder is
// shown).
func (i *InputField) SetPlaceholderStyle(style tcell.Style) *InputField {
	if i.textArea.GetPlaceholderStyle() != style {
		i.textArea.SetPlaceholderStyle(style)
	}
	return i
}

// GetPlaceholderStyle returns the style of the input area (when a placeholder
// is shown).
func (i *InputField) GetPlaceholderStyle() tcell.Style {
	return i.textArea.GetPlaceholderStyle()
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
		i.MarkDirty()
	}
	return i
}

// GetFieldWidth returns this primitive's field width.
func (i *InputField) GetFieldWidth() int {
	return i.fieldWidth
}

// GetFieldHeight returns this primitive's field height.
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
		i.MarkDirty()
		return i
	}
	maskStr := string(mask)
	maskWidth := uniseg.StringWidth(maskStr)
	i.textArea.setTransform(func(cluster, rest string, boundaries int) (newCluster string, newBoundaries int) {
		return maskStr, maskWidth << uniseg.ShiftWidth
	})
	i.MarkDirty()
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

// Focus is called when this primitive receives focus.
func (i *InputField) Focus(delegate func(p Primitive)) {
	// If we're part of a form and this item is disabled, there's nothing the
	// user can do here so we're finished.
	if i.finished != nil && i.textArea.GetDisabled() {
		i.finished(-1)
		return
	}

	i.Box.Focus(delegate)
}

// HasFocus returns whether or not this primitive has focus.
func (i *InputField) HasFocus() bool {
	return i.textArea.HasFocus() || i.Box.HasFocus()
}

// Blur is called when this primitive loses focus.
func (i *InputField) Blur() {
	i.textArea.Blur()
	i.Box.Blur()
}

// IsDirty returns whether this primitive or one of its children needs redraw.
func (i *InputField) IsDirty() bool {
	return i.Box.IsDirty() || i.textArea.IsDirty()
}

// MarkClean marks this primitive and children as clean.
func (i *InputField) MarkClean() {
	i.Box.MarkClean()
	i.textArea.MarkClean()
}

// Draw draws this primitive onto the screen.
func (i *InputField) Draw(screen tcell.Screen) {
	i.DrawForSubclass(screen, i)

	// Prepare
	x, y, width, height := i.GetInnerRect()
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

// InputHandler returns the handler for this primitive.
func (i *InputField) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return i.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p Primitive)) {
		if i.textArea.GetDisabled() {
			return
		}

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
		case tcell.KeyCtrlV:
			i.textArea.InputHandler()(event, setFocus)
		default:
			// Forward other key events to the text area.
			i.textArea.InputHandler()(event, setFocus)
		}
	})
}

// MouseHandler returns the mouse handler for this primitive.
func (i *InputField) MouseHandler() func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
	return i.WrapMouseHandler(func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
		if i.textArea.GetDisabled() {
			return false, nil
		}

		// Is mouse event within the input field?
		x, y := event.Position()
		if !i.InRect(x, y) {
			return false, nil
		}

		// Forward mouse event to the text area.
		consumed, capture = i.textArea.MouseHandler()(action, event, setFocus)

		// Focus in any case.
		if action == MouseLeftDown && !consumed {
			setFocus(i)
			consumed = true
		}

		return
	})
}

// PasteHandler returns the handler for this primitive.
func (i *InputField) PasteHandler() func(pastedText string, setFocus func(p Primitive)) {
	return i.WrapPasteHandler(func(pastedText string, setFocus func(p Primitive)) {
		// Input field may be disabled.
		if i.textArea.GetDisabled() {
			return
		}

		// Forward the pasted text to the text area.
		i.textArea.PasteHandler()(pastedText, setFocus)
	})
}
