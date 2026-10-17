package text

import (
	"slices"
	"testing"

	"github.com/gdamore/tcell/v3"
)

func TestWidth(t *testing.T) {
	if got := Width([]Span{{Text: "ab"}, {Text: "界"}}); got != 4 {
		t.Fatalf("width = %d, want 4", got)
	}
}

func TestLines(t *testing.T) {
	bold := tcell.StyleDefault.Bold(true)
	for _, tc := range []struct {
		name     string
		spans    []Span
		width    int
		wrapping Wrapping
		want     [][]Span
	}{
		{"splits at line breaks", []Span{{Text: "one\ntwo"}}, 0, WrappingNone, [][]Span{{{Text: "one"}}, {{Text: "two"}}}},
		{"keeps empty lines", []Span{{Text: "a\n\nb"}}, 0, WrappingNone, [][]Span{{{Text: "a"}}, nil, {{Text: "b"}}}},
		{"splits once at a carriage return and line feed", []Span{{Text: "a\r\nb"}}, 0, WrappingNone, [][]Span{{{Text: "a"}}, {{Text: "b"}}}},
		{"turns tabs into spaces", []Span{{Text: "a\tb"}}, 0, WrappingNone, [][]Span{{{Text: "a b"}}}},
		{"breaks at any character", []Span{{Text: "ab界c"}}, 3, WrappingGlyph, [][]Span{{{Text: "ab"}}, {{Text: "界c"}}}},
		{
			"breaks between words in the styles of the spans",
			[]Span{{Text: "hello "}, {Text: "big world", Style: bold}}, 10, WrappingWord,
			[][]Span{{{Text: "hello "}, {Text: "big", Style: bold}}, {{Text: "world", Style: bold}}},
		},
		{"lets a long word run past the edge", []Span{{Text: "abcdef g"}}, 3, WrappingWord, [][]Span{{{Text: "abcdef"}}, {{Text: "g"}}}},
		{"breaks a long word at any character", []Span{{Text: "abcdef g"}}, 3, WrappingWordOrGlyph, [][]Span{{{Text: "abc"}}, {{Text: "def"}}, {{Text: "g"}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := Lines(tc.spans, tc.width, tc.wrapping); !slices.EqualFunc(got, tc.want, slices.Equal[[]Span]) {
				t.Fatalf("lines = %v, want %v", got, tc.want)
			}
		})
	}
}
