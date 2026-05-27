// Package resize provides fast bilinear image scaling.
package resize

import (
	"image"

	"golang.org/x/image/draw"
)

// Fit scales src to fit within maxW×maxH while preserving aspect ratio.
// The scale factor is min(maxW/srcW, maxH/srcH), so both dimensions are
// satisfied simultaneously. Returns src unchanged if it already fits within
// the bounds.
func Fit(src *image.NRGBA, maxW, maxH int) *image.NRGBA {
	b := src.Bounds()
	sw, sh := b.Dx(), b.Dy()
	if sw <= maxW && sh <= maxH {
		return src
	}

	// Scale factor preserving aspect ratio.
	scaleW := float64(maxW) / float64(sw)
	scaleH := float64(maxH) / float64(sh)
	scale := scaleW
	if scaleH < scale {
		scale = scaleH
	}

	dw := max(1, int(float64(sw)*scale))
	dh := max(1, int(float64(sh)*scale))

	dst := image.NewNRGBA(image.Rect(0, 0, dw, dh))
	draw.BiLinear.Scale(dst, dst.Bounds(), src, b, draw.Src, nil)
	return dst
}
