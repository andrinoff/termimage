//go:build !linux

package termimage

import "os"

func termChars(_ *os.File) (int, int)          { return 220, 50 }
func cellPixels(_ *os.File) (cellW, cellH int) { return 8, 16 }
