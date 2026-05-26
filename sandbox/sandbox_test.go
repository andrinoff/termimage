package sandbox

import (
	"bytes"
	"image"
	"image/color"
	"testing"
)

func TestIsWorker(t *testing.T) {
	t.Setenv("TERMIMAGE_WORKER", "")
	if IsWorker() {
		t.Error("IsWorker() = true when env unset")
	}
	t.Setenv("TERMIMAGE_WORKER", "1")
	if !IsWorker() {
		t.Error("IsWorker() = false when env=1")
	}
	t.Setenv("TERMIMAGE_WORKER", "0")
	if IsWorker() {
		t.Error("IsWorker() = true when env=0")
	}
}

func TestPixelsRoundTrip(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 3, 2))
	src.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	src.SetNRGBA(1, 0, color.NRGBA{G: 255, A: 255})
	src.SetNRGBA(2, 0, color.NRGBA{B: 255, A: 255})
	src.SetNRGBA(0, 1, color.NRGBA{R: 1, G: 2, B: 3, A: 4})
	src.SetNRGBA(1, 1, color.NRGBA{R: 5, G: 6, B: 7, A: 8})
	src.SetNRGBA(2, 1, color.NRGBA{R: 9, G: 10, B: 11, A: 12})

	var buf bytes.Buffer
	if err := writePixels(&buf, src); err != nil {
		t.Fatalf("writePixels: %v", err)
	}

	got, err := readPixels(&buf)
	if err != nil {
		t.Fatalf("readPixels: %v", err)
	}

	if !got.Bounds().Eq(src.Bounds()) {
		t.Fatalf("bounds mismatch: got %v want %v", got.Bounds(), src.Bounds())
	}
	if !bytes.Equal(got.Pix, src.Pix) {
		t.Errorf("pixel bytes differ after round-trip")
	}
}

func TestReadPixels_RejectsSuspiciousDimensions(t *testing.T) {
	// Header claims 100000x100000 — readPixels should refuse before allocating.
	buf := bytes.NewBuffer([]byte{
		0xa0, 0x86, 0x01, 0x00, // 100000 LE
		0xa0, 0x86, 0x01, 0x00, // 100000 LE
	})
	if _, err := readPixels(buf); err == nil {
		t.Error("expected error on absurd dimensions")
	}
}

func TestReadPixels_RejectsZeroDimensions(t *testing.T) {
	buf := bytes.NewBuffer([]byte{0, 0, 0, 0, 0, 0, 0, 0})
	if _, err := readPixels(buf); err == nil {
		t.Error("expected error on zero dimensions")
	}
}

func TestReadPixels_ShortHeader(t *testing.T) {
	buf := bytes.NewBuffer([]byte{1, 2, 3})
	if _, err := readPixels(buf); err == nil {
		t.Error("expected error on truncated header")
	}
}
