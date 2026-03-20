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

type Model struct {
	*tview.Box
	styles           Styles
	keyMap           KeyMap
	showAll          bool
	compactModifiers bool
	shortSeparator   string
	fullSeparator    string
	ellipsis         string
}

func NewModel() *Model {
	return &Model{
		Box:            tview.NewBox(),
		styles:         DefaultStyles(),
		shortSeparator: " • ",
		fullSeparator:  "    ",
		ellipsis:       "…",
	}
}

// SetShowAll enables or disables full help mode.
func (m *Model) SetShowAll(showAll bool) *Model {
	m.showAll = showAll
	return m
}

// SetCompactModifiers enables or disables compact modifier rendering.
func (m *Model) SetCompactModifiers(compact bool) *Model {
	m.compactModifiers = compact
	return m
}

// ShowAll returns whether full help mode is enabled.
func (m *Model) ShowAll() bool {
	return m.showAll
}

// SetShortSeparator sets the separator used in short help mode.
func (m *Model) SetShortSeparator(separator string) *Model {
	m.shortSeparator = separator
	return m
}

// SetFullSeparator sets the separator used between full help columns.
func (m *Model) SetFullSeparator(separator string) *Model {
	m.fullSeparator = separator
	return m
}

// SetEllipsis sets the ellipsis marker used when content is truncated.
func (m *Model) SetEllipsis(ellipsis string) *Model {
	m.ellipsis = ellipsis
	return m
}

func (m *Model) Styles() Styles {
	return m.styles
}

func (m *Model) SetStyles(styles Styles) *Model {
	m.styles = styles
	return m
}

func (m *Model) KeyMap() KeyMap {
	return m.keyMap
}

func (m *Model) SetKeyMap(keyMap KeyMap) *Model {
	m.keyMap = keyMap
	return m
}

func (m *Model) Draw(screen tcell.Screen) {
	m.DrawForSubclass(screen, m)

	if m.keyMap == nil {
		return
	}

	x, y, width, height := m.InnerRect()

	var lines [][]segment
	if m.showAll {
		lines = m.fullHelpSegments(m.keyMap.FullHelp(), width)
	} else {
		lines = [][]segment{m.shortHelpSegments(m.keyMap.ShortHelp(), width)}
	}

	for row := 0; row < len(lines) && row < height; row++ {
		m.drawSegments(screen, x, y+row, width, lines[row])
	}
}

// FullHelpLines renders grouped help into full mode lines as plain text.
func (m *Model) FullHelpLines(groups [][]keybind.Keybind, maxWidth int) []string {
	styled := m.fullHelpSegments(groups, maxWidth)
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

func (m *Model) shortHelpSegments(bindings []keybind.Keybind, maxWidth int) []segment {
	items := make([][]segment, 0, len(bindings))
	for _, kb := range bindings {
		hp := kb.Help()
		item := shortItemSegments(m.formatKey(hp.Key), hp.Desc, m.styles.ShortKey, m.styles.ShortDesc)
		if len(item) == 0 {
			continue
		}
		items = append(items, item)
	}
	if len(items) == 0 {
		return nil
	}

	sepText := m.shortSeparator
	if sepText == "" {
		sepText = " "
	}
	sep := segment{text: sepText, style: m.styles.ShortSeparator}

	out := cloneSegments(items[0])
	for i := 1; i < len(items); i++ {
		candidate := append(cloneSegments(out), sep)
		candidate = append(candidate, items[i]...)
		if maxWidth > 0 && segmentsWidth(candidate) > maxWidth {
			tail := m.truncationTail(out, maxWidth)
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

func (m *Model) fullHelpSegments(groups [][]keybind.Keybind, maxWidth int) [][]segment {
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
			keyText := m.formatKey(hp.Key)
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

	sepText := m.fullSeparator
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
		return [][]segment{{{text: m.ellipsis, style: m.styles.Ellipsis}}}
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
				line = append(line, segment{text: sepText, style: m.styles.FullSeparator})
			}

			c := columns[col]
			cell := make([]segment, 0, 4)
			if row >= len(c.entries) {
				// Empty rows still occupy full column width so the following separators do not drift.
				cell = append(cell, segment{text: strings.Repeat(" ", c.colW), style: m.styles.FullDesc})
				line = append(line, cell...)
				continue
			}

			e := c.entries[row]
			keyPad := c.keyW - tview.TaggedStringWidth(e.key)
			if e.key != "" {
				cell = append(cell, segment{text: e.key, style: m.styles.FullKey})
			}
			if keyPad > 0 {
				cell = append(cell, segment{text: strings.Repeat(" ", keyPad), style: m.styles.FullKey})
			}
			if e.key != "" && e.desc != "" {
				cell = append(cell, segment{text: " ", style: m.styles.FullDesc})
			}
			if e.desc != "" {
				cell = append(cell, segment{text: e.desc, style: m.styles.FullDesc})
			}

			// Every non-last column is padded to fixed width so row-specific content lengths do not shift separators.
			if col < included-1 {
				cellWidth := segmentsWidth(cell)
				if pad := c.colW - cellWidth; pad > 0 {
					cell = append(cell, segment{text: strings.Repeat(" ", pad), style: m.styles.FullDesc})
				}
			}

			line = append(line, cell...)
		}
		lines = append(lines, line)
	}

	if truncated && len(lines) > 0 {
		tail := m.truncationTail(lines[0], maxWidth)
		if len(tail) > 0 {
			lines[0] = append(lines[0], tail...)
		}
	}

	return lines
}

func (m *Model) truncationTail(current []segment, maxWidth int) []segment {
	if maxWidth <= 0 || m.ellipsis == "" {
		return nil
	}
	// We only add an ellipsis when it fully fits because clipping looks broken in narrow widths.
	tail := []segment{
		{text: " ", style: m.styles.Ellipsis},
		{text: m.ellipsis, style: m.styles.Ellipsis},
	}
	if segmentsWidth(current)+segmentsWidth(tail) <= maxWidth {
		return tail
	}
	return nil
}

func (m *Model) drawSegments(screen tcell.Screen, x, y, width int, segments []segment) {
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

func (m *Model) formatKey(key string) string {
	if !m.compactModifiers {
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
