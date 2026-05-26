package render

import (
	"image"
	"image/color"
	"math"
	"sort"
)

// medianCut quantizes img to at most n colors using the median cut algorithm.
func medianCut(img image.Image, n int) []color.Color {
	b := img.Bounds()
	pixels := make([]color.RGBA, 0, b.Dx()*b.Dy())

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := img.At(x, y).RGBA()
			if a < 0x8000 {
				continue
			}
			pixels = append(pixels, color.RGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(bl >> 8),
				A: 255,
			})
		}
	}

	buckets := [][]color.RGBA{pixels}
	for len(buckets) < n && anyBucketSplittable(buckets) {
		buckets = splitLargest(buckets)
	}

	palette := make([]color.Color, len(buckets))
	for i, bucket := range buckets {
		palette[i] = average(bucket)
	}
	return palette
}

func anyBucketSplittable(bs [][]color.RGBA) bool {
	for _, b := range bs {
		if len(b) > 1 {
			return true
		}
	}
	return false
}

func splitLargest(buckets [][]color.RGBA) [][]color.RGBA {
	// Find largest bucket.
	idx := 0
	for i, b := range buckets {
		if len(b) > len(buckets[idx]) {
			idx = i
		}
	}

	bucket := buckets[idx]
	// Find channel with greatest range.
	ch := dominantChannel(bucket)

	sort.Slice(bucket, func(i, j int) bool {
		switch ch {
		case 0:
			return bucket[i].R < bucket[j].R
		case 1:
			return bucket[i].G < bucket[j].G
		default:
			return bucket[i].B < bucket[j].B
		}
	})

	mid := len(bucket) / 2
	left := bucket[:mid]
	right := bucket[mid:]

	result := make([][]color.RGBA, 0, len(buckets)+1)
	result = append(result, buckets[:idx]...)
	result = append(result, left, right)
	result = append(result, buckets[idx+1:]...)
	return result
}

func dominantChannel(pixels []color.RGBA) int {
	var minR, minG, minB uint8 = 255, 255, 255
	var maxR, maxG, maxB uint8

	for _, p := range pixels {
		if p.R < minR {
			minR = p.R
		}
		if p.R > maxR {
			maxR = p.R
		}
		if p.G < minG {
			minG = p.G
		}
		if p.G > maxG {
			maxG = p.G
		}
		if p.B < minB {
			minB = p.B
		}
		if p.B > maxB {
			maxB = p.B
		}
	}

	rr := int(maxR) - int(minR)
	gg := int(maxG) - int(minG)
	bb := int(maxB) - int(minB)

	if rr >= gg && rr >= bb {
		return 0
	}
	if gg >= bb {
		return 1
	}
	return 2
}

func average(pixels []color.RGBA) color.Color {
	if len(pixels) == 0 {
		return color.RGBA{}
	}
	var r, g, b int64
	for _, p := range pixels {
		r += int64(p.R)
		g += int64(p.G)
		b += int64(p.B)
	}
	n := int64(len(pixels))
	return color.RGBA{
		R: uint8(r / n),
		G: uint8(g / n),
		B: uint8(b / n),
		A: 255,
	}
}

func nearestColor(palette []color.Color, c color.Color) int {
	r0, g0, b0, _ := c.RGBA()
	best := 0
	bestDist := math.MaxFloat64

	for i, p := range palette {
		r1, g1, b1, _ := p.RGBA()
		dr := float64(int(r0>>8) - int(r1>>8))
		dg := float64(int(g0>>8) - int(g1>>8))
		db := float64(int(b0>>8) - int(b1>>8))
		d := dr*dr + dg*dg + db*db
		if d < bestDist {
			bestDist = d
			best = i
		}
	}
	return best
}
