package tview

import (
	"math"
	"sync"

	"github.com/gdamore/tcell/v3"
	"github.com/rivo/uniseg"
)

// TabSize is the number of spaces with which a tab character will be replaced.
var TabSize = 4

type textViewCell struct {
	text          string
	style         tcell.Style
	width         int
	optionalBreak bool
	mustBreak     bool
}

type textViewLogicalLine struct {
	segments Line
	cells    []textViewCell
	width    int
}

type textViewLine struct {
	logical int
	start   int
	end     int
	width   int
}

// TextViewWriter is a writer that can be used to write to and clear a TextView
// in batches, i.e. multiple writes with the lock only being acquired once. Don't
// instantiated this class directly but use the TextView's BatchWriter method
// instead.
type TextViewWriter struct {
	t *TextView
}

// Close implements io.Closer for the writer by unlocking the original TextView.
func (w TextViewWriter) Close() error {
	w.t.Unlock()
	return nil
}

// Clear removes all text from the buffer.
func (w TextViewWriter) Clear() {
	w.t.clear()
}

// Write implements the io.Writer interface. It behaves like the TextView's
// Write() method except that it does not acquire the lock.
func (w TextViewWriter) Write(p []byte) (n int, err error) {
	return w.t.write(p)
}

// HasFocus returns whether the underlying TextView has focus.
func (w TextViewWriter) HasFocus() bool {
	return w.t.hasFocus
}

// TextView is a component to display read-only text. The content is represented
// as styled segments grouped by lines.
type TextView struct {
	sync.Mutex
	*Box

	// The requested size of the text area. If set to 0, the text view will use
	// the entire available space. This only affects rendering in Draw.
	width, height int

	// The logical lines.
	lines []textViewLogicalLine

	// Wrapped visual lines for the current width.
	wrapped []textViewLine

	// The screen width of the longest visual line in the current wrap index.
	longestLine int

	// The width used to build wrapped.
	lastWidth int

	// The label text shown, usually when part of a form.
	label string

	// The width of the text area's label.
	labelWidth int

	// The label style.
	labelStyle tcell.Style

	// The text alignment, one of AlignLeft, AlignCenter, or AlignRight.
	alignment Alignment

	// The index of the first visual line shown in the text view.
	lineOffset int

	// If set to true, the text view will always remain at the end of the
	// content when text is added.
	trackEnd bool

	// The width of the characters to be skipped on each line (not used in wrap
	// mode).
	columnOffset int

	// The maximum number of logical lines kept in memory. Ignored if 0.
	maxLines int

	// If set to true, the text view will keep a buffer of text which can be
	// navigated when the text is longer than what fits into the box.
	scrollable bool

	// If set to true, lines that are longer than the available width are
	// wrapped onto the next line. If set to false, any characters beyond the
	// available width are discarded.
	wrap bool

	// If set to true and if wrap is also true, Unicode line breaking is
	// applied.
	wordWrap bool

	// The default style for newly written text.
	textStyle tcell.Style

	// An optional function which is called when the content of the text view
	// has changed.
	changed func()

	// An optional function which is called when the user presses one of the
	// following keys: Escape, Enter, Tab, Backtab.
	done func(tcell.Key)

	// A callback function set by the Form class and called when the user leaves
	// this form item.
	finished func(tcell.Key)
}

// NewTextView returns a new text view.
func NewTextView() *TextView {
	return &TextView{
		Box:        NewBox(),
		labelStyle: tcell.StyleDefault.Foreground(Styles.SecondaryTextColor),
		lineOffset: -1,
		scrollable: true,
		alignment:  AlignmentLeft,
		wrap:       true,
		wordWrap:   true,
		textStyle:  tcell.StyleDefault.Background(Styles.PrimitiveBackgroundColor).Foreground(Styles.PrimaryTextColor),
	}
}

// SetLabel sets the text to be displayed before the text view.
func (t *TextView) SetLabel(label string) *TextView {
	if t.label != label {
		t.label = label
		t.MarkDirty()
	}
	return t
}

