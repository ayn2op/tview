package help

import (
	"cmp"
	"strings"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/keybind"
	"github.com/ayn2op/tview/richtext"
	"github.com/gdamore/tcell/v3"
	"github.com/rivo/uniseg"
)

type KeyMap interface {
	// ShortHelp returns keybinds for single-line help.
	ShortHelp() []keybind.Keybind
	// FullHelp returns keybind groups, where each top-level entry is a column.
	FullHelp() [][]keybind.Keybind
}

// Widget draws the keybinds of a KeyMap on one line, or in columns when showing all.
type Widget struct {
	styles           Styles
	keyMap           KeyMap
	showAll          bool
	compactModifiers bool
	shortSeparator   string
	fullSeparator    string
	ellipsis         string
}

var _ tview.Element = Widget{}

// New returns help for keyMap showing its short help.
func New(keyMap KeyMap) Widget {
	return Widget{
		keyMap:         keyMap,
		styles:         DefaultStyles(),
		shortSeparator: " • ",
		fullSeparator:  "    ",
		ellipsis:       "…",
	}
}

// ShowAll sets whether the full help is shown in columns instead of the short help.
func (w Widget) ShowAll(showAll bool) Widget {
	w.showAll = showAll
	return w
}

// CompactModifiers sets whether modifiers are shown as symbols, such as ^ for ctrl.
func (w Widget) CompactModifiers(compact bool) Widget {
	w.compactModifiers = compact
	return w
}

// ShortSeparator sets the separator between keybinds in the short help.
func (w Widget) ShortSeparator(separator string) Widget {
	w.shortSeparator = separator
	return w
}

// FullSeparator sets the separator between columns in the full help.
func (w Widget) FullSeparator(separator string) Widget {
	w.fullSeparator = separator
	return w
}

// Ellipsis sets the symbol shown when keybinds are cut off for width.
func (w Widget) Ellipsis(ellipsis string) Widget {
	w.ellipsis = ellipsis
	return w
}

// Styles sets the styles of keys, descriptions, separators, and the ellipsis.
func (w Widget) Styles(styles Styles) Widget {
	w.styles = styles
	return w
}

// Draw draws the help lines that fit in area.
func (w Widget) Draw(screen tview.Screen, area tview.Rectangle) {
	if w.keyMap == nil {
		return
	}

	x, y, width, height := area.X, area.Y, area.Width, area.Height

	var lines []richtext.Line
	if w.showAll {
		lines = w.fullHelpSegments(w.keyMap.FullHelp(), width)
	} else {
		lines = []richtext.Line{w.shortHelpSegments(w.keyMap.ShortHelp(), width)}
	}

	for row := 0; row < len(lines) && row < height; row++ {
		w.drawSegments(screen, x, y+row, width, lines[row])
	}
}

// Handle passes msg through unchanged.
func (Widget) Handle(msg tview.Msg, area tview.Rectangle) tview.Msg { return msg }

// Rows returns the number of rows the help takes at width, where 0 means unlimited.
func (w Widget) Rows(width int) int {
	if !w.showAll || w.keyMap == nil {
		return 1
	}
	return len(w.fullHelpSegments(w.keyMap.FullHelp(), width))
}

