package image

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"errors"
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
		img := image.NewGray(image.Rect(0, 0, 100, 100))
		rng := rand.New(rand.NewPCG(1, 2))
		for i := range img.Pix {
			img.Pix[i] = uint8(rng.Uint32())
		}
		got, err := New(img).Kitty(1).transmission()
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(got, "\x1b_Ga=T,U=1,f=32,s=100,v=100,q=2,i=8,c=100,r=50,m=1;") || !strings.Contains(got, "\x1b_Gm=0;") || strings.Count(got, "m=1;") < 2 {
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
	t.Run("transmits with the compressor", func(t *testing.T) {
		singleColor := image.NewNRGBA(image.Rect(0, 0, 100, 100))
		got, err := New(singleColor).Kitty(1).Compressor(newBestCompressor).transmission()
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(got, "o=z,") || !bytes.Equal(sentPixels(t, got), pixels(singleColor)) {
			t.Fatal("the transmission does not hold the image's compressed RGBA pixels")
		}
	})
	t.Run("transmits uncompressed what does not shrink", func(t *testing.T) {
		twoPixels := column(red, red)
		got, err := New(twoPixels).Kitty(1).Compressor(newBestCompressor).transmission()
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(got, "o=z,") || !bytes.Equal(sentPixels(t, got), pixels(twoPixels)) {
			t.Fatal("the transmission does not hold the image's uncompressed RGBA pixels")
		}
	})
	t.Run("returns the compressor's error", func(t *testing.T) {
		want := errors.New("compressor failed")
		_, err := New(column(red)).Kitty(1).Compressor(func(io.Writer) (io.WriteCloser, error) {
			return nil, want
		}).transmission()
		if !errors.Is(err, want) {
			t.Fatalf("err = %v, want %v", err, want)
		}
	})
}

func BenchmarkWidgetTransmission(b *testing.B) {
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

func newBestCompressor(w io.Writer) (io.WriteCloser, error) {
	return zlib.NewWriterLevel(w, zlib.BestCompression)
}

// sentPixels returns the pixels a transmission carries: its chunks' data joined, decoded from base64 and inflated if compressed.
func sentPixels(t *testing.T, transmission string) []byte {
	t.Helper()
	var data strings.Builder
	for chunk := range strings.SplitSeq(transmission, "\x1b\\") {
		if _, payload, ok := strings.Cut(chunk, ";"); ok {
			data.WriteString(payload)
		}
	}

	sent, err := base64.StdEncoding.DecodeString(data.String())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(transmission, "o=z,") {
		return sent
	}

	reader, err := zlib.NewReader(bytes.NewReader(sent))
	if err != nil {
		t.Fatal(err)
	}
	pixels, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	return pixels
}
