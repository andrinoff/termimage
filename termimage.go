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
//
// # HalfBlock and text layout
//
// HalfBlock renders using real terminal character cells (▀ U+2580). Unlike
// Kitty or Sixel, the output occupies rows × cols cells in the terminal scroll
// buffer. Cursor save/restore (\x1b[s / \x1b[u) does not undo cell content.
// TUIs that need pixel-layer rendering should use Kitty or Sixel and set
// AllowHalfBlock: false on Options to prevent fallback.
package termimage

import (
	"bufio"
	"context"
	"fmt"
	"image"
	"io"
	"os"

	"github.com/floatpane/termimage/decode"
	"github.com/floatpane/termimage/detect"
	"github.com/floatpane/termimage/internal/resize"
	"github.com/floatpane/termimage/internal/source"
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
	// MaxWidth / MaxHeight are the pixel bounds for image scaling.
	//
	// For HalfBlock, 1 pixel = 1 character column (width) or half a character
	// row (height), so MaxHeight=100 produces at most 50 character rows.
	// For Kitty and Sixel, pixels map to physical screen pixels.
	//
	// 0 = detect from terminal. Default caps at (terminal_cols, terminal_rows-2)
	// in character-cell units, leaving two rows of headroom for the shell prompt.
	// Aspect ratio is always preserved: Fit scales uniformly so neither dimension
	// exceeds the limit. When only one dimension is set, the other is detected.
	MaxWidth, MaxHeight int

	// Protocol selects the rendering protocol. Auto detects from $TERM etc.
	// See detect.Best for detection rules.
	// TUIs that must avoid HalfBlock (which occupies real text cells — see
	// package doc) should call detect.Best() first; if it returns HalfBlock,
	// set Protocol explicitly or return an error before calling Display.
	Protocol Protocol

	// Sandboxed runs the decoder in a subprocess with Landlock + seccomp.
	// Requires the consuming binary to call MaybeRunWorker() in main().
	Sandboxed bool
}

// Display decodes the image at src and writes terminal graphics to w.
// src may be a local file path, a data: URI (base64), or an http(s):// URL.
func Display(w io.Writer, src string, opts Options) error {
	_, _, err := display(context.Background(), w, src, opts)
	return err
}

// DisplayContext is Display with caller-supplied context for cancellation of
// remote fetches and sandboxed decoding.
func DisplayContext(ctx context.Context, w io.Writer, src string, opts Options) error {
	_, _, err := display(ctx, w, src, opts)
	return err
}

// DisplayWithSize renders the image and returns the terminal character-cell
// dimensions (cols, rows) it occupies. Use this instead of Display + Dims to
// avoid decoding the image twice.
//
// For HalfBlock, cols = image pixel width, rows = ceil(image pixel height / 2).
// For Kitty and Sixel, cols and rows are derived from the cell pixel size
// reported by the terminal (TIOCGWINSZ), falling back to 8×16 px per cell.
func DisplayWithSize(w io.Writer, src string, opts Options) (cols, rows int, err error) {
	return display(context.Background(), w, src, opts)
}

// DisplayContextWithSize is DisplayWithSize with caller-supplied context.
func DisplayContextWithSize(ctx context.Context, w io.Writer, src string, opts Options) (cols, rows int, err error) {
	return display(ctx, w, src, opts)
}

// Clear erases a previously rendered image.
//
// For Kitty, rows is ignored — all visible image placements are deleted via the
// Kitty graphics protocol delete command. Call immediately after the image is
// no longer needed; no cursor positioning is required.
//
// For Sixel and HalfBlock, the caller must position the cursor on the last row
// of the image before calling Clear. rows should be the value returned by
// DisplayWithSize. Clear moves the cursor up by rows lines then erases to end
// of screen (\x1b[{rows}A\x1b[J).
func Clear(w io.Writer, proto Protocol, rows int) error {
	bw := bufio.NewWriterSize(w, 32)
	switch proto {
	case Kitty:
		if _, err := fmt.Fprint(bw, "\x1b_Ga=d,d=A\x1b\\"); err != nil {
			return err
		}
	default:
		if rows <= 0 {
			return nil
		}
		if _, err := fmt.Fprintf(bw, "\x1b[%dA\x1b[J", rows); err != nil {
			return err
		}
	}
	return bw.Flush()
}

