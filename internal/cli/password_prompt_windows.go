//go:build windows

package cli

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"

	"golang.org/x/sys/windows"
)

func readPasswordNoEcho(stdin *os.File) ([]byte, error) {
	if stdin == nil {
		return nil, errors.New("stdin unavailable")
	}

	handle := windows.Handle(stdin.Fd())
	var originalMode uint32
	if err := windows.GetConsoleMode(handle, &originalMode); err != nil {
		return nil, err
	}

	updatedMode := originalMode &^ windows.ENABLE_ECHO_INPUT
	if err := windows.SetConsoleMode(handle, updatedMode); err != nil {
		return nil, err
	}
	defer func() {
		_ = windows.SetConsoleMode(handle, originalMode)
	}()

	reader := bufio.NewReader(stdin)
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	line = strings.TrimRight(line, "\r\n")
	return []byte(line), nil
}
