package help

import (
	"strings"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/keybind"
	"github.com/gdamore/tcell/v3"
)

type KeyMap interface {
	// ShortHelp returns keybinds for single-line help.
	ShortHelp() []keybind.Keybind
	// FullHelp returns keybind groups, where each top-level entry is a column.
	FullHelp() [][]keybind.Keybind
}

type Help struct {
	*tview.Box
	Styles Styles

	keyMap           KeyMap
	showAll          bool
	compactModifiers bool
	shortSeparator   string
	fullSeparator    string
	ellipsis         string
}

func New() *Help {
	return &Help{
		Box:            tview.NewBox(),
		Styles:         DefaultStyles(),
		shortSeparator: " • ",
		fullSeparator:  "    ",
		ellipsis:       "…",
	}
}

// SetKeyMap sets the key map used by this help model.
func (h *Help) SetKeyMap(keyMap KeyMap) *Help {
	h.keyMap = keyMap
	return h
}

// SetShowAll enables or disables full help mode.
func (h *Help) SetShowAll(showAll bool) *Help {
	h.showAll = showAll
	return h
}

// SetCompactModifiers enables or disables compact modifier rendering.
func (h *Help) SetCompactModifiers(compact bool) *Help {
	h.compactModifiers = compact
	return h
}

// ShowAll returns whether full help mode is enabled.
func (h *Help) ShowAll() bool {
	return h.showAll
}

// SetShortSeparator sets the separator used in short help mode.
func (h *Help) SetShortSeparator(separator string) *Help {
	h.shortSeparator = separator
	return h
}

// SetFullSeparator sets the separator used between full help columns.
func (h *Help) SetFullSeparator(separator string) *Help {
	h.fullSeparator = separator
	return h
}

// SetEllipsis sets the ellipsis marker used when content is truncated.
func (h *Help) SetEllipsis(ellipsis string) *Help {
	h.ellipsis = ellipsis
	return h
}

// SetStyles sets help styles.
func (h *Help) SetStyles(styles Styles) *Help {
	h.Styles = styles
	return h
}

// Draw draws this model onto the screen.
func (h *Help) Draw(screen tcell.Screen) {
	h.DrawForSubclass(screen, h)

	if h.keyMap == nil {
		return
	}

	x, y, width, height := h.InnerRect()

	var lines [][]segment
	if h.showAll {
		lines = h.fullHelpSegments(h.keyMap.FullHelp(), width)
	} else {
		lines = [][]segment{h.shortHelpSegments(h.keyMap.ShortHelp(), width)}
	}

	for row := 0; row < len(lines) && row < height; row++ {
		h.drawSegments(screen, x, y+row, width, lines[row])
	}
}

// FullHelpLines renders grouped help into full mode lines as plain text.
func (h *Help) FullHelpLines(groups [][]keybind.Keybind, maxWidth int) []string {
	styled := h.fullHelpSegments(groups, maxWidth)
	lines := make([]string, 0, len(styled))
	for _, line := range styled {
		var b strings.Builder
		for _, s := range line {
			b.WriteString(s.text)
		}
		lines = append(lines, b.String())
	}
	return lines
}

type segment struct {
	text  string
	style tcell.Style
}

func (h *Help) shortHelpSegments(bindings []keybind.Keybind, maxWidth int) []segment {
	items := make([][]segment, 0, len(bindings))
	for _, kb := range bindings {
		hp := kb.Help()
		item := shortItemSegments(h.formatKey(hp.Key), hp.Desc, h.Styles.ShortKeyStyle, h.Styles.ShortDescStyle)
		if len(item) == 0 {
			continue
		}
		items = append(items, item)
	}
	if len(items) == 0 {
		return nil
	}

	sepText := h.shortSeparator
	if sepText == "" {
		sepText = " "
	}
	sep := segment{text: sepText, style: h.Styles.ShortSeparatorStyle}

	out := cloneSegments(items[0])
	for i := 1; i < len(items); i++ {
		candidate := append(cloneSegments(out), sep)
		candidate = append(candidate, items[i]...)
		if maxWidth > 0 && segmentsWidth(candidate) > maxWidth {
			tail := h.truncationTail(out, maxWidth)
			if len(tail) > 0 {
				out = append(out, tail...)
			}
			return out
		}
		out = candidate
	}

	if maxWidth > 0 && segmentsWidth(out) > maxWidth {
		return nil
	}
	return out
}

