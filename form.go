package tview

import (
	"slices"

	"github.com/gdamore/tcell/v3"
)

var (
	// DefaultFormFieldWidth is the default field screen width of form elements
	// whose field width is flexible (0). This is used in the Form class for
	// horizontal layouts.
	DefaultFormFieldWidth = 10

	// DefaultFormFieldHeight is the default field height of multi-line form
	// elements whose field height is flexible (0).
	DefaultFormFieldHeight = 5
)

// FormItem is the interface all form items must implement to be able to be
// included in a form.
type FormItem interface {
	Primitive

	// GetLabel returns the item's label text.
	GetLabel() string

	// SetFormAttributes sets a number of item attributes at once.
	SetFormAttributes(labelWidth int, labelColor, bgColor, fieldTextColor, fieldBgColor tcell.Color) FormItem

	// GetFieldWidth returns the width of the form item's field (the area which
	// is manipulated by the user) in number of screen cells. A value of 0
	// indicates the field width is flexible and may use as much space as
	// required.
	GetFieldWidth() int

	// GetFieldHeight returns the height of the form item's field (the area which
	// is manipulated by the user). This value must be greater than 0.
	GetFieldHeight() int

	// SetFinishedFunc sets the handler function for when the user finished
	// entering data into the item. The handler may receive events for the
	// Enter key (we're done), the Escape key (cancel input), the Tab key (move
	// to next field), the Backtab key (move to previous field), or a negative
	// value, indicating that the action for the last known key should be
	// repeated.
	SetFinishedFunc(handler func(key tcell.Key)) FormItem

	// SetDisabled sets whether or not the item is disabled / read-only. A form
	// must have at least one item that is not disabled.
	SetDisabled(disabled bool) FormItem

	// GetDisabled returns whether or not the item is disabled / read-only.
	GetDisabled() bool
}

// Form allows you to combine multiple one-line form elements into a vertical
// or horizontal layout. Form elements include types such as InputField or
// Checkbox. These elements can be optionally followed by one or more buttons
// for which you can define form-wide actions (e.g. Save, Clear, Cancel).
//
// See https://github.com/ayn2op/tview/wiki/Form for an example.
type Form struct {
	*Box

	// The items of the form (one row per item).
	items []FormItem

	// The buttons of the form.
	buttons []*Button

	// If set to true, instead of position items and buttons from top to bottom,
	// they are positioned from left to right.
	horizontal bool

	// The alignment of the buttons.
	buttonsAlignment Alignment

	// The number of empty cells between items.
	itemPadding int

	// The label color.
	labelColor tcell.Color

	// The style of the input area.
	fieldStyle tcell.Style

	// The style of the buttons when they are not focused.
	buttonStyle tcell.Style

	// The style of the buttons when they are focused.
	buttonActivatedStyle tcell.Style

	// The style of the buttons when they are disabled.
	buttonDisabledStyle tcell.Style

	// The index of the item or button for which the user requested focus.
	// Applied the next time the form itself receives focus. Negative if no
	// specific item was requested.
	requestedFocus int

	// A function to set the application's current focus. Does nothing
	// initially.
	setFocus func(Primitive)

	// The last (valid) key that wsa sent to a "finished" handler or -1 if no
	// such key is known yet.
	lastFinishedKey tcell.Key

	// An optional function which is called when the user hits Escape.
	cancel func()
}

// NewForm returns a new form.
func NewForm() *Form {
	box := NewBox().SetBorderPadding(1, 1, 1, 1)

	f := &Form{
		Box:                  box,
		itemPadding:          1,
		labelColor:           Styles.SecondaryTextColor,
		fieldStyle:           tcell.StyleDefault.Background(Styles.ContrastBackgroundColor).Foreground(Styles.PrimaryTextColor),
		buttonStyle:          tcell.StyleDefault.Background(Styles.ContrastBackgroundColor).Foreground(Styles.PrimaryTextColor),
		buttonActivatedStyle: tcell.StyleDefault.Reverse(true),
		buttonDisabledStyle:  tcell.StyleDefault.Background(Styles.ContrastBackgroundColor).Foreground(Styles.ContrastSecondaryTextColor),
		requestedFocus:       -1,
		setFocus:             func(Primitive) {},
		lastFinishedKey:      tcell.KeyTab, // To skip over inactive elements at the beginning of the form.
	}

	return f
}

