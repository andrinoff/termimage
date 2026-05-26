package render

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"io"
)

const kittyChunk = 4096

// Kitty encodes img and writes a Kitty graphics protocol sequence to w.
// The image is re-encoded as PNG (Kitty format f=100) and chunked.
func Kitty(w io.Writer, img image.Image) error {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return fmt.Errorf("kitty: png encode: %w", err)
	}
	return kittyChunks(w, buf.Bytes())
}

func kittyChunks(w io.Writer, payload []byte) error {
	enc := base64.StdEncoding.EncodeToString(payload)

	for i := 0; i < len(enc); i += kittyChunk {
		end := i + kittyChunk
		if end > len(enc) {
			end = len(enc)
		}
		chunk := enc[i:end]
		more := 1
		if end == len(enc) {
			more = 0
		}

		var ctrl string
		if i == 0 {
			// a=T: transmit & display. f=100: PNG. q=2: suppress ACK.
			ctrl = fmt.Sprintf("a=T,f=100,q=2,m=%d", more)
		} else {
			ctrl = fmt.Sprintf("q=2,m=%d", more)
		}

		if _, err := fmt.Fprintf(w, "\x1b_G%s;%s\x1b\\", ctrl, chunk); err != nil {
			return err
		}
	}

	_, err := fmt.Fprintln(w)
	return err
}
