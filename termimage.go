// Package termimage renders images in a terminal using Kitty graphics, Sixel,
// or Unicode half-block characters as fallback. Image decoding runs inside a
// sandboxed subprocess (Landlock + seccomp on Linux) so untrusted bytes never
// touch the parent process.
//
// The consuming binary must call MaybeRunWorker() at the very top of main():
//
//	func main() {
//	    termimage.MaybeRunWorker()
//	    // ... rest of your app
//	}
package termimage

import (
	"image"
	"io"
	"os"

	"github.com/floatpane/termimage/decode"
	"github.com/floatpane/termimage/detect"
	"github.com/floatpane/termimage/internal/resize"
	"github.com/floatpane/termimage/render"
	"github.com/floatpane/termimage/sandbox"
)

// MaybeRunWorker must be called at the top of main(). If this process was
// spawned as a sandbox worker it applies OS restrictions, decodes the image,
// writes pixels to stdout, and exits. Otherwise it returns immediately.
func MaybeRunWorker() { sandbox.MaybeRunWorker() }

// Protocol selects a terminal rendering protocol.
type Protocol = detect.Protocol

const (
	Auto      Protocol = -1
	HalfBlock          = detect.HalfBlock
	Sixel              = detect.Sixel
	Kitty              = detect.Kitty
)

// Options configures image display.
type Options struct {
	// MaxWidth / MaxHeight in pixels. 0 = detect from terminal.
	MaxWidth, MaxHeight int

	// Protocol selects the rendering protocol. Auto detects from $TERM etc.
	Protocol Protocol

	// Sandboxed runs the decoder in a subprocess with Landlock + seccomp.
	// Requires the consuming binary to call MaybeRunWorker() in main().
	Sandboxed bool
}

// Display decodes the image at path and writes terminal graphics to w.
func Display(w io.Writer, path string, opts Options) error {
	maxW, maxH := effectiveDimensions(opts)

	proto := opts.Protocol
	if proto == Auto {
		proto = detect.Best()
	}

	var img *image.NRGBA
	var err error

	if opts.Sandboxed {
		img, err = sandbox.Decode(path)
	} else {
		img, err = decode.File(path)
	}
	if err != nil {
		return err
	}

	scaled := resize.Fit(img, maxW, maxH)
	return renderWith(w, scaled, proto)
}

func renderWith(w io.Writer, img *image.NRGBA, proto Protocol) error {
	switch proto {
	case Kitty:
		return render.Kitty(w, img)
	case Sixel:
		return render.Sixel(w, img)
	default:
		return render.HalfBlock(w, img)
	}
}

func effectiveDimensions(opts Options) (int, int) {
	w, h := opts.MaxWidth, opts.MaxHeight
	if w > 0 && h > 0 {
		return w, h
	}
	tw, th := detectTermPixels()
	if w <= 0 {
		w = tw
	}
	if h <= 0 {
		h = th
	}
	return w, h
}

func detectTermPixels() (int, int) {
	f, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return 1920, 1080
	}
	defer func() { _ = f.Close() }()
	return termPixels(f)
}