// SetItemPadding sets the number of empty rows between form items for vertical
// layouts and the number of empty cells between form items for horizontal
// layouts. In vertical layouts, there is always at least one empty line between
// the last item and the buttons, if any.
func (f *Form) SetItemPadding(padding int) *Form {
	if f.itemPadding != padding {
		f.itemPadding = padding
	}
	return f
}

// SetHorizontal sets the direction the form elements are laid out. If set to
// true, instead of positioning them from top to bottom (the default), they are
// positioned from left to right, moving into the next row if there is not
// enough space.
func (f *Form) SetHorizontal(horizontal bool) *Form {
	if f.horizontal != horizontal {
		f.horizontal = horizontal
	}
	return f
}

// SetLabelColor sets the color of the labels.
func (f *Form) SetLabelColor(color tcell.Color) *Form {
	if f.labelColor != color {
		f.labelColor = color
	}
	return f
}

// SetFieldStyle sets the style of the input areas.
func (f *Form) SetFieldStyle(style tcell.Style) *Form {
	if f.fieldStyle != style {
		f.fieldStyle = style
	}
	return f
}

// SetButtonsAlignment sets how the buttons align horizontally.
func (f *Form) SetButtonsAlignment(alignment Alignment) *Form {
	if f.buttonsAlignment != alignment {
		f.buttonsAlignment = alignment
	}
	return f
}

// SetButtonStyle sets the style of the buttons when they are not focused.
func (f *Form) SetButtonStyle(style tcell.Style) *Form {
	if f.buttonStyle != style {
		f.buttonStyle = style
	}
	return f
}

// SetButtonActivatedStyle sets the style of the buttons when they are focused.
func (f *Form) SetButtonActivatedStyle(style tcell.Style) *Form {
	if f.buttonActivatedStyle != style {
		f.buttonActivatedStyle = style
	}
	return f
}

// SetButtonDisabledStyle sets the style of the buttons when they are disabled.
func (f *Form) SetButtonDisabledStyle(style tcell.Style) *Form {
	if f.buttonDisabledStyle != style {
		f.buttonDisabledStyle = style
	}
	return f
}

// SetFocus shifts the focus to the form element with the given index, counting
// non-button items first and buttons last. This does not change the
// application's focus immediately, but the next time the form itself receives
// focus, the given element will be focused once. Set to a negative value to
// focus the first (enabled) element.
func (f *Form) SetFocus(index int) *Form {
	if f.requestedFocus != index {
		f.requestedFocus = index
	}
	return f
}

// AddTextArea adds a text area to the form. It has a label, an optional initial
// text, a size (width and height) referring to the actual input area (a
// fieldWidth of 0 extends it as far right as possible, a fieldHeight of 0 will
// cause it to be [DefaultFormFieldHeight]), and a maximum number of bytes of
// text allowed (0 means no limit).
//
// The optional callback function is invoked when the content of the text area
// has changed. Note that especially for larger texts, this is an expensive
// operation due to technical constraints of the [TextArea] primitive (every key
// stroke leads to a new reallocation of the entire text).
func (f *Form) AddTextArea(label, text string, fieldWidth, fieldHeight, maxLength int, changed func(text string)) *Form {
	if fieldHeight == 0 {
		fieldHeight = DefaultFormFieldHeight
	}
	textArea := NewTextArea().
		SetLabel(label).
		SetSize(fieldHeight, fieldWidth).
		SetMaxLength(maxLength)
	if text != "" {
		textArea.SetText(text, true)
	}
	if changed != nil {
		textArea.SetChangedFunc(func() {
			changed(textArea.GetText())
		})
	}
	textArea.SetFinishedFunc(f.finished)
	f.items = append(f.items, textArea)
	return f
}

