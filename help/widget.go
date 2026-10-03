package help

import (
	"cmp"
	"strings"

	"github.com/ayn2op/tview/layout"
	"github.com/gdamore/tcell/v3"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/keybind"
	"github.com/ayn2op/tview/richtext"
	"github.com/rivo/uniseg"
)

const (
	defaultShortSeparator = " • "
	defaultFullSeparator  = "    "
	defaultEllipsis       = "…"
)

type KeyMap interface {
	// ShortHelp returns keybinds for single-line help.
	ShortHelp() []keybind.Keybind
	// FullHelp returns keybind groups, where each top-level entry is a column.
	FullHelp() [][]keybind.Keybind
}

// Widget draws the keybinds of a KeyMap on one line, or in columns when showing all.
type Widget struct {
	shortKeyStyle, shortDescStyle           tview.Style
	fullKeyStyle, fullDescStyle             tview.Style
	shortSeparatorStyle, fullSeparatorStyle tview.Style
	ellipsisStyle                           tview.Style
	keyMap                                  KeyMap
	showAll                                 bool
	compactModifiers                        bool
	shortSeparator                          string
	fullSeparator                           string
	ellipsis                                string
}

var _ tview.Element = Widget{}

// New returns help for keyMap showing its short help.
func New(keyMap KeyMap) Widget {
	dim := tcell.StyleDefault.Dim(true)
	return Widget{
		keyMap:              keyMap,
		shortKeyStyle:       dim,
		fullKeyStyle:        dim,
		shortSeparatorStyle: dim,
		fullSeparatorStyle:  dim,
		ellipsisStyle:       dim,
		shortSeparator:      defaultShortSeparator,
		fullSeparator:       defaultFullSeparator,
		ellipsis:            defaultEllipsis,
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

// ShortKeyStyle sets the style of keys in the short help.
func (w Widget) ShortKeyStyle(style tview.Style) Widget {
	w.shortKeyStyle = style
	return w
}

// ShortDescStyle sets the style of descriptions in the short help.
func (w Widget) ShortDescStyle(style tview.Style) Widget {
	w.shortDescStyle = style
	return w
}

// FullKeyStyle sets the style of keys in the full help.
func (w Widget) FullKeyStyle(style tview.Style) Widget {
	w.fullKeyStyle = style
	return w
}

// FullDescStyle sets the style of descriptions in the full help.
func (w Widget) FullDescStyle(style tview.Style) Widget {
	w.fullDescStyle = style
	return w
}

// ShortSeparatorStyle sets the style of the separator in the short help.
func (w Widget) ShortSeparatorStyle(style tview.Style) Widget {
	w.shortSeparatorStyle = style
	return w
}

// FullSeparatorStyle sets the style of the separator in the full help.
func (w Widget) FullSeparatorStyle(style tview.Style) Widget {
	w.fullSeparatorStyle = style
	return w
}

// EllipsisStyle sets the style of the ellipsis.
func (w Widget) EllipsisStyle(style tview.Style) Widget {
	w.ellipsisStyle = style
	return w
}

// Size returns Fill, as the help takes the area its parent gives it. Layout reports the rows it needs.
func (Widget) Size() (width, height layout.Length) {
	return layout.Fill, layout.Fill
}

// Layout returns the size of the help within limits, as tall as the rows it takes at the width of limits, where a width of 0 is unlimited.
func (w Widget) Layout(limits layout.Limits) layout.Size {
	return layout.Sized(limits, layout.Fill, layout.Shrink, func(limits layout.Limits) layout.Size {
		if !w.showAll || w.keyMap == nil {
			return layout.Size{Height: 1}
		}
		return layout.Size{Height: len(w.fullHelpSegments(w.keyMap.FullHelp(), limits.Max.Width))}
	})
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

func (w Widget) shortHelpSegments(bindings []keybind.Keybind, maxWidth int) richtext.Line {
	items := make([]richtext.Line, 0, len(bindings))
	for _, kb := range bindings {
		hp := kb.Help()
		item := shortItemSegments(w.formatKey(hp.Key), hp.Desc, w.shortKeyStyle, w.shortDescStyle)
		if len(item) == 0 {
			continue
		}
		items = append(items, item)
	}
	if len(items) == 0 {
		return nil
	}

	sep := richtext.Segment{Text: cmp.Or(w.shortSeparator, " "), Style: w.shortSeparatorStyle}

	out := items[0].Clone()
	for i := 1; i < len(items); i++ {
		candidate := append(out.Clone(), sep)
		candidate = append(candidate, items[i]...)
		if maxWidth > 0 && candidate.Width() > maxWidth {
			return append(out, w.truncationTail(out, maxWidth)...)
		}
		out = candidate
	}

	if maxWidth > 0 && out.Width() > maxWidth {
		return nil
	}
	return out
}

func (w Widget) fullHelpSegments(groups [][]keybind.Keybind, maxWidth int) []richtext.Line {
	sep := richtext.Segment{Text: cmp.Or(w.fullSeparator, " "), Style: w.fullSeparatorStyle}
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
			return []richtext.Line{{{Text: w.ellipsis, Style: w.ellipsisStyle}}}
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
				cell = append(cell.Clone(), richtext.Segment{Text: strings.Repeat(" ", pad), Style: w.fullDescStyle})
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
		rows[i] = richtext.Line{{Text: h.Key + strings.Repeat(" ", keyWidth-uniseg.StringWidth(h.Key)), Style: w.fullKeyStyle}}
		if h.Key != "" && h.Desc != "" {
			rows[i] = append(rows[i], richtext.Segment{Text: " ", Style: w.fullDescStyle})
		}
		if h.Desc != "" {
			rows[i] = append(rows[i], richtext.Segment{Text: h.Desc, Style: w.fullDescStyle})
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
	tail := richtext.Line{{Text: " " + w.ellipsis, Style: w.ellipsisStyle}}
	if current.Width()+tail.Width() <= maxWidth {
		return tail
	}
	return nil
}

func (w Widget) drawSegments(screen tview.Screen, x, y, width int, segments richtext.Line) {
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

func shortItemSegments(key, desc string, keyStyle, descStyle tview.Style) richtext.Line {
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
