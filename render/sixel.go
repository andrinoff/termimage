package render

import (
	"bufio"
	"fmt"
	"image"
	"io"
)

// Sixel encodes img as a DEC Sixel sequence and writes it to w.
// Colors are quantized to 256 using a fast median-cut algorithm.
func Sixel(w io.Writer, img image.Image) error {
	bw := bufio.NewWriter(w)
	b := img.Bounds()
	width, height := b.Dx(), b.Dy()

	// Build palette via median cut.
	palette := medianCut(img, 256)

	// Sixel header.
	if _, err := bw.WriteString("\x1bPq"); err != nil {
		return err
	}

	// Write color definitions.
	for i, c := range palette {
		r, g, bl, _ := c.RGBA()
		// Sixel color values are 0-100 (percentage).
		if _, err := fmt.Fprintf(bw, "#%d;2;%d;%d;%d", i,
			int(r>>8)*100/255,
			int(g>>8)*100/255,
			int(bl>>8)*100/255,
		); err != nil {
			return err
		}
	}

	// Each sixel row covers 6 pixel rows. Build one color band at a time.
	for bandY := 0; bandY < height; bandY += 6 {
		bands := make([][]byte, len(palette))
		for i := range bands {
			bands[i] = make([]byte, width)
		}

		for x := 0; x < width; x++ {
			for bit := 0; bit < 6; bit++ {
				py := bandY + bit
				if py >= height {
					break
				}
				c := img.At(b.Min.X+x, b.Min.Y+py)
				idx := nearestColor(palette, c)
				bands[idx][x] |= 1 << uint(bit)
			}
		}

		for i, band := range bands {
			if allZero(band) {
				continue
			}
			if _, err := fmt.Fprintf(bw, "#%d", i); err != nil {
				return err
			}
			for _, v := range band {
				if err := bw.WriteByte(v + 63); err != nil {
					return err
				}
			}
			if err := bw.WriteByte('$'); err != nil { // carriage return within sixel row
				return err
			}
		}
		if err := bw.WriteByte('-'); err != nil { // next sixel row
			return err
		}
	}

	if _, err := bw.WriteString("\x1b\\\n"); err != nil {
		return err
	}
	return bw.Flush()
}

func allZero(b []byte) bool {
	for _, v := range b {
		if v != 0 {
			return false
		}
	}
	return true
}
