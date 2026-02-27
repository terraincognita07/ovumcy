//go:build linux

package cli

import "golang.org/x/sys/unix"

const (
	termiosReadRequest  = unix.TCGETS
	termiosWriteRequest = unix.TCSETS
)
