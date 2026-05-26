package render

import (
	"bytes"
	"image"
	"image/color"
	"strings"
	"testing"
)

func TestSixel_HeaderAndTrailer(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	img.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	img.SetNRGBA(1, 1, color.NRGBA{B: 255, A: 255})

	var buf bytes.Buffer
	if err := Sixel(&buf, img); err != nil {
		t.Fatalf("Sixel: %v", err)
	}
	out := buf.String()

	if !strings.HasPrefix(out, "\x1bPq") {
		t.Errorf("sixel output missing DCS header; got prefix %q", out[:min(8, len(out))])
	}
	if !strings.HasSuffix(out, "\x1b\\\n") {
		t.Errorf("sixel output missing ST trailer; got suffix %q", out[max(0, len(out)-8):])
	}
}

func TestSixel_EmitsPaletteDefinitions(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	img.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})

	var buf bytes.Buffer
	if err := Sixel(&buf, img); err != nil {
		t.Fatalf("Sixel: %v", err)
	}
	// Palette entries look like "#N;2;R;G;B".
	if !strings.Contains(buf.String(), "#0;2;") {
		t.Errorf("sixel output missing palette entry; got %q", buf.String())
	}
}

func TestSixel_RowSeparator(t *testing.T) {
	// 8px tall -> at least one '-' between sixel bands (6px per band).
	img := image.NewNRGBA(image.Rect(0, 0, 1, 8))
	for y := 0; y < 8; y++ {
		img.SetNRGBA(0, y, color.NRGBA{R: 255, A: 255})
	}

	var buf bytes.Buffer
	if err := Sixel(&buf, img); err != nil {
		t.Fatalf("Sixel: %v", err)
	}
	if !strings.Contains(buf.String(), "-") {
		t.Errorf("expected sixel row separator '-' in output")
	}
}

func TestAllZero(t *testing.T) {
	if !allZero([]byte{0, 0, 0}) {
		t.Error("allZero(zeros) = false, want true")
	}
	if allZero([]byte{0, 1, 0}) {
		t.Error("allZero(non-zeros) = true, want false")
	}
	if !allZero(nil) {
		t.Error("allZero(nil) = false, want true")
	}
}