func (w Widget) shortHelpSegments(bindings []keybind.Keybind, maxWidth int) richtext.Line {
	items := make([]richtext.Line, 0, len(bindings))
	for _, kb := range bindings {
		hp := kb.Help()
		item := shortItemSegments(w.formatKey(hp.Key), hp.Desc, w.styles.ShortKey, w.styles.ShortDesc)
		if len(item) == 0 {
			continue
		}
		items = append(items, item)
	}
	if len(items) == 0 {
		return nil
	}

	sepText := w.shortSeparator
	if sepText == "" {
		sepText = " "
	}
	sep := richtext.Segment{Text: sepText, Style: w.styles.ShortSeparator}

	out := items[0].Clone()
	for i := 1; i < len(items); i++ {
		candidate := append(out.Clone(), sep)
		candidate = append(candidate, items[i]...)
		if maxWidth > 0 && candidate.Width() > maxWidth {
			tail := w.truncationTail(out, maxWidth)
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

func (w Widget) fullHelpSegments(groups [][]keybind.Keybind, maxWidth int) []richtext.Line {
	sep := richtext.Segment{Text: cmp.Or(w.fullSeparator, " "), Style: w.styles.FullSeparator}
	var columns [][]richtext.Line
	var widths []int
	width, truncated := 0, false
	// Columns are included left to right until the next one would overflow maxWidth.
	for _, group := range groups {
		column, columnWidth := w.fullColumn(group)
		if len(column) == 0 {
			continue
		}
		next := columnWidth
		if len(columns) > 0 {
			next += uniseg.StringWidth(sep.Text)
		}
		if maxWidth > 0 && width+next > maxWidth {
			truncated = true
			break
		}
		columns, widths = append(columns, column), append(widths, columnWidth)
		width += next
	}
	if len(columns) == 0 {
		if truncated {
			return []richtext.Line{{{Text: w.ellipsis, Style: w.styles.Ellipsis}}}
		}
		return nil
	}

	rows := 0
	for _, column := range columns {
		rows = max(rows, len(column))
	}
	var lines []richtext.Line
	for row := range rows {
		var line richtext.Line
		for i, column := range columns {
			if i > 0 {
				line = append(line, sep)
			}
			var cell richtext.Line
			if row < len(column) {
				cell = column[row]
			}
			// Every column but the last is padded to its width so the separators line up.
			if pad := widths[i] - cell.Width(); pad > 0 && (i < len(columns)-1 || row >= len(column)) {
				cell = append(cell.Clone(), richtext.Segment{Text: strings.Repeat(" ", pad), Style: w.styles.FullDesc})
			}
			line = append(line, cell...)
		}
		lines = append(lines, line)
	}
	if truncated {
		lines[0] = append(lines[0], w.truncationTail(lines[0], maxWidth)...)
	}
	return lines
}

// fullColumn returns the rows of group, each key padded to the widest one and followed by its description, and the width of the widest row.
func (w Widget) fullColumn(group []keybind.Keybind) ([]richtext.Line, int) {
	var helps []keybind.Help
	keyWidth := 0
	for _, kb := range group {
		if h := kb.Help(); h.Key != "" || h.Desc != "" {
			h.Key = w.formatKey(h.Key)
			helps = append(helps, h)
			keyWidth = max(keyWidth, uniseg.StringWidth(h.Key))
		}
	}
	rows := make([]richtext.Line, len(helps))
	width := 0
	for i, h := range helps {
		rows[i] = richtext.Line{{Text: h.Key + strings.Repeat(" ", keyWidth-uniseg.StringWidth(h.Key)), Style: w.styles.FullKey}}
		if h.Key != "" && h.Desc != "" {
			rows[i] = append(rows[i], richtext.Segment{Text: " ", Style: w.styles.FullDesc})
		}
		if h.Desc != "" {
			rows[i] = append(rows[i], richtext.Segment{Text: h.Desc, Style: w.styles.FullDesc})
		}
		width = max(width, rows[i].Width())
	}
	return rows, width
}

func (w Widget) truncationTail(current richtext.Line, maxWidth int) richtext.Line {
	if maxWidth <= 0 || w.ellipsis == "" {
		return nil
	}
	// We only add an ellipsis when it fully fits because clipping looks broken in narrow widths.
	tail := richtext.Line{
		{Text: " ", Style: w.styles.Ellipsis},
		{Text: w.ellipsis, Style: w.styles.Ellipsis},
	}
	if current.Width()+tail.Width() <= maxWidth {
		return tail
	}
	return nil
}

func (w Widget) drawSegments(screen tview.Screen, x, y, width int, segments richtext.Line) {
	if width <= 0 || len(segments) == 0 {
		return
	}

	cursor := x
	remaining := width
	for _, s := range segments {
		if s.Text == "" || remaining <= 0 {
			continue
		}
		printedWidth := tview.Print(screen, s.Text, cursor, y, remaining, tview.AlignmentLeft, s.Style)
		cursor += printedWidth
		remaining -= printedWidth
	}
}

func shortItemSegments(key, desc string, keyStyle, descStyle tcell.Style) richtext.Line {
	switch {
	case key == "" && desc == "":
		return nil
	case key == "":
		return richtext.Line{{Text: desc, Style: descStyle}}
	case desc == "":
		return richtext.Line{{Text: key, Style: keyStyle}}
	default:
		return richtext.Line{{Text: key, Style: keyStyle}, {Text: " ", Style: descStyle}, {Text: desc, Style: descStyle}}
	}
}

func (w Widget) formatKey(key string) string {
	if !w.compactModifiers {
		return key
	}
	return compactModifierReplacer.Replace(key)
}
