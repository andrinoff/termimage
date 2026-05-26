package render

import (
	"bytes"
	"image"
	"image/color"
	"strings"
	"testing"
)

func TestHalfBlock_EmitsHalfBlockGlyph(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	img.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	img.SetNRGBA(1, 0, color.NRGBA{G: 255, A: 255})
	img.SetNRGBA(0, 1, color.NRGBA{B: 255, A: 255})
	img.SetNRGBA(1, 1, color.NRGBA{R: 255, G: 255, B: 255, A: 255})

	var buf bytes.Buffer
	if err := HalfBlock(&buf, img); err != nil {
		t.Fatalf("HalfBlock: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "▀") {
		t.Errorf("output missing half-block glyph; got %q", out)
	}
	if !strings.Contains(out, "\x1b[38;2;255;0;0m") {
		t.Errorf("output missing red foreground escape; got %q", out)
	}
	if !strings.Contains(out, "\x1b[48;2;0;0;255m") {
		t.Errorf("output missing blue background escape; got %q", out)
	}
	if !strings.Contains(out, "\x1b[0m") {
		t.Errorf("output missing reset escape; got %q", out)
	}
}

func TestHalfBlock_RowsCountMatchesImageHeight(t *testing.T) {
	// Even height: 4 rows -> 2 lines of output (2 px per line).
	img := image.NewNRGBA(image.Rect(0, 0, 1, 4))
	var buf bytes.Buffer
	if err := HalfBlock(&buf, img); err != nil {
		t.Fatalf("HalfBlock: %v", err)
	}
	if got := strings.Count(buf.String(), "\n"); got != 2 {
		t.Errorf("expected 2 newlines for 4-px-tall image, got %d", got)
	}
}

func TestHalfBlock_OddHeightStillTerminatesEachRow(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 1, 3))
	var buf bytes.Buffer
	if err := HalfBlock(&buf, img); err != nil {
		t.Fatalf("HalfBlock: %v", err)
	}
	// ceil(3/2) = 2 lines.
	if got := strings.Count(buf.String(), "\n"); got != 2 {
		t.Errorf("expected 2 newlines for 3-px-tall image, got %d", got)
	}
}