// AddTextView adds a text view to the form. It has a label and text, a size
// (width and height) referring to the actual text element (a fieldWidth of 0
// extends it as far right as possible, a fieldHeight of 0 will cause it to be
// [DefaultFormFieldHeight]), and a flag to turn on/off scrolling. If scrolling
// is turned off, the text view will not receive focus.
func (f *Form) AddTextView(label, text string, fieldWidth, fieldHeight int, scrollable bool) *Form {
	if fieldHeight == 0 {
		fieldHeight = DefaultFormFieldHeight
	}
	textArea := NewTextView().
		SetLabel(label).
		SetSize(fieldHeight, fieldWidth).
		SetScrollable(scrollable).
		SetText(text)
	textArea.SetFinishedFunc(f.finished)
	f.items = append(f.items, textArea)
	return f
}

// AddInputField adds an input field to the form. It has a label, an optional
// initial value, a field width (a value of 0 extends it as far as possible),
// and an (optional) callback function which is invoked when the input field's
// text has changed.
func (f *Form) AddInputField(label, value string, fieldWidth int, changed func(text string)) *Form {
	inputField := NewInputField().
		SetLabel(label).
		SetText(value).
		SetFieldWidth(fieldWidth).
		SetChangedFunc(changed)
	inputField.SetFinishedFunc(f.finished)
	f.items = append(f.items, inputField)
	return f
}

// AddPasswordField adds a password field to the form. This is similar to an
// input field except that the user's input not shown. Instead, a "mask"
// character is displayed. The password field has a label, an optional initial
// value, a field width (a value of 0 extends it as far as possible), and an
// (optional) callback function which is invoked when the input field's text has
// changed.
func (f *Form) AddPasswordField(label, value string, fieldWidth int, mask rune, changed func(text string)) *Form {
	if mask == 0 {
		mask = '*'
	}
	password := NewInputField().
		SetLabel(label).
		SetText(value).
		SetFieldWidth(fieldWidth).
		SetMaskCharacter(mask).
		SetChangedFunc(changed)
	password.SetFinishedFunc(f.finished)
	f.items = append(f.items, password)
	return f
}

// AddCheckbox adds a checkbox to the form. It has a label, an initial state,
// and an (optional) callback function which is invoked when the state of the
// checkbox was changed by the user.
func (f *Form) AddCheckbox(label string, checked bool, changed func(checked bool)) *Form {
	checkbox := NewCheckbox().
		SetLabel(label).
		SetChecked(checked).
		SetChangedFunc(changed)
	checkbox.SetFinishedFunc(f.finished)
	f.items = append(f.items, checkbox)
	return f
}

// AddButton adds a new button to the form. The "selected" function is called
// when the user selects this button. It may be nil.
func (f *Form) AddButton(label string, selected func()) *Form {
	button := NewButton(label).
		SetSelectedFunc(selected).
		SetExitFunc(f.finished)
	f.buttons = append(f.buttons, button)
	return f
}

// GetButton returns the button at the specified 0-based index. Note that
// buttons have been specially prepared for this form and modifying some of
// their attributes may have unintended side effects.
func (f *Form) GetButton(index int) *Button {
	return f.buttons[index]
}

// RemoveButton removes the button at the specified position, starting with 0
// for the button that was added first.
func (f *Form) RemoveButton(index int) *Form {
	f.buttons = slices.Delete(f.buttons, index, index+1)
	return f
}

// GetButtonCount returns the number of buttons in this form.
func (f *Form) GetButtonCount() int {
	return len(f.buttons)
}

// GetButtonIndex returns the index of the button with the given label, starting
// with 0 for the button that was added first. If no such label was found, -1
// is returned.
func (f *Form) GetButtonIndex(label string) int {
	for index, button := range f.buttons {
		if button.GetLabel() == label {
			return index
		}
	}
	return -1
}

// Clear removes all input elements from the form, including the buttons if
// specified.
func (f *Form) Clear(includeButtons bool) *Form {
	f.items = nil
	if includeButtons {
		f.ClearButtons()
	}
	return f
}

// ClearButtons removes all buttons from the form.
func (f *Form) ClearButtons() *Form {
	if len(f.buttons) > 0 {
		f.buttons = nil
	}
	return f
}

