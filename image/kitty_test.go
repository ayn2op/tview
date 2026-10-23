package image

import (
	"image"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/ayn2op/tview"
	"github.com/ayn2op/tview/internal/screentest"
	"github.com/gdamore/tcell/v3/color"
)

func TestWidgetKitty(t *testing.T) {
	t.Run("draws placeholders", func(t *testing.T) {
		screen := screentest.New(t, 2, 1)
		New(column(red, red)).Kitty(257).Draw(screen, tview.Rectangle{Width: 2, Height: 1})
		// Image 257 is palette color 16 with a most significant byte of 1.
		str, style, _ := screen.Get(0, 0)
		if want := string([]rune{placeholder, diacritics[0], diacritics[0], diacritics[1]}); str != want || style.GetForeground() != color.PaletteColor(16) {
			t.Fatalf("cell = %q %v, want %q in palette color 16", str, style.GetForeground(), want)
		}
	})
	t.Run("transmits without keeping the pixels", func(t *testing.T) {
		widget := New(column(red, red)).Width(1).Kitty(1)
		want, err := widget.transmission()
		if err != nil {
			t.Fatal(err)
		}
		sent, cmd := widget.Transmit()
		if _, ok := sent.src.(image.Rectangle); !ok {
			t.Fatalf("Transmit kept %T", sent.src)
		}
		if msg := cmd(); msg != tview.WriteRaw(want)() {
			t.Fatal("Transmit did not send the image's pixels")
		}
		sentWidth, sentHeight := sent.Size()
		width, height := widget.Size()
		if sentWidth != width || sentHeight != height || !slices.Equal(sent.Lines()[0], widget.Lines()[0]) {
			t.Fatal("the sent widget is drawn differently")
		}
	})
	t.Run("deletes the image", func(t *testing.T) {
		msg := New(column(red)).Kitty(257).Delete()()
		if want := tview.WriteRaw("\x1b_Ga=d,d=I,q=2,i=16777232\x1b\\")(); msg != want {
			t.Fatalf("Delete = %v, want %v", msg, want)
		}
	})
	t.Run("deletes nothing without kitty", func(t *testing.T) {
		if New(column(red)).Delete() != nil {
			t.Fatal("Delete is not nil")
		}
	})
	t.Run("transmits in chunks", func(t *testing.T) {
		// Noise does not compress, so its PNG needs several chunks.
		img := image.NewGray(image.Rect(0, 0, 100, 100))
		rng := rand.New(rand.NewPCG(1, 2))
		for i := range img.Pix {
			img.Pix[i] = uint8(rng.Uint32())
		}
		got, err := New(img).Kitty(1).transmission()
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(got, "\x1b_Ga=T,U=1,f=100,q=2,i=8,c=100,r=50,m=1;") || !strings.Contains(got, "\x1b_Gm=0;") || strings.Count(got, "m=1;") < 2 {
			t.Fatalf("transmission = %.80q...", got)
		}
	})
}

func BenchmarkWidgetTransmission(b *testing.B) {
	// Noise does not compress, so its PNG is as large as the image allows.
	for _, side := range []int{40, 400} {
		b.Run(strconv.Itoa(side)+"px", func(b *testing.B) {
			img := image.NewNRGBA(image.Rect(0, 0, side, side))
			rng := rand.New(rand.NewPCG(1, 2))
			for i := range img.Pix {
				img.Pix[i] = uint8(rng.Uint32())
			}
			widget := New(img).Kitty(1)
			b.ReportAllocs()
			for b.Loop() {
				if _, err := widget.transmission(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