// GetLabel returns the text to be displayed before the text view.
func (t *TextView) GetLabel() string {
	return t.label
}

// SetLabelWidth sets the screen width of the label. A value of 0 will cause the
// primitive to use the width of the label string.
func (t *TextView) SetLabelWidth(width int) *TextView {
	if t.labelWidth != width {
		t.labelWidth = width
		t.MarkDirty()
	}
	return t
}

// SetSize sets the screen size of the main text element of the text view.
func (t *TextView) SetSize(rows, columns int) *TextView {
	if t.width != columns || t.height != rows {
		t.width = columns
		t.height = rows
		t.MarkDirty()
	}
	return t
}

// GetFieldWidth returns this primitive's field width.
func (t *TextView) GetFieldWidth() int {
	return t.width
}

// GetFieldHeight returns this primitive's field height.
func (t *TextView) GetFieldHeight() int {
	return t.height
}

// SetDisabled sets whether or not the item is disabled / read-only.
func (t *TextView) SetDisabled(disabled bool) FormItem {
	return t // Text views are always read-only.
}

// GetDisabled returns whether or not the item is disabled / read-only.
func (t *TextView) GetDisabled() bool {
	return true // Text views are always read-only.
}

// SetScrollable sets the flag that decides whether or not the text view is
// scrollable. If false, text that moves above the text view's top row will be
// permanently deleted.
func (t *TextView) SetScrollable(scrollable bool) *TextView {
	if t.scrollable != scrollable {
		t.scrollable = scrollable
		if !scrollable {
			t.trackEnd = true
		}
		t.MarkDirty()
	}
	return t
}

// SetWrap sets the flag that, if true, leads to lines that are longer than the
// available width being wrapped onto the next line. If false, any characters
// beyond the available width are not displayed.
func (t *TextView) SetWrap(wrap bool) *TextView {
	if t.wrap != wrap {
		t.resetLayout()
		t.MarkDirty()
	}
	t.wrap = wrap
	return t
}

// SetWordWrap sets the flag that, if true and if the "wrap" flag is also true,
// wraps according to Unicode line break opportunities.
func (t *TextView) SetWordWrap(wrapOnWords bool) *TextView {
	if t.wordWrap != wrapOnWords {
		t.resetLayout()
		t.MarkDirty()
	}
	t.wordWrap = wrapOnWords
	return t
}

// SetMaxLines sets the maximum number of logical lines for this text view.
func (t *TextView) SetMaxLines(maxLines int) *TextView {
	if t.maxLines != maxLines {
		t.maxLines = maxLines
		t.MarkDirty()
	}
	return t
}

// SetTextAlign sets the text alignment within the text view. This must be
// either AlignLeft, AlignCenter, or AlignRight.
func (t *TextView) SetTextAlign(alignment Alignment) *TextView {
	if t.alignment != alignment {
		t.alignment = alignment
		t.resetLayout()
		t.MarkDirty()
	}
	return t
}

// SetBackgroundColor overrides its implementation in Box to set the background
// color of this primitive. For backwards compatibility reasons, it also sets
// the background color of the default text style.
func (t *TextView) SetBackgroundColor(color tcell.Color) *Box {
	t.Box.SetBackgroundColor(color)
	style := t.textStyle.Background(color)
	if t.textStyle != style {
		t.textStyle = style
		t.MarkDirty()
	}
	return t.Box
}

// SetTextStyle sets the default style for newly written text.
func (t *TextView) SetTextStyle(style tcell.Style) *TextView {
	if t.textStyle != style {
		t.textStyle = style
		t.MarkDirty()
	}
	return t
}

// SetText sets the text of this text view to the provided plain string.
func (t *TextView) SetText(text string) *TextView {
	t.Lock()
	defer t.Unlock()
	if t.GetText() == text {
		return t
	}
	t.clear()
	t.appendText(text, t.textStyle)
	t.MarkDirty()
	if t.changed != nil {
		go t.changed()
	}
	return t
}

