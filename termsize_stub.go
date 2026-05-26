//go:build !linux

package termimage

import "os"

func termPixels(_ *os.File) (int, int) { return 1920, 1080 }
