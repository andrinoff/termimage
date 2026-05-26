package sandbox

import (
	"image"

	"github.com/floatpane/termimage/decode"
)

// decodeBytes is the CGo stb_image call, done inside the worker after sandbox.
func decodeBytes(data []byte) (*image.NRGBA, error) {
	return decode.Bytes(data)
}
