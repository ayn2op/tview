package help

import (
	"strings"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/keybind"
	"github.com/ayn2op/tview/text"
	"github.com/gdamore/tcell/v3"
	"github.com/rivo/uniseg"
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

// ShowAll returns whether full help mode is enabled.
func (m *Model) ShowAll() bool {
	return m.showAll
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

func (m *Model) View(screen tcell.Screen) {
	m.Box.View(screen)

	if m.keyMap == nil {
		return
	}

	x, y, width, height := m.InnerRect()

	var lines []text.Line
	if m.showAll {
		lines = m.fullHelpSegments(m.keyMap.FullHelp(), width)
	} else {
		lines = []text.Line{m.shortHelpSegments(m.keyMap.ShortHelp(), width)}
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
			b.WriteString(s.Text)
		}
		lines = append(lines, b.String())
	}
	return lines
}

func (m *Model) shortHelpSegments(bindings []keybind.Keybind, maxWidth int) text.Line {
	items := make([]text.Line, 0, len(bindings))
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
	sep := text.Segment{Text: sepText, Style: m.styles.ShortSeparator}

	out := items[0].Clone()
	for i := 1; i < len(items); i++ {
		candidate := append(out.Clone(), sep)
		candidate = append(candidate, items[i]...)
		if maxWidth > 0 && candidate.Width() > maxWidth {
			tail := m.truncationTail(out, maxWidth)
			if len(tail) > 0 {
				out = append(out, tail...)
			}
			return out
		}
		out = candidate
	}

	if maxWidth > 0 && out.Width() > maxWidth {
		return nil
	}
	return out
}

func (m *Model) fullHelpSegments(groups [][]keybind.Keybind, maxWidth int) []text.Line {
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
			col.keyW = max(col.keyW, uniseg.StringWidth(keyText))
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
			w += uniseg.StringWidth(e.desc)
			col.colW = max(col.colW, w)
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
	sepW := uniseg.StringWidth(sepText)

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
		return []text.Line{{{Text: m.ellipsis, Style: m.styles.Ellipsis}}}
	}
	truncated := included < len(columns)

	maxRows := 0
	for i := range included {
		maxRows = max(maxRows, len(columns[i].entries))
	}

	lines := make([]text.Line, 0, maxRows)
	for row := range maxRows {
		line := make(text.Line, 0, included*4)
		for col := range included {
			if col > 0 {
				line = append(line, text.Segment{Text: sepText, Style: m.styles.FullSeparator})
			}

			c := columns[col]
			cell := make(text.Line, 0, 4)
			if row >= len(c.entries) {
				// Empty rows still occupy full column width so the following separators do not drift.
				cell = append(cell, text.Segment{Text: strings.Repeat(" ", c.colW), Style: m.styles.FullDesc})
				line = append(line, cell...)
				continue
			}

			e := c.entries[row]
			keyPad := c.keyW - uniseg.StringWidth(e.key)
			if e.key != "" {
				cell = append(cell, text.Segment{Text: e.key, Style: m.styles.FullKey})
			}
			if keyPad > 0 {
				cell = append(cell, text.Segment{Text: strings.Repeat(" ", keyPad), Style: m.styles.FullKey})
			}
			if e.key != "" && e.desc != "" {
				cell = append(cell, text.Segment{Text: " ", Style: m.styles.FullDesc})
			}
			if e.desc != "" {
				cell = append(cell, text.Segment{Text: e.desc, Style: m.styles.FullDesc})
			}

			// Every non-last column is padded to fixed width so row-specific content lengths do not shift separators.
			if col < included-1 {
				cellWidth := cell.Width()
				if pad := c.colW - cellWidth; pad > 0 {
					cell = append(cell, text.Segment{Text: strings.Repeat(" ", pad), Style: m.styles.FullDesc})
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

func (m *Model) truncationTail(current text.Line, maxWidth int) text.Line {
	if maxWidth <= 0 || m.ellipsis == "" {
		return nil
	}
	// We only add an ellipsis when it fully fits because clipping looks broken in narrow widths.
	tail := text.Line{
		{Text: " ", Style: m.styles.Ellipsis},
		{Text: m.ellipsis, Style: m.styles.Ellipsis},
	}
	if current.Width()+tail.Width() <= maxWidth {
		return tail
	}
	return nil
}

func (m *Model) drawSegments(screen tcell.Screen, x, y, width int, segments text.Line) {
	if width <= 0 || len(segments) == 0 {
		return
	}

	cursor := x
	remaining := width
	for _, s := range segments {
		if s.Text == "" || remaining <= 0 {
			continue
		}
		_, printedWidth := tview.PrintWithStyle(screen, s.Text, cursor, y, remaining, tview.AlignmentLeft, s.Style)
		cursor += printedWidth
		remaining -= printedWidth
	}
}

func shortItemSegments(key, desc string, keyStyle, descStyle tcell.Style) text.Line {
	switch {
	case key == "" && desc == "":
		return nil
	case key == "":
		return text.Line{{Text: desc, Style: descStyle}}
	case desc == "":
		return text.Line{{Text: key, Style: keyStyle}}
	default:
		return text.Line{{Text: key, Style: keyStyle}, {Text: " ", Style: descStyle}, {Text: desc, Style: descStyle}}
	}
}

func (m *Model) formatKey(key string) string {
	if !m.compactModifiers {
		return key
	}
	return compactModifierReplacer.Replace(key)
}
