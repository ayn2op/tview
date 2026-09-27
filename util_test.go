package tview_test

import (
	"testing"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/gdamore/tcell/v3"
)

func TestPrint(t *testing.T) {
	for _, tt := range []struct {
		name      string
		text      string
		width     int
		alignment tview.Alignment
		want      string
		drawn     int
	}{
		{"left", "ab", 6, tview.AlignmentLeft, "ab    ", 2},
		{"center", "ab", 6, tview.AlignmentCenter, "  ab  ", 2},
		{"right", "ab", 6, tview.AlignmentRight, "    ab", 2},
		{"cut at the end", "abcdefgh", 6, tview.AlignmentLeft, "abcdef", 6},
		{"cut at the start", "abcdefgh", 6, tview.AlignmentRight, "cdefgh", 6},
		{"cut at both ends", "abcdefgh", 6, tview.AlignmentCenter, "bcdefg", 6},
		{"wide cluster that does not fit", "abcde👍", 6, tview.AlignmentLeft, "abcde ", 5},
	} {
		t.Run(tt.name, func(t *testing.T) {
			screen := screentest.New(t, 6, 1)
			if drawn := tview.Print(screen, tt.text, 0, 0, tt.width, tt.alignment, tcell.StyleDefault); drawn != tt.drawn {
				t.Fatalf("drawn = %d, want %d", drawn, tt.drawn)
			}
			if got := screentest.Row(screen, 0, 6); got != tt.want {
				t.Fatalf("row = %q, want %q", got, tt.want)
			}
		})
	}
}
