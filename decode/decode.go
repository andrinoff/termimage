// Package decode wraps stb_image via CGo for fast multi-format image decoding.
// Supported: JPEG, PNG, BMP, GIF, TGA, PSD, HDR, PNM.
package decode

/*
#cgo CFLAGS: -O3
#cgo LDFLAGS: -lm
#include "stb_image.h"
#include <stdlib.h>

static unsigned char* stb_decode(
    const unsigned char* buf, int len,
    int* w, int* h
) {
    int ch;
    return stbi_load_from_memory(buf, len, w, h, &ch, 4);
}

static void stb_free(unsigned char* p) { stbi_image_free(p); }
static const char* stb_err(void) { return stbi_failure_reason(); }
*/
import "C"

import (
	"fmt"
	"image"
	"unsafe"
)

// File decodes the image at path into an NRGBA image.
func File(path string) (*image.NRGBA, error) {
	// Read outside CGo to keep C heap small.
	data, err := readFile(path)
	if err != nil {
		return nil, err
	}
	return Bytes(data)
}

// Bytes decodes raw image bytes into an NRGBA image.
// Output is always 4-channel RGBA, regardless of source format.
func Bytes(data []byte) (*image.NRGBA, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty input")
	}

	var w, h C.int
	ptr := C.stb_decode(
		(*C.uchar)(unsafe.Pointer(&data[0])),
		C.int(len(data)),
		&w, &h,
	)
	if ptr == nil {
		return nil, fmt.Errorf("stb_image: %s", C.GoString(C.stb_err()))
	}
	defer C.stb_free(ptr)

	iw, ih := int(w), int(h)
	img := image.NewNRGBA(image.Rect(0, 0, iw, ih))

	// Zero-copy slice over C memory; copy before stb_free runs.
	n := iw * ih * 4
	src := unsafe.Slice((*byte)(unsafe.Pointer(ptr)), n)
	copy(img.Pix, src)

	return img, nil
}
