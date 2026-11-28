package main

import (
	"fmt"
	"os"
	"path/filepath"
	"unsafe"

	"golang.org/x/sys/windows"
)

func main() {
	if err := startRailKeeper(); err != nil {
		message, _ := windows.UTF16PtrFromString("RailKeeper konnte nicht gestartet werden. / Could not start RailKeeper.\n\n" + err.Error())
		title, _ := windows.UTF16PtrFromString("RailKeeper")
		_, _ = windows.MessageBox(0, message, title, windows.MB_OK|windows.MB_ICONERROR)
		os.Exit(1)
	}
}

func startRailKeeper() error {
	launcherPath, err := os.Executable()
	if err != nil {
		return err
	}
	directory := filepath.Dir(launcherPath)
	target := filepath.Join(directory, "RailKeeper.exe")
	executable, err := windows.UTF16PtrFromString(target)
	if err != nil {
		return err
	}
	commandLine, err := windows.UTF16PtrFromString(windows.ComposeCommandLine([]string{target, "--standalone"}))
	if err != nil {
		return err
	}
	workingDirectory, err := windows.UTF16PtrFromString(directory)
	if err != nil {
		return err
	}
	startup := windows.StartupInfo{
		Cb:         uint32(unsafe.Sizeof(windows.StartupInfo{})),
		Flags:      windows.STARTF_USESHOWWINDOW,
		ShowWindow: windows.SW_SHOWMINNOACTIVE,
	}
	var process windows.ProcessInformation
	err = windows.CreateProcess(executable, commandLine, nil, nil, false,
		windows.CREATE_NEW_CONSOLE, nil, workingDirectory, &startup, &process)
	if err != nil {
		return fmt.Errorf("start %s: %w", target, err)
	}
	_ = windows.CloseHandle(process.Thread)
	_ = windows.CloseHandle(process.Process)
	return nil
}
