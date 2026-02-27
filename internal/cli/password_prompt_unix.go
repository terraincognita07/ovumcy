//go:build linux || darwin || freebsd || netbsd || openbsd || dragonfly

package cli

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

func readPasswordNoEcho(stdin *os.File) ([]byte, error) {
	if stdin == nil {
		return nil, errors.New("stdin unavailable")
	}

	fd := int(stdin.Fd())
	termios, err := unix.IoctlGetTermios(fd, termiosReadRequest)
	if err != nil {
		return nil, err
	}
	originalTermios := *termios
	updatedTermios := originalTermios
	updatedTermios.Lflag &^= unix.ECHO

	if err := unix.IoctlSetTermios(fd, termiosWriteRequest, &updatedTermios); err != nil {
		return nil, err
	}
	defer func() {
		_ = unix.IoctlSetTermios(fd, termiosWriteRequest, &originalTermios)
	}()

	reader := bufio.NewReader(stdin)
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	line = strings.TrimRight(line, "\r\n")
	return []byte(line), nil
}
