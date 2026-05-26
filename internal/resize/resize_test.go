package resize

import (
	"image"
	"testing"
)

func TestFit_ReturnsUnchangedWhenAlreadyFits(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 100, 50))
	out := Fit(src, 200, 200)
	if out != src {
		t.Errorf("Fit should return src unchanged when it already fits; got new image")
	}
}

func TestFit_ScalesDownPreservingAspect(t *testing.T) {
	// 200x100 src (2:1) into 100x100 box -> should produce 100x50.
	src := image.NewNRGBA(image.Rect(0, 0, 200, 100))
	out := Fit(src, 100, 100)
	b := out.Bounds()
	if b.Dx() != 100 || b.Dy() != 50 {
		t.Errorf("expected 100x50, got %dx%d", b.Dx(), b.Dy())
	}
}

func TestFit_ScalesByHeightConstraint(t *testing.T) {
	// 100x200 (1:2) into 100x100 -> 50x100.
	src := image.NewNRGBA(image.Rect(0, 0, 100, 200))
	out := Fit(src, 100, 100)
	b := out.Bounds()
	if b.Dx() != 50 || b.Dy() != 100 {
		t.Errorf("expected 50x100, got %dx%d", b.Dx(), b.Dy())
	}
}

func TestFit_MinimumOnePixel(t *testing.T) {
	// Extreme downscale must clamp to >= 1 in both dimensions.
	src := image.NewNRGBA(image.Rect(0, 0, 1000, 1))
	out := Fit(src, 10, 10)
	b := out.Bounds()
	if b.Dx() < 1 || b.Dy() < 1 {
		t.Errorf("dimensions clamped below 1: %dx%d", b.Dx(), b.Dy())
	}
}

func TestFit_SquareIntoSquare(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 400, 400))
	out := Fit(src, 100, 100)
	b := out.Bounds()
	if b.Dx() != 100 || b.Dy() != 100 {
		t.Errorf("expected 100x100, got %dx%d", b.Dx(), b.Dy())
	}
}
