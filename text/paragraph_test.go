package text

import (
	"slices"
	"testing"
)

func TestParagraphUpdate(t *testing.T) {
	spans := []Span{{Text: "one two three"}}
	var p Paragraph
	first := p.update(spans, 9, WrappingWord)

	for _, tt := range []struct {
		name    string
		spans   []Span
		width   int
		relayed bool
	}{
		{"the same width keeps the lines", spans, 9, false},
		{"a narrower width they still fit keeps them", spans, Width(first[0]), false},
		{"a width they do not fit lays them out again", spans, 4, true},
		{"other spans lay them out again", []Span{{Text: "four"}}, 9, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := Paragraph{}
			before := p.update(spans, 9, WrappingWord)
			got := p.update(tt.spans, tt.width, WrappingWord)
			if kept := &got[0] == &before[0]; kept == tt.relayed {
				t.Fatalf("kept the lines = %v, want %v", kept, !tt.relayed)
			}
			if want := Lines(tt.spans, tt.width, WrappingWord); !slices.EqualFunc(got, want, slices.Equal) {
				t.Fatalf("lines = %v, want %v", got, want)
			}
		})
	}
}