// AddFormItem adds a new item to the form. This can be used to add your own
// objects to the form. Note, however, that the Form class will override some
// of its attributes to make it work in the form context. Specifically, these
// are:
//
//   - The label width
//   - The label color
//   - The background color
//   - The field text color
//   - The field background color
func (f *Form) AddFormItem(item FormItem) *Form {
	item.SetFinishedFunc(f.finished)
	f.items = append(f.items, item)
	return f
}

// GetFormItemCount returns the number of items in the form (not including the
// buttons).
func (f *Form) GetFormItemCount() int {
	return len(f.items)
}

// GetFormItem returns the form item at the given position, starting with index
// 0. Elements are referenced in the order they were added. Buttons are not
// included.
func (f *Form) GetFormItem(index int) FormItem {
	return f.items[index]
}

// RemoveFormItem removes the form element at the given position, starting with
// index 0. Elements are referenced in the order they were added. Buttons are
// not included.
func (f *Form) RemoveFormItem(index int) *Form {
	f.items = slices.Delete(f.items, index, index+1)
	return f
}

// GetFormItemByLabel returns the first form element with the given label. If
// no such element is found, nil is returned. Buttons are not searched and will
// therefore not be returned.
func (f *Form) GetFormItemByLabel(label string) FormItem {
	for _, item := range f.items {
		if item.GetLabel() == label {
			return item
		}
	}
	return nil
}

// GetFormItemIndex returns the index of the first form element with the given
// label. If no such element is found, -1 is returned. Buttons are not searched
// and will therefore not be returned.
func (f *Form) GetFormItemIndex(label string) int {
	for index, item := range f.items {
		if item.GetLabel() == label {
			return index
		}
	}
	return -1
}

// GetFocusedItemIndex returns the indices of the form element or button which
// currently has focus. If they don't, -1 is returned respectively.
func (f *Form) GetFocusedItemIndex() (formItem, button int) {
	index := f.focusIndex()
	if index < 0 {
		return -1, -1
	}
	if index < len(f.items) {
		return index, -1
	}
	return -1, index - len(f.items)
}

// SetCancelFunc sets a handler which is called when the user hits the Escape
// key.
func (f *Form) SetCancelFunc(callback func()) *Form {
	f.cancel = callback
	return f
}

