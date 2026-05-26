package render

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"io"
)

// HalfBlock renders img using Unicode half-block characters (▀ U+2580).
// Each character cell encodes two vertically stacked pixels using
// foreground (top) and background (bottom) ANSI 24-bit color.
// Works on any terminal with UTF-8 and truecolor support.
func HalfBlock(w io.Writer, img image.Image) error {
	bw := bufio.NewWriter(w)
	b := img.Bounds()
	height := b.Dy()

	for y := b.Min.Y; y < b.Max.Y; y += 2 {
		for x := b.Min.X; x < b.Max.X; x++ {
			top := toRGB(img.At(x, y))

			var bot [3]uint8
			if y+1 < height+b.Min.Y {
				bot = toRGB(img.At(x, y+1))
			}

			if _, err := fmt.Fprintf(bw, "\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm▀",
				top[0], top[1], top[2],
				bot[0], bot[1], bot[2],
			); err != nil {
				return err
			}
		}
		if _, err := bw.WriteString("\x1b[0m\n"); err != nil {
			return err
		}
	}

	return bw.Flush()
}

func toRGB(c color.Color) [3]uint8 {
	r, g, b, _ := c.RGBA()
	return [3]uint8{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8)}
}