// SetLines replaces the content with styled lines.
func (t *TextView) SetLines(lines []Line) *TextView {
	t.Lock()
	defer t.Unlock()

	t.lines = make([]textViewLogicalLine, 0, len(lines))
	for _, line := range lines {
		copied := make(Line, 0, len(line))
		for _, seg := range line {
			if seg.Text == "" {
				continue
			}
			copied = append(copied, seg)
		}
		t.lines = append(t.lines, textViewLogicalLine{segments: copied})
	}
	t.rebuildCells()
	t.resetLayout()
	t.MarkDirty()
	if t.changed != nil {
		go t.changed()
	}
	return t
}

// GetLines returns a copy of the styled content.
func (t *TextView) GetLines() []Line {
	t.Lock()
	defer t.Unlock()

	out := make([]Line, 0, len(t.lines))
	for _, line := range t.lines {
		copied := make(Line, len(line.segments))
		copy(copied, line.segments)
		out = append(out, copied)
	}
	return out
}

// AppendSegments appends styled segments to the last line.
func (t *TextView) AppendSegments(segments ...Segment) *TextView {
	t.Lock()
	defer t.Unlock()
	for _, seg := range segments {
		t.appendText(seg.Text, seg.Style)
	}
	t.MarkDirty()
	if t.changed != nil {
		go t.changed()
	}
	return t
}

// AppendLine appends a new line made of segments.
func (t *TextView) AppendLine(line Line) *TextView {
	t.Lock()
	defer t.Unlock()
	if len(t.lines) == 0 {
		t.lines = append(t.lines, textViewLogicalLine{})
	}
	for _, seg := range line {
		t.appendText(seg.Text, seg.Style)
	}
	t.lines = append(t.lines, textViewLogicalLine{})
	t.rebuildCells()
	t.resetLayout()
	t.MarkDirty()
	if t.changed != nil {
		go t.changed()
	}
	return t
}

// GetText returns the current plain text of this text view.
func (t *TextView) GetText() string {
	if len(t.lines) == 0 {
		return ""
	}
	result := ""
	for i, line := range t.lines {
		for _, seg := range line.segments {
			result += seg.Text
		}
		if i < len(t.lines)-1 {
			result += "\n"
		}
	}
	return result
}

// GetOriginalLineCount returns the number of logical lines in the current text.
func (t *TextView) GetOriginalLineCount() int {
	if len(t.lines) == 0 {
		return 0
	}
	return len(t.lines)
}

// GetWrappedLineCount returns the number of visual lines, taking wrapping into account.
func (t *TextView) GetWrappedLineCount() int {
	if len(t.lines) == 0 {
		return 0
	}
	width := t.lastWidth
	if width == 0 {
		width = t.width
	}
	t.buildWrapped(width)
	return len(t.wrapped)
}

// Height returns the required height for rendering the text view at the given
// width when used as a scroll list item.
func (t *TextView) Height(width int) int {
	if width < 1 {
		return 1
	}
	if len(t.lines) == 0 {
		return 1
	}
	t.buildWrapped(width)
	if len(t.wrapped) == 0 {
		return 1
	}
	return len(t.wrapped)
}

// SetChangedFunc sets a handler function which is called when the text of the
// text view has changed.
func (t *TextView) SetChangedFunc(handler func()) *TextView {
	t.changed = handler
	return t
}

// SetDoneFunc sets a handler which is called when the user presses on the
// following keys: Escape, Enter, Tab, Backtab.
func (t *TextView) SetDoneFunc(handler func(key tcell.Key)) *TextView {
	t.done = handler
	return t
}

// SetFinishedFunc sets a callback invoked when the user leaves this form item.
func (t *TextView) SetFinishedFunc(handler func(key tcell.Key)) FormItem {
	t.finished = handler
	return t
}

