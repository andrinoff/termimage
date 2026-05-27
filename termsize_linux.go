//go:build linux

package termimage

import (
	"os"

	"golang.org/x/sys/unix"
)

// termChars returns the terminal dimensions as (cols, rows).
func termChars(f *os.File) (int, int) {
	ws, err := unix.IoctlGetWinsize(int(f.Fd()), unix.TIOCGWINSZ)
	if err != nil || ws.Col == 0 || ws.Row == 0 {
		return 220, 50
	}
	return int(ws.Col), int(ws.Row)
}

// cellPixels returns the pixel dimensions of a single terminal character cell.
func cellPixels(f *os.File) (cellW, cellH int) {
	ws, err := unix.IoctlGetWinsize(int(f.Fd()), unix.TIOCGWINSZ)
	if err != nil || ws.Col == 0 || ws.Row == 0 || ws.Xpixel == 0 || ws.Ypixel == 0 {
		return 8, 16
	}
	return int(ws.Xpixel) / int(ws.Col), int(ws.Ypixel) / int(ws.Row)
}
