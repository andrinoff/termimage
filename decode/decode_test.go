package decode

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestBytes_EmptyInput(t *testing.T) {
	if _, err := Bytes(nil); err == nil {
		t.Error("expected error on nil input")
	}
	if _, err := Bytes([]byte{}); err == nil {
		t.Error("expected error on empty input")
	}
}

func TestBytes_InvalidInput(t *testing.T) {
	if _, err := Bytes([]byte("not an image")); err == nil {
		t.Error("expected error on garbage input")
	}
}

func TestBytes_DecodesPNG(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	src.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	src.SetNRGBA(1, 0, color.NRGBA{G: 255, A: 255})
	src.SetNRGBA(0, 1, color.NRGBA{B: 255, A: 255})
	src.SetNRGBA(1, 1, color.NRGBA{R: 255, G: 255, B: 255, A: 255})

	var buf bytes.Buffer
	if err := png.Encode(&buf, src); err != nil {
		t.Fatalf("encode: %v", err)
	}

	got, err := Bytes(buf.Bytes())
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	b := got.Bounds()
	if b.Dx() != 2 || b.Dy() != 2 {
		t.Fatalf("expected 2x2, got %dx%d", b.Dx(), b.Dy())
	}

	// Top-left red pixel should survive decode (allow channel order = RGBA).
	r, g, bl, a := got.At(0, 0).RGBA()
	if r>>8 != 255 || g>>8 != 0 || bl>>8 != 0 || a>>8 != 255 {
		t.Errorf("top-left pixel = (%d,%d,%d,%d), want red", r>>8, g>>8, bl>>8, a>>8)
	}
}

func TestFile_DecodesPNGFromDisk(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	src.SetNRGBA(0, 0, color.NRGBA{R: 200, G: 100, B: 50, A: 255})

	var buf bytes.Buffer
	if err := png.Encode(&buf, src); err != nil {
		t.Fatalf("encode: %v", err)
	}

	path := filepath.Join(t.TempDir(), "pixel.png")
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write tempfile: %v", err)
	}

	img, err := File(path)
	if err != nil {
		t.Fatalf("File: %v", err)
	}
	b := img.Bounds()
	if b.Dx() != 1 || b.Dy() != 1 {
		t.Errorf("expected 1x1, got %dx%d", b.Dx(), b.Dy())
	}
}

func TestFile_NonexistentPath(t *testing.T) {
	if _, err := File("/nonexistent/path/xyzzy.png"); err == nil {
		t.Error("expected error reading nonexistent file")
	}
}
