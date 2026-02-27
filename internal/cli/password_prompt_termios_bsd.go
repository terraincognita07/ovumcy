//go:build darwin || freebsd || netbsd || openbsd || dragonfly

package cli

import "golang.org/x/sys/unix"

const (
	termiosReadRequest  = unix.TIOCGETA
	termiosWriteRequest = unix.TIOCSETA
)
