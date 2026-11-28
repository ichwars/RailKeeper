package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

func main() {
	if err := startRailKeeper(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		message, _ := windows.UTF16PtrFromString("RailKeeper konnte nicht gestartet werden. / Could not start RailKeeper.\n\n" + err.Error())
		title, _ := windows.UTF16PtrFromString("RailKeeper")
		_, _ = windows.MessageBox(0, message, title, windows.MB_OK|windows.MB_ICONERROR)
		os.Exit(1)
	}
}

type instance struct {
	process windows.Handle
	stop    *os.File
	ready   *os.File
}

func launchInstance(directory string) (_ *instance, err error) {
	childStop, stop, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("create shutdown pipe: %w", err)
	}
	defer func() { _ = childStop.Close() }()

	ready, childReady, err := os.Pipe()
	if err != nil {
		_ = stop.Close()
		return nil, fmt.Errorf("create readiness pipe: %w", err)
	}
	defer func() { _ = childReady.Close() }()
	defer func() {
		if err != nil {
			_ = stop.Close()
			_ = ready.Close()
		}
	}()

	handles := [2]windows.Handle{windows.Handle(childStop.Fd()), windows.Handle(childReady.Fd())}
	for _, handle := range handles {
		if err := windows.SetHandleInformation(handle, windows.HANDLE_FLAG_INHERIT, windows.HANDLE_FLAG_INHERIT); err != nil {
			return nil, fmt.Errorf("prepare inherited pipes: %w", err)
		}
	}
	process, err := createRailKeeperProcess(directory, handles)
	if err != nil {
		return nil, err
	}
	return &instance{process: process, stop: stop, ready: ready}, nil
}

func createRailKeeperProcess(directory string, handles [2]windows.Handle) (windows.Handle, error) {
	attributes, err := windows.NewProcThreadAttributeList(1)
	if err != nil {
		return 0, fmt.Errorf("prepare process startup: %w", err)
	}
	defer attributes.Delete()
	// Only these two handles are inherited. In particular, the server never
	// gets a copy of the launcher's shutdown writer.
	if err := attributes.Update(windows.PROC_THREAD_ATTRIBUTE_HANDLE_LIST, unsafe.Pointer(&handles[0]), unsafe.Sizeof(handles)); err != nil {
		return 0, fmt.Errorf("select inherited pipes: %w", err)
	}

	target := filepath.Join(directory, "RailKeeper.exe")
	executable, err := windows.UTF16PtrFromString(target)
	if err != nil {
		return 0, err
	}
	controlArg := fmt.Sprintf("--desktop-control=%d,%d", handles[0], handles[1])
	commandLine, err := windows.UTF16PtrFromString(windows.ComposeCommandLine([]string{target, "--standalone", controlArg}))
	if err != nil {
		return 0, err
	}
	workingDirectory, err := windows.UTF16PtrFromString(directory)
	if err != nil {
		return 0, err
	}
	startup := windows.StartupInfoEx{
		StartupInfo: windows.StartupInfo{
			Cb:         uint32(unsafe.Sizeof(windows.StartupInfoEx{})),
			Flags:      windows.STARTF_USESHOWWINDOW,
			ShowWindow: windows.SW_SHOWMINNOACTIVE,
		},
		ProcThreadAttributeList: attributes.List(),
	}
	var process windows.ProcessInformation
	err = windows.CreateProcess(executable, commandLine, nil, nil, true,
		windows.CREATE_NEW_CONSOLE|windows.EXTENDED_STARTUPINFO_PRESENT, nil, workingDirectory, &startup.StartupInfo, &process)
	if err != nil {
		return 0, fmt.Errorf("start %s: %w", target, err)
	}
	_ = windows.CloseHandle(process.Thread)
	return process.Process, nil
}

func (i *instance) readiness() (string, error) {
	defer func() { _ = i.ready.Close() }()
	line, err := bufio.NewReader(io.LimitReader(i.ready, 256)).ReadString('\n')
	if err != nil {
		if len(line) > 0 {
			return "", errors.New("incomplete RailKeeper readiness address")
		}
		return "", err
	}
	address := strings.TrimSuffix(line, "\n")
	parsed, err := url.Parse(address)
	if err != nil || parsed.Scheme != "http" || parsed.User != nil || parsed.Path != "" ||
		parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" {
		return "", errors.New("invalid RailKeeper readiness address")
	}
	host, hostErr := netip.ParseAddr(parsed.Hostname())
	port, portErr := strconv.ParseUint(parsed.Port(), 10, 16)
	if hostErr != nil || !host.IsLoopback() || portErr != nil || port == 0 {
		return "", errors.New("RailKeeper readiness address is not a loopback endpoint")
	}
	return address, nil
}

func (i *instance) wait() error {
	if _, err := windows.WaitForSingleObject(i.process, windows.INFINITE); err != nil {
		return fmt.Errorf("wait for RailKeeper: %w", err)
	}
	var code uint32
	if err := windows.GetExitCodeProcess(i.process, &code); err != nil {
		return fmt.Errorf("read RailKeeper exit status: %w", err)
	}
	if code != 0 {
		return fmt.Errorf("RailKeeper exited with code 0x%x", code)
	}
	return nil
}