// SetFormAttributes sets attributes shared by all form items.
func (t *TextView) SetFormAttributes(labelWidth int, labelColor, bgColor, fieldTextColor, fieldBgColor tcell.Color) FormItem {
	changed := false
	if t.labelWidth != labelWidth {
		t.labelWidth = labelWidth
		changed = true
	}
	if t.backgroundColor != bgColor {
		t.backgroundColor = bgColor
		changed = true
	}
	labelStyle := t.labelStyle.Foreground(labelColor)
	if t.labelStyle != labelStyle {
		t.labelStyle = labelStyle
		changed = true
	}
	textStyle := tcell.StyleDefault.Foreground(fieldTextColor).Background(bgColor)
	if t.textStyle != textStyle {
		t.textStyle = textStyle
		changed = true
	}
	if changed {
		t.MarkDirty()
	}
	return t
}

// ScrollTo scrolls to the specified row and column (both starting with 0).
func (t *TextView) ScrollTo(row, column int) *TextView {
	if !t.scrollable {
		return t
	}
	if t.lineOffset != row || t.columnOffset != column || t.trackEnd {
		t.lineOffset = row
		t.columnOffset = column
		t.trackEnd = false
		t.MarkDirty()
	}
	return t
}

// ScrollToBeginning scrolls to the top left corner of the text if the text view
// is scrollable.
func (t *TextView) ScrollToBeginning() *TextView {
	if !t.scrollable {
		return t
	}
	if t.trackEnd || t.lineOffset != 0 || t.columnOffset != 0 {
		t.trackEnd = false
		t.lineOffset = 0
		t.columnOffset = 0
		t.MarkDirty()
	}
	return t
}

// ScrollToEnd scrolls to the bottom left corner of the text if the text view
// is scrollable.
func (t *TextView) ScrollToEnd() *TextView {
	if !t.scrollable {
		return t
	}
	if !t.trackEnd || t.columnOffset != 0 {
		t.trackEnd = true
		t.columnOffset = 0
		t.MarkDirty()
	}
	return t
}

// GetScrollOffset returns the number of rows and columns that are skipped at
// the top left corner when the text view has been scrolled.
func (t *TextView) GetScrollOffset() (row, column int) {
	return t.lineOffset, t.columnOffset
}

// Clear removes all text from the buffer. This triggers the "changed" callback.
func (t *TextView) Clear() *TextView {
	t.Lock()
	defer t.Unlock()
	if len(t.lines) == 0 {
		return t
	}
	t.clear()
	if t.changed != nil {
		go t.changed()
	}
	return t
}

func (t *TextView) clear() {
	t.lines = nil
	t.resetLayout()
	t.MarkDirty()
}

// Focus is called when this primitive receives focus.
func (t *TextView) Focus(delegate func(p Primitive)) {
	t.Lock()
	if finished := t.finished; finished != nil && !t.scrollable {
		t.Unlock()
		finished(-1)
		return
	}
	t.Box.Focus(delegate)
	t.Unlock()
}

// HasFocus returns whether or not this primitive has focus.
func (t *TextView) HasFocus() bool {
	t.Lock()
	defer t.Unlock()
	return t.Box.HasFocus()
}

// Write lets us implement the io.Writer interface.
func (t *TextView) Write(p []byte) (n int, err error) {
	t.Lock()
	defer t.Unlock()
	return t.write(p)
}

func (t *TextView) write(p []byte) (n int, err error) {
	changed := t.changed
	if changed != nil {
		defer func() {
			go changed()
		}()
	}

	if len(p) == 0 {
		return 0, nil
	}

	t.appendText(string(p), t.textStyle)
	t.MarkDirty()
	return len(p), nil
}

// BatchWriter returns a new writer that can be used to write into the buffer
// but without Locking/Unlocking the buffer on every write.
func (t *TextView) BatchWriter() TextViewWriter {
	t.Lock()
	return TextViewWriter{t: t}
}