func (h *Help) fullHelpSegments(groups [][]keybind.Keybind, maxWidth int) [][]segment {
	type entry struct {
		key  string
		desc string
	}
	type column struct {
		entries []entry
		keyW    int
		colW    int
	}

	columns := make([]column, 0, len(groups))
	for _, group := range groups {
		col := column{}
		for _, kb := range group {
			hp := kb.Help()
			if hp.Key == "" && hp.Desc == "" {
				continue
			}
			keyText := h.formatKey(hp.Key)
			col.entries = append(col.entries, entry{key: keyText, desc: hp.Desc})
			kw := tview.TaggedStringWidth(keyText)
			if kw > col.keyW {
				col.keyW = kw
			}
		}
		if len(col.entries) == 0 {
			continue
		}
		// colW stores the widest rendered row in this column so we can keep separators aligned.
		for _, e := range col.entries {
			w := col.keyW
			if e.key != "" && e.desc != "" {
				w += 1
			}
			w += tview.TaggedStringWidth(e.desc)
			if w > col.colW {
				col.colW = w
			}
		}
		columns = append(columns, col)
	}

	if len(columns) == 0 {
		return nil
	}

	sepText := h.fullSeparator
	if sepText == "" {
		sepText = " "
	}
	sepW := tview.TaggedStringWidth(sepText)

	included := 0
	totalW := 0
	// We include columns left-to-right until the next column would overflow maxWidth.
	for i, col := range columns {
		nextW := col.colW
		if i > 0 {
			nextW += sepW
		}
		if maxWidth > 0 && totalW+nextW > maxWidth {
			break
		}
		included++
		totalW += nextW
	}

	if included == 0 {
		return [][]segment{{{text: h.ellipsis, style: h.Styles.EllipsisStyle}}}
	}
	truncated := included < len(columns)

	maxRows := 0
	for i := range included {
		if len(columns[i].entries) > maxRows {
			maxRows = len(columns[i].entries)
		}
	}

	lines := make([][]segment, 0, maxRows)
	for row := range maxRows {
		line := make([]segment, 0, included*4)
		for col := range included {
			if col > 0 {
				line = append(line, segment{text: sepText, style: h.Styles.FullSeparatorStyle})
			}

			c := columns[col]
			cell := make([]segment, 0, 4)
			if row >= len(c.entries) {
				// Empty rows still occupy full column width so the following separators do not drift.
				cell = append(cell, segment{text: strings.Repeat(" ", c.colW), style: h.Styles.FullDescStyle})
				line = append(line, cell...)
				continue
			}

			e := c.entries[row]
			keyPad := c.keyW - tview.TaggedStringWidth(e.key)
			if e.key != "" {
				cell = append(cell, segment{text: e.key, style: h.Styles.FullKeyStyle})
			}
			if keyPad > 0 {
				cell = append(cell, segment{text: strings.Repeat(" ", keyPad), style: h.Styles.FullKeyStyle})
			}
			if e.key != "" && e.desc != "" {
				cell = append(cell, segment{text: " ", style: h.Styles.FullDescStyle})
			}
			if e.desc != "" {
				cell = append(cell, segment{text: e.desc, style: h.Styles.FullDescStyle})
			}

			// Every non-last column is padded to fixed width so row-specific content lengths do not shift separators.
			if col < included-1 {
				cellWidth := segmentsWidth(cell)
				if pad := c.colW - cellWidth; pad > 0 {
					cell = append(cell, segment{text: strings.Repeat(" ", pad), style: h.Styles.FullDescStyle})
				}
			}

			line = append(line, cell...)
		}
		lines = append(lines, line)
	}

	if truncated && len(lines) > 0 {
		tail := h.truncationTail(lines[0], maxWidth)
		if len(tail) > 0 {
			lines[0] = append(lines[0], tail...)
		}
	}

	return lines
}

func (h *Help) truncationTail(current []segment, maxWidth int) []segment {
	if maxWidth <= 0 || h.ellipsis == "" {
		return nil
	}
	// We only add an ellipsis when it fully fits because clipping looks broken in narrow widths.
	tail := []segment{
		{text: " ", style: h.Styles.EllipsisStyle},
		{text: h.ellipsis, style: h.Styles.EllipsisStyle},
	}
	if segmentsWidth(current)+segmentsWidth(tail) <= maxWidth {
		return tail
	}
	return nil
}

func (h *Help) drawSegments(screen tcell.Screen, x, y, width int, segments []segment) {
	if width <= 0 || len(segments) == 0 {
		return
	}

	cursor := x
	remaining := width
	for _, s := range segments {
		if s.text == "" || remaining <= 0 {
			continue
		}
		_, printedWidth := tview.PrintWithStyle(screen, s.text, cursor, y, remaining, tview.AlignmentLeft, s.style)
		cursor += printedWidth
		remaining -= printedWidth
	}
}

func shortItemSegments(key, desc string, keyStyle, descStyle tcell.Style) []segment {
	switch {
	case key == "" && desc == "":
		return nil
	case key == "":
		return []segment{{text: desc, style: descStyle}}
	case desc == "":
		return []segment{{text: key, style: keyStyle}}
	default:
		return []segment{{text: key, style: keyStyle}, {text: " ", style: descStyle}, {text: desc, style: descStyle}}
	}
}

func (h *Help) formatKey(key string) string {
	if !h.compactModifiers {
		return key
	}

	replacer := strings.NewReplacer(
		"Ctrl+", "^",
		"ctrl+", "^",
		"Control+", "^",
		"control+", "^",
		"Shift+", "S-",
		"shift+", "S-",
		"Alt+", "A-",
		"alt+", "A-",
		"Meta+", "M-",
		"meta+", "M-",
	)
	return replacer.Replace(key)
}

func segmentsWidth(segments []segment) int {
	width := 0
	for _, segment := range segments {
		width += tview.TaggedStringWidth(segment.text)
	}
	return width
}

func cloneSegments(in []segment) []segment {
	out := make([]segment, len(in))
	copy(out, in)
	return out
}
