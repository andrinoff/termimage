//go:build linux

package sandbox

import (
	"github.com/landlock-lsm/go-landlock/landlock"
)

// apply locks down the worker process before any untrusted bytes are read.
// Landlock (kernel ≥5.13) restricts filesystem access to the target file only.
// Seccomp syscall filtering is a TODO — see sandbox_seccomp_linux.go once the
// allowlist is tuned for the Go runtime + stb_image.
func apply(path string) error {
	// BestEffort silently skips Landlock if the kernel is too old.
	return landlock.V3.BestEffort().RestrictPaths(
		landlock.ROFiles(path),
	)
}
