//go:build !linux

package sandbox

// apply is a no-op on non-Linux platforms.
func apply(_ string) error { return nil }
