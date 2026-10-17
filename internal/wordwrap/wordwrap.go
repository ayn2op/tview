// Package wordwrap splits text into lines between words.
package wordwrap

import (
	"strings"
	"unicode/utf8"

	"github.com/rockorager/go-uucode"
)

// Split splits text into lines no wider than width, breaking where a line may break. Spaces at the end of a line hang past its edge instead of wrapping. A word wider than width is broken at any character if glyph is set, and otherwise runs past the end of its line.
func Split(text string, width int, glyph bool) []string {
	var lines []string
	start, lineWidth := 0, 0 // Where the current line starts in text, and how wide it is.
	it := uucode.NewLineIterator(text)
	for segment, ok := it.Next(); ok; segment, ok = it.Next() {
		word := text[segment.Start:segment.End]
		visible := strings.TrimRight(word, " ")
		wordWidth, visibleWidth := uucode.StringWidth(word), uucode.StringWidth(visible)
		if lineWidth > 0 && lineWidth+visibleWidth > width {
			lines = append(lines, text[start:segment.Start])
			start, lineWidth = segment.Start, 0
		}

		if glyph && visibleWidth > width {
			clusters := uucode.NewGraphemeWidthIterator(visible)
			for cluster, ok := clusters.Next(); ok; cluster, ok = clusters.Next() {
				if lineWidth > 0 && lineWidth+cluster.Width > width {
					lines = append(lines, text[start:segment.Start+cluster.Start])
					start, lineWidth = segment.Start+cluster.Start, 0
				}
				lineWidth += cluster.Width
			}
			lineWidth += wordWidth - visibleWidth
		} else {
			lineWidth += wordWidth
		}

		// The last segment ends in a mandatory break even without a line break at its end.
		if segment.Break == uucode.LineMustBreak && endsWithLineBreak(word) {
			lines = append(lines, strings.TrimRight(text[start:segment.End], "\n\r"))
			start, lineWidth = segment.End, 0
		}
	}
	return append(lines, text[start:])
}

// endsWithLineBreak reports whether s ends with a character that ends a line.
func endsWithLineBreak(s string) bool {
	r, _ := utf8.DecodeLastRuneInString(s)
	switch uucode.LineBreak(r) {
	case uucode.LineBreakBK, uucode.LineBreakCR, uucode.LineBreakLF, uucode.LineBreakNL:
		return true
	}
	return false
}