// display is the shared implementation for all Display* variants.
func display(ctx context.Context, w io.Writer, src string, opts Options) (cols, rows int, err error) {
	proto, err := resolveProto(opts)
	if err != nil {
		return 0, 0, err
	}

	img, err := loadScaled(ctx, src, opts, proto)
	if err != nil {
		return 0, 0, err
	}

	cols, rows = pixelsToCells(img.Bounds().Dx(), img.Bounds().Dy(), proto)
	return cols, rows, renderWith(w, img, proto)
}

func resolveProto(opts Options) (Protocol, error) {
	if opts.Protocol != Auto {
		return opts.Protocol, nil
	}
	return detect.Best(), nil
}

// loadScaled resolves src, decodes, and scales to effectiveDimensions.
func loadScaled(ctx context.Context, src string, opts Options, proto Protocol) (*image.NRGBA, error) {
	maxW, maxH := effectiveDimensions(opts, proto)

	resolved, err := source.Resolve(ctx, src)
	if err != nil {
		return nil, err
	}

	var img *image.NRGBA
	switch resolved.Kind {
	case source.KindFile:
		if opts.Sandboxed {
			img, err = sandbox.DecodeContext(ctx, resolved.Path)
		} else {
			img, err = decode.File(resolved.Path)
		}
	case source.KindBytes:
		if opts.Sandboxed {
			img, err = sandbox.DecodeBytesContext(ctx, resolved.Bytes)
		} else {
			img, err = decode.Bytes(resolved.Bytes)
		}
	}
	if err != nil {
		return nil, err
	}

	return resize.Fit(img, maxW, maxH), nil
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

// effectiveDimensions returns pixel bounds for image scaling.
//
// Both protocols share the same calculation: character grid dimensions drive
// the limit, converted to pixels via cell size. Two rows are reserved as
// headroom for the shell prompt / surrounding content.
//
// For HalfBlock: 1 col = 1 px wide, 1 row = 2 px tall.
// For Kitty/Sixel: multiply by cell pixel dimensions from TIOCGWINSZ.
func effectiveDimensions(opts Options, proto Protocol) (int, int) {
	w, h := opts.MaxWidth, opts.MaxHeight
	if w > 0 && h > 0 {
		return w, h
	}

	cols, rows := detectTermChars()
	cw, ch := detectCellPixels()

	const headroom = 2
	effectiveRows := rows - headroom
	if effectiveRows < 1 {
		effectiveRows = 1
	}

	var tw, th int
	if proto == HalfBlock {
		tw, th = cols, effectiveRows*2
	} else {
		tw, th = cols*cw, effectiveRows*ch
	}

	if w <= 0 {
		w = tw
	}
	if h <= 0 {
		h = th
	}
	return w, h
}

// pixelsToCells converts scaled image pixel dimensions to terminal character
// cell dimensions (cols, rows) for the given protocol.
func pixelsToCells(pw, ph int, proto Protocol) (cols, rows int) {
	if proto == HalfBlock {
		return pw, (ph + 1) / 2
	}
	cw, ch := detectCellPixels()
	return (pw + cw - 1) / cw, (ph + ch - 1) / ch
}

func detectTermChars() (int, int) {
	f, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return 220, 50
	}
	defer func() { _ = f.Close() }()
	return termChars(f)
}

func detectCellPixels() (int, int) {
	f, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return 8, 16
	}
	defer func() { _ = f.Close() }()
	return cellPixels(f)
}

// detectTermPixels is retained for callers that need raw screen pixel dimensions.
func detectTermPixels() (int, int) {
	f, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return 1920, 1080
	}
	defer func() { _ = f.Close() }()
	return termPixels(f)
}
