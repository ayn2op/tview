package tview

import (
	"math"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v3"
	"github.com/rivo/uniseg"
)

const (
	AutocompletedNavigate = iota // The user navigated the autocomplete list (using the errow keys).
	AutocompletedTab             // The user selected an autocomplete entry using the tab key.
	AutocompletedEnter           // The user selected an autocomplete entry using the enter key.
	AutocompletedClick           // The user selected an autocomplete entry by clicking the mouse button on it.
)

// InputField is a one-line box into which the user can enter text. Use
// [InputField.SetAcceptanceFunc] to accept or reject input,
// [InputField.SetChangedFunc] to listen for changes, and
// [InputField.SetMaskCharacter] to hide input from onlookers (e.g. for password
// input).
//
// The input field also has an optional autocomplete feature. It is initialized
// by the [InputField.SetAutocompleteFunc] function. For more control over the
// autocomplete drop-down's behavior, you can also set the
// [InputField.SetAutocompletedFunc].
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
// If autocomplete functionality is configured:
//
//   - Down arrow: Open the autocomplete drop-down.
//   - Tab, Enter: Select the current autocomplete entry.
//
// See https://github.com/ayn2op/tview/wiki/InputField for an example.
type InputField struct {
	*Box

	// The text area providing the core functionality of the input field.
	textArea *TextArea

	// The screen width of the input area. A value of 0 means extend as much as
	// possible.
	fieldWidth int

	// An optional autocomplete function which receives the current text of the
	// input field and returns a slice of strings to be displayed in a drop-down
	// selection.
	autocomplete func(text string) []string

	// The List object which shows the selectable autocomplete entries. If not
	// nil, the list's items represent the current autocomplete entries.
	autocompleteList      *List
	autocompleteListMutex sync.Mutex
	autocompleteEntries   []string    // Canonical entry text by index.
	autocompleteItems     []*TextView // Cached non-selected row views.
	autocompleteSelected  *TextView   // Reused selected-row view.

	// The styles of the autocomplete entries.
	autocompleteStyles struct {
		main       tcell.Style
		selected   tcell.Style
		background tcell.Color
	}

	// An optional function which is called when the user selects an
	// autocomplete entry. The text and index of the selected entry (within the
	// list) is provided, as well as the user action causing the selection (one
	// of the "Autocompleted" values). The function should return true if the
	// autocomplete list should be closed. If nil, the input field will be
	// updated automatically when the user navigates the autocomplete list.
	autocompleted func(text string, index int, source int) bool

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

func (i *InputField) clearAutocompleteListLocked() {
	if i.autocompleteList != nil {
		unbindDirtyParent(i.autocompleteList, i.Box)
		i.autocompleteList = nil
	}
	i.autocompleteEntries = nil
	i.autocompleteItems = nil
	i.autocompleteSelected = nil
	i.MarkDirty()
}

func (i *InputField) newAutocompleteItem(text string, style tcell.Style) *TextView {
	return NewTextView().
		SetScrollable(false).
		SetWrap(false).
		SetWordWrap(false).
		SetTextStyle(style).
		SetLines([]Line{{{Text: text, Style: style}}})
}

func (i *InputField) rebuildAutocompleteItemsLocked(entries []string) {
	i.autocompleteEntries = append(i.autocompleteEntries[:0], entries...)
	i.autocompleteItems = make([]*TextView, len(i.autocompleteEntries))

	for index, entry := range i.autocompleteEntries {
		i.autocompleteItems[index] = i.newAutocompleteItem(entry, i.autocompleteStyles.main)
	}
	if i.autocompleteSelected == nil {
		i.autocompleteSelected = i.newAutocompleteItem("", i.autocompleteStyles.selected)
	}

	i.autocompleteList.SetBuilder(func(index int, cursor int) ListItem {
		if index < 0 || index >= len(i.autocompleteEntries) {
			return nil
		}
		if index == cursor {
			style := i.autocompleteStyles.selected
			i.autocompleteSelected.SetTextStyle(style)
			i.autocompleteSelected.SetLines([]Line{{{Text: i.autocompleteEntries[index], Style: style}}})
			return i.autocompleteSelected
		}
		return i.autocompleteItems[index]
	})
}

func (i *InputField) selectAutocompleteEntryLocked(index int, source int, currentText *string, skipAutocomplete *bool) {
	if index < 0 || index >= len(i.autocompleteEntries) {
		return
	}

	text := i.autocompleteEntries[index]
	if i.autocompleted != nil {
		if i.autocompleted(text, index, source) {
			i.clearAutocompleteListLocked()
			*currentText = i.GetText()
		}
		return
	}

	i.SetText(text)
	if source == AutocompletedNavigate {
		*currentText = text
		return
	}

	*skipAutocomplete = true
	i.clearAutocompleteListLocked()
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
	i.autocompleteStyles.main = tcell.StyleDefault.Background(Styles.MoreContrastBackgroundColor).Foreground(Styles.PrimitiveBackgroundColor)
	i.autocompleteStyles.selected = tcell.StyleDefault.Background(Styles.PrimaryTextColor).Foreground(Styles.PrimitiveBackgroundColor)
	i.autocompleteStyles.background = Styles.MoreContrastBackgroundColor
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

// SetAutocompleteStyles sets the colors and styles of autocomplete entries.
func (i *InputField) SetAutocompleteStyles(background tcell.Color, main, selected tcell.Style) *InputField {
	if i.autocompleteStyles.background != background || i.autocompleteStyles.main != main || i.autocompleteStyles.selected != selected {
		i.autocompleteStyles.background = background
		i.autocompleteStyles.main = main
		i.autocompleteStyles.selected = selected
		i.MarkDirty()
	}
	return i
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

// SetAutocompleteFunc sets an autocomplete callback function which may return
// strings to be selected from a drop-down based on the current text of the
// input field. The drop-down appears only if len(entries) > 0. The callback is
// invoked in this function and whenever the current text changes or when
// [InputField.Autocomplete] is called. Entries are cleared when the user
// selects an entry or presses Escape.
func (i *InputField) SetAutocompleteFunc(callback func(currentText string) (entries []string)) *InputField {
	i.autocomplete = callback
	i.Autocomplete()
	return i
}

// SetAutocompletedFunc sets a callback function which is invoked when the user
// selects an entry from the autocomplete drop-down list. The function is passed
// the text of the selected entry, the index of the
// entry, and the user action that caused the selection, for example
// [AutocompletedNavigate]. It returns true if the autocomplete drop-down should
// be closed after the callback returns or false if it should remain open, in
// which case [InputField.Autocomplete] is called to update the drop-down's
// contents.
//
// If no such callback is set (or nil is provided), the input field will be
// updated with the selection any time the user navigates the autocomplete
// drop-down list. So this function essentially gives you more control over the
// autocomplete functionality.
func (i *InputField) SetAutocompletedFunc(autocompleted func(text string, index int, source int) bool) *InputField {
	i.autocompleted = autocompleted
	return i
}

// Autocomplete invokes the autocomplete callback (if there is one, see
// [InputField.SetAutocompleteFunc]). If the length of the returned autocomplete
// entries slice is greater than 0, the input field will present the user with a
// corresponding drop-down list the next time the input field is drawn.
//
// It is safe to call this function from any goroutine. Note that the input
// field is not redrawn automatically unless called from the main goroutine
// (e.g. in response to events).
func (i *InputField) Autocomplete() *InputField {
	i.autocompleteListMutex.Lock()
	defer i.autocompleteListMutex.Unlock()
	if i.autocomplete == nil {
		return i
	}

	// Do we have any autocomplete entries?
	text := i.textArea.GetText()
	entries := i.autocomplete(text)
	if len(entries) == 0 {
		// No entries, no list.
		i.clearAutocompleteListLocked()
		return i
	}

	// Make a list if we have none.
	if i.autocompleteList == nil {
		i.autocompleteList = NewList()
		bindDirtyParent(i.autocompleteList, i.Box)
		i.autocompleteList.SetSnapToItems(true).
			SetBackgroundColor(i.autocompleteStyles.background)
	}

	// Fill it with the entries.
	currentIndex := i.autocompleteList.Cursor()
	var currentSelection string
	if currentIndex >= 0 && currentIndex < len(i.autocompleteEntries) {
		currentSelection = i.autocompleteEntries[currentIndex]
	}
	currentEntry := -1
	suffixLength := math.MaxInt
	i.rebuildAutocompleteItemsLocked(entries)
	for index, entry := range entries {
		if currentSelection != "" && entry == currentSelection {
			currentEntry = index
		}
		if currentSelection == "" && strings.HasPrefix(entry, text) && len(entry)-len(text) < suffixLength {
			currentEntry = index
			suffixLength = len(text) - len(entry)
		}
	}

	// Set the selection if we have one.
	i.autocompleteList.SetCursor(currentEntry)

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
	i.autocompleteListMutex.Lock()
	defer i.autocompleteListMutex.Unlock()
	i.clearAutocompleteListLocked() // Hide the autocomplete drop-down.
}

// IsDirty returns whether this primitive or one of its children needs redraw.
func (i *InputField) IsDirty() bool {
	if i.Box.IsDirty() || i.textArea.IsDirty() {
		return true
	}
	i.autocompleteListMutex.Lock()
	defer i.autocompleteListMutex.Unlock()
	return i.autocompleteList != nil && i.autocompleteList.IsDirty()
}

// MarkClean marks this primitive and children as clean.
func (i *InputField) MarkClean() {
	i.Box.MarkClean()
	i.textArea.MarkClean()
	i.autocompleteListMutex.Lock()
	defer i.autocompleteListMutex.Unlock()
	if i.autocompleteList != nil {
		i.autocompleteList.MarkClean()
	}
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

	// Draw autocomplete list.
	i.autocompleteListMutex.Lock()
	defer i.autocompleteListMutex.Unlock()
	if i.autocompleteList != nil && i.HasFocus() {
		// How much space do we need?
		lheight := len(i.autocompleteEntries)
		lwidth := 0
		for _, entry := range i.autocompleteEntries {
			width := TaggedStringWidth(entry)
			if width > lwidth {
				lwidth = width
			}
		}

		// We prefer to drop down but if there is no space, maybe drop up?
		lx := x + labelWidth
		ly := y + 1
		_, sheight := screen.Size()
		if ly+lheight >= sheight && ly-2 > lheight-ly {
			ly = max(y-lheight, 0)
		}
		if ly+lheight >= sheight {
			lheight = sheight - ly
		}
		i.autocompleteList.SetRect(lx, ly, lwidth, lheight)
		i.autocompleteList.Draw(screen)
	}
}

// InputHandler returns the handler for this primitive.
func (i *InputField) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return i.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p Primitive)) {
		if i.textArea.GetDisabled() {
			return
		}

		// Trigger changed events.
		var skipAutocomplete bool
		currentText := i.textArea.GetText()
		defer func() {
			if skipAutocomplete {
				return
			}
			if i.textArea.GetText() != currentText {
				i.Autocomplete()
			}
		}()

		// If we have an autocomplete list, there are certain keys we will
		// forward to it.
		i.autocompleteListMutex.Lock()
		defer i.autocompleteListMutex.Unlock()
		if i.autocompleteList != nil {
			i.autocompleteList.SetChangedFunc(nil)
			switch key := event.Key(); key {
			case tcell.KeyEscape: // Close the list.
				i.clearAutocompleteListLocked()
				return
			case tcell.KeyEnter, tcell.KeyTab: // Intentional selection.
				source := AutocompletedEnter
				if key == tcell.KeyTab {
					source = AutocompletedTab
				}
				i.selectAutocompleteEntryLocked(i.autocompleteList.Cursor(), source, &currentText, &skipAutocomplete)
				return
			case tcell.KeyDown, tcell.KeyUp, tcell.KeyPgDn, tcell.KeyPgUp:
				i.autocompleteList.SetChangedFunc(func(index int) {
					i.selectAutocompleteEntryLocked(index, AutocompletedNavigate, &currentText, &skipAutocomplete)
				})
				i.autocompleteList.InputHandler()(event, setFocus)
				return
			}
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
		case tcell.KeyDown:
			i.autocompleteListMutex.Unlock() // We're still holding a lock.
			i.Autocomplete()
			i.autocompleteListMutex.Lock()
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

		var skipAutocomplete bool
		currentText := i.GetText()
		defer func() {
			if skipAutocomplete {
				return
			}
			if i.textArea.GetText() != currentText {
				i.Autocomplete()
			}
		}()

		// If we have an autocomplete list, forward the mouse event to it.
		i.autocompleteListMutex.Lock()
		defer i.autocompleteListMutex.Unlock()
		if i.autocompleteList != nil {
			i.autocompleteList.SetChangedFunc(nil)
			previousCursor := i.autocompleteList.Cursor()
			if consumed, _ = i.autocompleteList.MouseHandler()(action, event, setFocus); consumed {
				setFocus(i)
				if action == MouseLeftClick {
					i.selectAutocompleteEntryLocked(i.autocompleteList.Cursor(), AutocompletedClick, &currentText, &skipAutocomplete)
				} else if i.autocompleteList.Cursor() != previousCursor {
					i.selectAutocompleteEntryLocked(i.autocompleteList.Cursor(), AutocompletedNavigate, &currentText, &skipAutocomplete)
				}
				return
			}
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

		// The autocomplete drop down may be open.
		i.autocompleteListMutex.Lock()
		defer i.autocompleteListMutex.Unlock()
		if i.autocompleteList != nil {
			return
		}

		// Forward the pasted text to the text area.
		i.textArea.PasteHandler()(pastedText, setFocus)
	})
}
