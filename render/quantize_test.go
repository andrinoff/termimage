package render

import (
	"image"
	"image/color"
	"testing"
)

func TestMedianCut_PaletteSize(t *testing.T) {
	// 4-color checkerboard; palette should converge on 4 colors when n>=4.
	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	colors := []color.RGBA{
		{R: 255, A: 255},
		{G: 255, A: 255},
		{B: 255, A: 255},
		{R: 255, G: 255, B: 255, A: 255},
	}
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, colors[(y*4+x)%4])
		}
	}

	pal := medianCut(img, 16)
	if len(pal) == 0 {
		t.Fatal("expected non-empty palette")
	}
	if len(pal) > 16 {
		t.Errorf("palette exceeds requested size: got %d", len(pal))
	}
}

func TestMedianCut_SkipsTransparent(t *testing.T) {
	// All-transparent image -> palette built from 0 pixels.
	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	pal := medianCut(img, 4)
	// 0 pixels means one empty bucket -> palette of 1 averaged-zero color.
	if len(pal) != 1 {
		t.Errorf("expected palette of 1 for all-transparent input, got %d", len(pal))
	}
}

func TestNearestColor_ExactMatch(t *testing.T) {
	palette := []color.Color{
		color.RGBA{R: 255, A: 255},
		color.RGBA{G: 255, A: 255},
		color.RGBA{B: 255, A: 255},
	}
	if idx := nearestColor(palette, color.RGBA{G: 255, A: 255}); idx != 1 {
		t.Errorf("expected index 1 for pure green, got %d", idx)
	}
	if idx := nearestColor(palette, color.RGBA{B: 255, A: 255}); idx != 2 {
		t.Errorf("expected index 2 for pure blue, got %d", idx)
	}
}

func TestNearestColor_ClosestMatch(t *testing.T) {
	palette := []color.Color{
		color.RGBA{A: 255},                         // black
		color.RGBA{R: 255, G: 255, B: 255, A: 255}, // white
	}
	if idx := nearestColor(palette, color.RGBA{R: 20, G: 20, B: 20, A: 255}); idx != 0 {
		t.Errorf("dark grey should map to black (0), got %d", idx)
	}
	if idx := nearestColor(palette, color.RGBA{R: 230, G: 230, B: 230, A: 255}); idx != 1 {
		t.Errorf("light grey should map to white (1), got %d", idx)
	}
}

func TestAverage_EmptyReturnsZero(t *testing.T) {
	c := average(nil)
	r, g, b, a := c.RGBA()
	if r != 0 || g != 0 || b != 0 || a != 0 {
		t.Errorf("average(nil) = (%d,%d,%d,%d), want zeros", r, g, b, a)
	}
}

func TestAverage_MeanOfChannels(t *testing.T) {
	pixels := []color.RGBA{
		{R: 100, G: 200, B: 50, A: 255},
		{R: 200, G: 100, B: 150, A: 255},
	}
	got := average(pixels).(color.RGBA)
	want := color.RGBA{R: 150, G: 150, B: 100, A: 255}
	if got != want {
		t.Errorf("average = %+v, want %+v", got, want)
	}
}

func TestDominantChannel(t *testing.T) {
	// Red has range 0..200; green 0..50; blue 0..10. -> R dominates (0).
	pixels := []color.RGBA{
		{R: 0, G: 0, B: 0},
		{R: 200, G: 50, B: 10},
	}
	if got := dominantChannel(pixels); got != 0 {
		t.Errorf("dominantChannel = %d, want 0 (red)", got)
	}

	// Green dominates.
	pixels = []color.RGBA{
		{R: 0, G: 0, B: 5},
		{R: 5, G: 200, B: 0},
	}
	if got := dominantChannel(pixels); got != 1 {
		t.Errorf("dominantChannel = %d, want 1 (green)", got)
	}
}