// Draw draws this primitive onto the screen.
func (f *Form) Draw(screen tcell.Screen) {
	f.DrawForSubclass(screen, f)

	// Determine the dimensions.
	x, y, width, height := f.GetInnerRect()
	topLimit := y
	bottomLimit := y + height
	rightLimit := x + width
	startX := x

	// Find the longest label.
	var maxLabelWidth int
	for _, item := range f.items {
		labelWidth := TaggedStringWidth(item.GetLabel())
		if labelWidth > maxLabelWidth {
			maxLabelWidth = labelWidth
		}
	}
	maxLabelWidth++ // Add one space.

	// Calculate positions of form items.
	type position struct{ x, y, width, height int }
	positions := make([]position, len(f.items)+len(f.buttons))
	var (
		focusedPosition position
		lineHeight      = 1
	)
	for index, item := range f.items {
		// Calculate the space needed.
		labelWidth := TaggedStringWidth(item.GetLabel())
		var itemWidth int
		if f.horizontal {
			fieldWidth := item.GetFieldWidth()
			if fieldWidth <= 0 {
				fieldWidth = DefaultFormFieldWidth
			}
			labelWidth++
			itemWidth = labelWidth + fieldWidth
		} else {
			// We want all fields to align vertically.
			labelWidth = maxLabelWidth
			itemWidth = width
		}
		itemHeight := item.GetFieldHeight()
		if itemHeight <= 0 {
			itemHeight = DefaultFormFieldHeight
		}

		// Advance to next line if there is no space.
		if f.horizontal && x+labelWidth+1 >= rightLimit {
			x = startX
			y += lineHeight + 1
			lineHeight = itemHeight
		}

		// Update line height.
		if itemHeight > lineHeight {
			lineHeight = itemHeight
		}

		// Adjust the item's attributes.
		if x+itemWidth >= rightLimit {
			itemWidth = rightLimit - x
		}
		fieldTextColor := f.fieldStyle.GetForeground()
		fieldBackgroundColor := f.fieldStyle.GetBackground()
		item.SetFormAttributes(
			labelWidth,
			f.labelColor,
			f.backgroundColor,
			fieldTextColor,
			fieldBackgroundColor,
		)

		// Save position.
		positions[index].x = x
		positions[index].y = y
		positions[index].width = itemWidth
		positions[index].height = itemHeight
		if item.HasFocus() {
			focusedPosition = positions[index]
		}

		// Advance to next item.
		if f.horizontal {
			x += itemWidth + f.itemPadding
		} else {
			y += itemHeight + f.itemPadding
		}
	}

	// How wide are the buttons?
	buttonWidths := make([]int, len(f.buttons))
	buttonsWidth := 0
	for index, button := range f.buttons {
		w := TaggedStringWidth(button.GetLabel()) + 4
		buttonWidths[index] = w
		buttonsWidth += w + 1
	}
	buttonsWidth--

	// Where do we place them?
	if !f.horizontal && x+buttonsWidth < rightLimit {
		switch f.buttonsAlignment {
		case AlignmentRight:
			x = rightLimit - buttonsWidth
		case AlignmentCenter:
			x = (x + rightLimit - buttonsWidth) / 2
		}

		// In vertical layouts, buttons always appear after an empty line.
		if f.itemPadding == 0 && len(f.items) > 0 {
			y++
		}
	}

	// Calculate positions of buttons.
	for index, button := range f.buttons {
		space := rightLimit - x
		buttonWidth := buttonWidths[index]
		if f.horizontal {
			if space < buttonWidth-4 {
				x = startX
				y += lineHeight + 1
				space = width
				lineHeight = 1
			}
		} else {
			if space < 1 {
				break // No space for this button anymore.
			}
		}
		if buttonWidth > space {
			buttonWidth = space
		}
		button.SetStyle(f.buttonStyle).
			SetActivatedStyle(f.buttonActivatedStyle).
			SetDisabledStyle(f.buttonDisabledStyle)

		buttonIndex := index + len(f.items)
		positions[buttonIndex].x = x
		positions[buttonIndex].y = y
		positions[buttonIndex].width = buttonWidth
		positions[buttonIndex].height = 1

		if button.HasFocus() {
			focusedPosition = positions[buttonIndex]
		}

		x += buttonWidth + 1
	}

	// Determine vertical offset based on the position of the focused item.
	var offset int
	if focusedPosition.y+focusedPosition.height > bottomLimit {
		offset = focusedPosition.y + focusedPosition.height - bottomLimit
		if focusedPosition.y-offset < topLimit {
			offset = focusedPosition.y - topLimit
		}
	}

	// Draw items.
	for index, item := range f.items {
		// Set position.
		y := positions[index].y - offset
		height := positions[index].height
		item.SetRect(positions[index].x, y, positions[index].width, height)

		// Is this item visible?
		if y+height <= topLimit || y >= bottomLimit {
			continue
		}

		// Draw items with focus last (in case of overlaps).
		if item.HasFocus() {
			defer item.Draw(screen)
		} else {
			item.Draw(screen)
		}
	}

	// Draw buttons.
	for index, button := range f.buttons {
		// Set position.
		buttonIndex := index + len(f.items)
		y := positions[buttonIndex].y - offset
		height := positions[buttonIndex].height
		button.SetRect(positions[buttonIndex].x, y, positions[buttonIndex].width, height)

		// Is this button visible?
		if y+height <= topLimit || y >= bottomLimit {
			continue
		}

		// Draw button.
		button.Draw(screen)
	}
}

// Focus is called by the application when the primitive receives focus.
func (f *Form) Focus(delegate func(p Primitive)) {
	f.setFocus = delegate

	// If there is no current focus, pick one.
	focus := f.focusIndex()
	if f.requestedFocus >= 0 {
		focus = f.requestedFocus
	}

	// Delegate focus.
	for index, item := range f.items {
		if (focus < 0 || focus == index) && !item.GetDisabled() {
			f.requestedFocus = index
			delegate(item)
			return
		}
	}
	for index, button := range f.buttons {
		if (focus < 0 || focus == len(f.items)+index) && !button.GetDisabled() {
			f.requestedFocus = len(f.items) + index
			delegate(button)
			return
		}
	}

	f.Box.Focus(delegate)
}

