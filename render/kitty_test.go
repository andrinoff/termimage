package render

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
)

func TestKitty_StartsWithAPC(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	img.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})

	var buf bytes.Buffer
	if err := Kitty(&buf, img); err != nil {
		t.Fatalf("Kitty: %v", err)
	}
	out := buf.String()

	if !strings.HasPrefix(out, "\x1b_G") {
		t.Errorf("missing Kitty APC prefix; got prefix %q", out[:min(8, len(out))])
	}
	if !strings.Contains(out, "a=T,f=100") {
		t.Errorf("first chunk missing transmit-and-display controls; got %q", out)
	}
	if !strings.HasSuffix(out, "\n") {
		t.Errorf("output should end with newline")
	}
}

func TestKitty_PayloadDecodesToPNG(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 3, 3))
	img.SetNRGBA(0, 0, color.NRGBA{R: 200, G: 100, B: 50, A: 255})

	var buf bytes.Buffer
	if err := Kitty(&buf, img); err != nil {
		t.Fatalf("Kitty: %v", err)
	}

	// Concatenate every chunk's base64 payload, decode, expect a valid PNG.
	out := buf.String()
	var b64 strings.Builder
	for {
		i := strings.Index(out, "\x1b_G")
		if i < 0 {
			break
		}
		out = out[i+3:]
		semi := strings.Index(out, ";")
		end := strings.Index(out, "\x1b\\")
		if semi < 0 || end < 0 {
			break
		}
		b64.WriteString(out[semi+1 : end])
		out = out[end+2:]
	}

	raw, err := base64.StdEncoding.DecodeString(b64.String())
	if err != nil {
		t.Fatalf("base64 decode: %v", err)
	}
	decoded, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("png decode: %v", err)
	}
	b := decoded.Bounds()
	if b.Dx() != 3 || b.Dy() != 3 {
		t.Errorf("expected decoded 3x3, got %dx%d", b.Dx(), b.Dy())
	}
}

func TestKitty_LargeImageChunked(t *testing.T) {
	// Use a noisy pattern so PNG can't compress it small — we need the base64
	// payload to exceed the 4096-byte chunk boundary.
	img := image.NewNRGBA(image.Rect(0, 0, 256, 256))
	seed := uint32(2166136261)
	for y := 0; y < 256; y++ {
		for x := 0; x < 256; x++ {
			seed = (seed ^ uint32(x*7+y*13)) * 16777619
			img.SetNRGBA(x, y, color.NRGBA{
				R: uint8(seed),
				G: uint8(seed >> 8),
				B: uint8(seed >> 16),
				A: 255,
			})
		}
	}

	var buf bytes.Buffer
	if err := Kitty(&buf, img); err != nil {
		t.Fatalf("Kitty: %v", err)
	}
	out := buf.String()
	if c := strings.Count(out, "\x1b_G"); c < 2 {
		t.Errorf("expected multiple Kitty chunks, got %d", c)
	}
	if !strings.Contains(out, "m=1") {
		t.Errorf("expected at least one chunk with m=1 (more)")
	}
	if !strings.Contains(out, "m=0") {
		t.Errorf("expected final chunk with m=0 (last)")
	}
}