func (t *TextView) appendText(text string, style tcell.Style) {
	if len(t.lines) == 0 {
		t.lines = append(t.lines, textViewLogicalLine{})
	}

	lineIndex := len(t.lines) - 1
	for len(text) > 0 {
		nl := -1
		for i := 0; i < len(text); i++ {
			if text[i] == '\n' {
				nl = i
				break
			}
		}

		if nl < 0 {
			t.appendSegment(lineIndex, Segment{Text: text, Style: style})
			break
		}

		if nl > 0 {
			t.appendSegment(lineIndex, Segment{Text: text[:nl], Style: style})
		}

		t.lines = append(t.lines, textViewLogicalLine{})
		lineIndex = len(t.lines) - 1
		text = text[nl+1:]
	}

	t.rebuildCells()
	t.resetLayout()
}

func (t *TextView) appendSegment(lineIndex int, seg Segment) {
	if seg.Text == "" {
		return
	}
	line := &t.lines[lineIndex]
	if n := len(line.segments); n > 0 && line.segments[n-1].Style == seg.Style {
		line.segments[n-1].Text += seg.Text
		return
	}
	line.segments = append(line.segments, seg)
}

func (t *TextView) rebuildCells() {
	for i := range t.lines {
		line := &t.lines[i]
		cells := make([]textViewCell, 0)
		width := 0
		for _, seg := range line.segments {
			state := -1
			str := seg.Text
			for len(str) > 0 {
				cluster, rest, boundaries, next := uniseg.StepString(str, state)
				state = next
				str = rest
				if cluster == "" {
					continue
				}
				// Treat each segment as an in-line fragment unless it explicitly
				// ends with a line break.
				if rest == "" && !uniseg.HasTrailingLineBreakInString(cluster) {
					boundaries &^= uniseg.MaskLine
				}
				cellWidth := boundaries >> uniseg.ShiftWidth
				optionalBreak := (boundaries & uniseg.MaskLine) == uniseg.LineCanBreak
				mustBreak := (boundaries & uniseg.MaskLine) == uniseg.LineMustBreak
				cells = append(cells, textViewCell{
					text:          cluster,
					style:         seg.Style,
					width:         cellWidth,
					optionalBreak: optionalBreak,
					mustBreak:     mustBreak,
				})
				width += cellWidth
			}
		}
		line.cells = cells
		line.width = width
	}
}

func (t *TextView) resetLayout() {
	t.wrapped = nil
	t.longestLine = 0
	t.lastWidth = 0
}

func (t *TextView) buildWrapped(width int) {
	if width <= 0 {
		width = math.MaxInt
	}
	if t.lastWidth == width && t.wrapped != nil {
		return
	}

	t.wrapped = nil
	t.longestLine = 0

	for lineIndex, line := range t.lines {
		cells := line.cells
		if len(cells) == 0 {
			t.wrapped = append(t.wrapped, textViewLine{logical: lineIndex, start: 0, end: 0, width: 0})
			continue
		}

		if !t.wrap || width == math.MaxInt {
			t.wrapped = append(t.wrapped, textViewLine{logical: lineIndex, start: 0, end: len(cells), width: line.width})
			if line.width > t.longestLine {
				t.longestLine = line.width
			}
			continue
		}

		start := 0
		for start < len(cells) {
			pos := start
			lineWidth := 0
			lastOption := -1
			lastOptionWidth := 0
			mustBreak := false

			for pos < len(cells) {
				cw := t.cellWidth(cells[pos], lineWidth)
				if lineWidth+cw > width {
					break
				}
				lineWidth += cw
				if t.wordWrap && cells[pos].optionalBreak {
					lastOption = pos + 1
					lastOptionWidth = lineWidth
				}
				if cells[pos].mustBreak {
					pos++
					mustBreak = true
					break
				}
				pos++
			}

			if pos == start {
				cw := t.cellWidth(cells[pos], 0)
				pos++
				lineWidth = cw
			}

			if !mustBreak && pos < len(cells) && t.wordWrap && lastOption > start {
				pos = lastOption
				lineWidth = lastOptionWidth
			}

			t.wrapped = append(t.wrapped, textViewLine{logical: lineIndex, start: start, end: pos, width: lineWidth})
			if lineWidth > t.longestLine {
				t.longestLine = lineWidth
			}
			start = pos
		}
	}

	t.lastWidth = width
}