// finished handles a form item's "finished" event.
func (f *Form) finished(key tcell.Key) {
	focus := f.focusIndex()
	if key >= 0 {
		f.lastFinishedKey = key
	}

	totalCount := len(f.items) + len(f.buttons)
	switch key {
	case tcell.KeyTab, tcell.KeyEnter:
		// Find the next focusable item.
		for range totalCount {
			focus = (focus + 1) % totalCount
			if focus < len(f.items) {
				if !f.items[focus].GetDisabled() {
					f.setFocus(f.items[focus])
					return
				}
			} else {
				if !f.buttons[focus-len(f.items)].GetDisabled() {
					f.setFocus(f.buttons[focus-len(f.items)])
					return
				}
			}
		}
	case tcell.KeyBacktab:
		// Find the previous focusable item.
		for range totalCount {
			focus = (focus + totalCount - 1) % totalCount
			if focus < len(f.items) {
				if !f.items[focus].GetDisabled() {
					f.setFocus(f.items[focus])
					return
				}
			} else {
				if !f.buttons[focus-len(f.items)].GetDisabled() {
					f.setFocus(f.buttons[focus-len(f.items)])
					return
				}
			}
		}
	case tcell.KeyEscape:
		if f.cancel != nil {
			f.cancel()
		}
	default:
		if key < 0 && f.lastFinishedKey >= 0 {
			// Repeat the last action.
			f.finished(f.lastFinishedKey)
		}
	}
}

// focusIndex returns the index of the currently focused item, counting form
// items first, then buttons. A negative value indicates that no containeed item
// has focus.
func (f *Form) focusIndex() int {
	for index, item := range f.items {
		if item.HasFocus() {
			return index
		}
	}
	for index, button := range f.buttons {
		if button.HasFocus() {
			return len(f.items) + index
		}
	}
	return -1
}

// HasFocus returns whether or not this primitive has focus.
func (f *Form) HasFocus() bool {
	if f.focusIndex() >= 0 {
		return true
	}
	return f.Box.HasFocus()
}

// MouseHandler returns the mouse handler for this primitive.
func (f *Form) MouseHandler(action MouseAction, event *tcell.EventMouse) (Primitive, Command) {
	var (
		capture Primitive
		cmd     Command
	)
	// Determine items to pass mouse events to.
	for _, item := range f.items {
		if item.GetDisabled() {
			continue
		}
		var childCmds Command
		capture, childCmds = item.MouseHandler(action, event)
		cmd = AppendCommand(cmd, childCmds)
		if childCmds != nil {
			return capture, cmd
		}
	}
	for _, button := range f.buttons {
		if button.GetDisabled() {
			continue
		}
		var childCmds Command
		capture, childCmds = button.MouseHandler(action, event)
		cmd = AppendCommand(cmd, childCmds)
		if childCmds != nil {
			return capture, cmd
		}
	}

	// A mouse down anywhere else will focus this form.
	if action == MouseLeftDown && f.InRect(event.Position()) {
		cmd = AppendCommand(cmd, SetFocusCommand{Target: f})
		cmd = AppendCommand(cmd, ConsumeEventCommand{})
	}

	return capture, cmd
}

// InputHandler returns the handler for this primitive.
func (f *Form) InputHandler(event *tcell.EventKey) Command {
	for _, item := range f.items {
		if item.HasFocus() {
			return item.InputHandler(event)
		}
	}

	for _, button := range f.buttons {
		if button.HasFocus() {
			return button.InputHandler(event)
		}
	}
	return nil
}

// PasteHandler handles pasted text for this primitive.
func (f *Form) PasteHandler(pastedText string) Command {
	for _, item := range f.items {
		if item.HasFocus() {
			return item.PasteHandler(pastedText)
		}
	}

	for _, button := range f.buttons {
		if button.HasFocus() {
			return button.PasteHandler(pastedText)
		}
	}
	return nil
}
