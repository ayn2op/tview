package textarea

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/ayn2op/tview/internal/grapheme"
)

// isWord reports whether the grapheme cluster starting at offset in value is part of a word, as readline's are: letters and digits.
func isWord(value string, offset int) bool {
	r, _ := utf8.DecodeRuneInString(value[offset:])
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

// isNotSpace reports whether the grapheme cluster starting at offset in value is not whitespace.
func isNotSpace(value string, offset int) bool {
	r, _ := utf8.DecodeRuneInString(value[offset:])
	return !unicode.IsSpace(r)
}

// runStart returns the start of the last run of grapheme clusters that in matches before cursor, or 0 if there is none.
func runStart(value string, cursor int, in func(string, int) bool) int {
	start, inRun := 0, false
	for offset := 0; offset < cursor; offset = grapheme.Next(value, offset) {
		switch {
		case !in(value, offset):
			inRun = false
		case !inRun:
			start, inRun = offset, true
		}
	}
	return start
}

// runEnd returns the end of the first run of grapheme clusters that in matches after cursor, or the end of value if there is none.
func runEnd(value string, cursor int, in func(string, int) bool) int {
	offset := cursor
	for offset < len(value) && !in(value, offset) {
		offset = grapheme.Next(value, offset)
	}
	for offset < len(value) && in(value, offset) {
		offset = grapheme.Next(value, offset)
	}
	return offset
}

// lineStart returns the start of the line holding cursor, between newlines rather than as wrapped.
func lineStart(value string, cursor int) int {
	return strings.LastIndexByte(value[:cursor], '\n') + 1
}

// lineEnd returns the end of the line holding cursor, between newlines rather than as wrapped.
func lineEnd(value string, cursor int) int {
	if end := strings.IndexByte(value[cursor:], '\n'); end >= 0 {
		return cursor + end
	}
	return len(value)
}