func (t *TextView) cellWidth(cell textViewCell, leftPos int) int {
	if cell.text == "\t" {
		if t.alignment == AlignmentLeft {
			return TabSize - leftPos%TabSize
		}
		return TabSize
	}
	return cell.width
}

// Draw draws this primitive onto the screen.
func (t *TextView) Draw(screen tcell.Screen) {
	t.DrawForSubclass(screen, t)
	t.Lock()
	defer t.Unlock()

	x, y, width, height := t.GetInnerRect()
	labelBg := t.labelStyle.GetBackground()
	if t.labelWidth > 0 {
		labelWidth := min(t.labelWidth, width)
		printWithStyle(screen, t.label, x, y, 0, labelWidth, AlignmentLeft, t.labelStyle, labelBg == tcell.ColorDefault)
		x += labelWidth
		width -= labelWidth
	} else {
		_, _, drawnWidth := printWithStyle(screen, t.label, x, y, 0, width, AlignmentLeft, t.labelStyle, labelBg == tcell.ColorDefault)
		x += drawnWidth
		width -= drawnWidth
	}

	if t.width > 0 && t.width < width {
		width = t.width
	}
	if t.height > 0 && t.height < height {
		height = t.height
	}
	if width <= 0 {
		return
	}

	bg := t.textStyle.GetBackground()
	if bg != t.backgroundColor {
		for row := range height {
			for column := range width {
				screen.Put(x+column, y+row, " ", t.textStyle)
			}
		}
	}

	t.buildWrapped(width)

	if t.trackEnd {
		t.lineOffset = len(t.wrapped) - height
	}
	if t.lineOffset > len(t.wrapped)-height {
		t.lineOffset = len(t.wrapped) - height
	}
	if t.lineOffset < 0 {
		t.lineOffset = 0
	}

	if t.alignment == AlignmentLeft || t.alignment == AlignmentRight {
		if t.columnOffset+width > t.longestLine {
			t.columnOffset = t.longestLine - width
		}
		if t.columnOffset < 0 {
			t.columnOffset = 0
		}
	} else {
		half := (t.longestLine - width) / 2
		if half > 0 {
			if t.columnOffset > half {
				t.columnOffset = half
			}
			if t.columnOffset < -half {
				t.columnOffset = -half
			}
		} else {
			t.columnOffset = 0
		}
	}

	for line := t.lineOffset; line < len(t.wrapped); line++ {
		if line-t.lineOffset >= height {
			break
		}

		info := t.wrapped[line]
		cells := t.lines[info.logical].cells[info.start:info.end]

		var skipWidth, xPos int
		switch t.alignment {
		case AlignmentLeft:
			skipWidth = t.columnOffset
		case AlignmentCenter:
			skipWidth = t.columnOffset + (info.width-width)/2
			if skipWidth < 0 {
				skipWidth = 0
				xPos = (width-info.width)/2 - t.columnOffset
			}
		case AlignmentRight:
			maxWidth := max(t.longestLine, width)
			skipWidth = t.columnOffset - (maxWidth - info.width)
			if skipWidth < 0 {
				skipWidth = 0
				xPos = maxWidth - info.width - t.columnOffset
			}
		}

		for _, cell := range cells {
			if xPos >= width {
				break
			}

			w := t.cellWidth(cell, xPos)
			if skipWidth > 0 {
				skipWidth -= w
				continue
			}

			if w > 0 {
				ch := cell.text
				if ch == "\t" {
					ch = " "
				}
				for offset := w - 1; offset >= 0; offset-- {
					if offset == 0 {
						screen.PutStrStyled(x+xPos+offset, y+line-t.lineOffset, ch, cell.style)
					} else {
						screen.Put(x+xPos+offset, y+line-t.lineOffset, " ", cell.style)
					}
				}
			}

			xPos += w
		}
	}

	if !t.scrollable && len(t.lines) > height {
		trim := len(t.lines) - height
		t.lines = t.lines[trim:]
		t.resetLayout()
		t.lineOffset = 0
	}
	if t.maxLines > 0 && len(t.lines) > t.maxLines {
		trim := len(t.lines) - t.maxLines
		t.lines = t.lines[trim:]
		t.resetLayout()
		t.lineOffset = 0
	}
}

