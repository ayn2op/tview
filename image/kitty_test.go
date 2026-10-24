package image

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"image"
	"io"
	"math/rand/v2"
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
		// Noise does not compress, so it needs several chunks.
		img := image.NewGray(image.Rect(0, 0, 100, 100))
		rng := rand.New(rand.NewPCG(1, 2))
		for i := range img.Pix {
			img.Pix[i] = uint8(rng.Uint32())
		}
		got, err := New(img).Kitty(1).transmission()
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(got, "\x1b_Ga=T,U=1,f=32,o=z,s=100,v=100,q=2,i=8,c=100,r=50,m=1;") || !strings.Contains(got, "\x1b_Gm=0;") || strings.Count(got, "m=1;") < 2 {
			t.Fatalf("transmission = %.80q...", got)
		}

		var want []byte
		for _, gray := range img.Pix {
			want = append(want, gray, gray, gray, 0xff)
		}
		if pixels := sentPixels(t, got); !bytes.Equal(pixels, want) {
			t.Fatal("the transmission does not hold the image's RGBA pixels")
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

// sentPixels returns the pixels a transmission carries: its chunks' data joined, decoded from base64 and inflated.
func sentPixels(t *testing.T, transmission string) []byte {
	t.Helper()
	var data strings.Builder
	for chunk := range strings.SplitSeq(transmission, "\x1b\\") {
		if _, payload, ok := strings.Cut(chunk, ";"); ok {
			data.WriteString(payload)
		}
	}

	compressed, err := base64.StdEncoding.DecodeString(data.String())
	if err != nil {
		t.Fatal(err)
	}
	reader, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatal(err)
	}
	pixels, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	return pixels
}
