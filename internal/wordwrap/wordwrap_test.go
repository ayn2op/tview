package wordwrap

import (
	"slices"
	"testing"
)

func TestSplit(t *testing.T) {
	for _, tc := range []struct {
		name  string
		text  string
		width int
		glyph bool
		want  []string
	}{
		{"empty", "", 3, false, []string{""}},
		{"breaks between words", "ab cd ef", 6, false, []string{"ab cd ", "ef"}},
		{"breaks between ideographs", "界界界", 4, false, []string{"界界", "界"}},
		{"ends a line at a line break", "a\r\nb", 5, false, []string{"a", "b"}},
		{"lets a long word run past the edge", "abcdef g", 3, false, []string{"abcdef ", "g"}},
		{"breaks a long word at any character", "abcdef g", 3, true, []string{"abc", "def ", "g"}},
		{"hangs spaces past the edge", "abc   def", 3, false, []string{"abc   ", "def"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := Split(tc.text, tc.width, tc.glyph); !slices.Equal(got, tc.want) {
				t.Fatalf("Split(%q, %d, %t) = %q, want %q", tc.text, tc.width, tc.glyph, got, tc.want)
			}
		})
	}
}
