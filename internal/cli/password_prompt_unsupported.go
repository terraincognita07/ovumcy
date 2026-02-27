//go:build !windows && !linux && !darwin && !freebsd && !netbsd && !openbsd && !dragonfly

package cli

import (
	"errors"
	"os"
)

func readPasswordNoEcho(_ *os.File) ([]byte, error) {
	return nil, errors.New("unsupported platform")
}