// InputHandler returns the handler for this primitive.
func (t *TextView) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return t.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p Primitive)) {
		previousLineOffset, previousColumnOffset, previousTrackEnd := t.lineOffset, t.columnOffset, t.trackEnd
		key := event.Key()

		if key == tcell.KeyEscape || key == tcell.KeyEnter || key == tcell.KeyTab || key == tcell.KeyBacktab {
			if t.done != nil {
				t.done(key)
			}
			if t.finished != nil {
				t.finished(key)
			}
			return
		}

		if !t.scrollable {
			return
		}

		switch key {
		case tcell.KeyRune:
			switch event.Str() {
			case "g":
				t.trackEnd = false
				t.lineOffset = 0
				t.columnOffset = 0
			case "G":
				t.trackEnd = true
				t.columnOffset = 0
			case "j":
				t.lineOffset++
			case "k":
				t.trackEnd = false
				t.lineOffset--
			case "h":
				t.columnOffset--
			case "l":
				t.columnOffset++
			}
		case tcell.KeyHome:
			t.trackEnd = false
			t.lineOffset = 0
			t.columnOffset = 0
		case tcell.KeyEnd:
			t.trackEnd = true
			t.columnOffset = 0
		case tcell.KeyUp:
			t.trackEnd = false
			t.lineOffset--
		case tcell.KeyDown:
			t.lineOffset++
		case tcell.KeyLeft:
			t.columnOffset--
		case tcell.KeyRight:
			t.columnOffset++
		case tcell.KeyPgDn, tcell.KeyCtrlF:
			_, _, _, pageSize := t.GetInnerRect()
			t.lineOffset += pageSize
		case tcell.KeyPgUp, tcell.KeyCtrlB:
			_, _, _, pageSize := t.GetInnerRect()
			t.trackEnd = false
			t.lineOffset -= pageSize
		}
		if t.lineOffset != previousLineOffset || t.columnOffset != previousColumnOffset || t.trackEnd != previousTrackEnd {
			t.MarkDirty()
		}
	})
}

// MouseHandler returns the mouse handler for this primitive.
func (t *TextView) MouseHandler() func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
	return t.WrapMouseHandler(func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
		previousLineOffset, previousTrackEnd := t.lineOffset, t.trackEnd
		x, y := event.Position()
		if !t.InRect(x, y) {
			return false, nil
		}

		_, _, width, _ := t.GetInnerRect()
		switch action {
		case MouseLeftDown:
			setFocus(t)
			consumed = true
		case MouseLeftClick:
			consumed = true
		case MouseScrollUp:
			if !t.scrollable {
				break
			}
			t.trackEnd = false
			t.lineOffset--
			consumed = true
		case MouseScrollDown:
			if !t.scrollable {
				break
			}
			t.lineOffset++
			consumed = true
		case MouseScrollLeft:
			if !t.scrollable {
				break
			}
			t.columnOffset -= width / 2
			consumed = true
		case MouseScrollRight:
			if !t.scrollable {
				break
			}
			t.columnOffset += width / 2
			consumed = true
		}

		if t.lineOffset != previousLineOffset || t.trackEnd != previousTrackEnd {
			t.MarkDirty()
		}

		return consumed, capture
	})
}
