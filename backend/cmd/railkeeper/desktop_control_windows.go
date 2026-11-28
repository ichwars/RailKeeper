package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/sys/windows"
)

func openDesktopControl(args []string, standalone bool) (*desktopControl, error) {
	for _, arg := range args {
		value, found := strings.CutPrefix(arg, "--desktop-control=")
		if !found {
			continue
		}
		if !standalone {
			return nil, errors.New("desktop control requires Windows standalone mode")
		}
		stopValue, readyValue, found := strings.Cut(value, ",")
		if !found || stopValue == readyValue {
			return nil, errors.New("invalid desktop control argument")
		}
		stop, err := openDesktopPipe(stopValue)
		if err != nil {
			return nil, fmt.Errorf("open shutdown pipe: %w", err)
		}
		ready, err := openDesktopPipe(readyValue)
		if err != nil {
			_ = stop.Close()
			return nil, fmt.Errorf("open readiness pipe: %w", err)
		}
		return &desktopControl{stop: stop, ready: ready}, nil
	}
	return nil, nil
}

func openDesktopPipe(value string) (*os.File, error) {
	n, err := strconv.ParseUint(value, 10, strconv.IntSize)
	if err != nil || strconv.FormatUint(n, 10) != value {
		return nil, errors.New("invalid inherited pipe handle")
	}
	handle := windows.Handle(n)
	kind, err := windows.GetFileType(handle)
	if err != nil || kind != windows.FILE_TYPE_PIPE {
		return nil, errors.New("inherited handle is not a pipe")
	}
	if err := windows.SetHandleInformation(handle, windows.HANDLE_FLAG_INHERIT, 0); err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(handle), "desktop control"), nil
}
