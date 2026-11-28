package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"unsafe"

	wv "github.com/wailsapp/wails/webview2/pkg/edge"
	"golang.org/x/sys/windows"
)

const (
	windowTitle   = "RailKeeper"
	windowClass   = "RailKeeperDesktopWindow"
	startingTitle = "RailKeeper startet … / Starting RailKeeper …"
	stoppingTitle = "RailKeeper wird beendet … / Closing RailKeeper …"

	wmDestroy       = 0x0002
	wmMove          = 0x0003
	wmSize          = 0x0005
	wmSetFocus      = 0x0007
	wmClose         = 0x0010
	wmTimer         = 0x0113
	wmApp           = 0x8000
	wmServerReady   = wmApp + 1
	wmServerExited  = wmApp + 2
	wmWebViewFailed = wmApp + 3
	wmConfirmClose  = wmApp + 4
	wmWebViewReady  = wmApp + 5

	cwUseDefault       = 0x80000000
	wsOverlappedWindow = 0x00cf0000
)

var (
	user32           = windows.NewLazySystemDLL("user32.dll")
	registerClass    = user32.NewProc("RegisterClassExW")
	createWindow     = user32.NewProc("CreateWindowExW")
	defWindowProc    = user32.NewProc("DefWindowProcW")
	destroyWindow    = user32.NewProc("DestroyWindow")
	showWindow       = user32.NewProc("ShowWindow")
	getMessage       = user32.NewProc("GetMessageW")
	translateMessage = user32.NewProc("TranslateMessage")
	dispatchMessage  = user32.NewProc("DispatchMessageW")
	postMessage      = user32.NewProc("PostMessageW")
	postQuitMessage  = user32.NewProc("PostQuitMessage")
	setTimer         = user32.NewProc("SetTimer")
	killTimer        = user32.NewProc("KillTimer")
	setWindowText    = user32.NewProc("SetWindowTextW")
	getClientRect    = user32.NewProc("GetClientRect")
	loadCursor       = user32.NewProc("LoadCursorW")
	loadIcon         = user32.NewProc("LoadIconW")
	windowProc       = windows.NewCallback(handleWindowMessage)

	// One desktop window per launcher process. Its state belongs to the STA thread.
	activeWindow *desktopWindow
)

type windowClassEx struct {
	size, style              uint32
	proc                     uintptr
	classExtra, windowExtra  int32
	module                   windows.Handle
	icon, cursor, background uintptr
	menu, name               *uint16
	smallIcon                uintptr
}

type windowMessage struct {
	hwnd           windows.Handle
	message        uint32
	wParam, lParam uintptr
	time           uint32
	point          struct{ X, Y int32 }
	private        uint32
}

type readinessResult struct {
	address string
	err     error
}

type desktopWindow struct {
	directory    string
	hwnd         windows.Handle
	instance     *instance
	ready        chan readinessResult
	exited       chan error
	exitObserved bool
	stopping     bool
	err          error

	address    string
	controller *wv.ICoreWebView2Controller
	view       *wv.ICoreWebView2

	pageLoaded   bool
	closePending bool
	closeRequest uintptr
}

func startRailKeeper() error {
	profile, err := webViewProfile()
	if err != nil {
		return err
	}
	path, err := os.Executable()
	if err != nil {
		return err
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := windows.CoInitializeEx(0, windows.COINIT_APARTMENTTHREADED); err != nil && !errors.Is(err, syscall.Errno(windows.S_FALSE)) {
		return fmt.Errorf("initialize WebView2 COM thread: %w", err)
	}
	defer windows.CoUninitialize()

	window := &desktopWindow{directory: filepath.Dir(path), ready: make(chan readinessResult, 1), exited: make(chan error, 1)}
	activeWindow = window
	defer func() { activeWindow = nil }()
	if err := window.create(); err != nil {
		return err
	}
	defer window.destroy()

	// Initialize WebView2 before launching the server. A broken runtime cannot
	// leave a newly started RailKeeper instance behind.
	if err := window.createWebView(profile); err != nil {
		return err
	}
	showWindow.Call(uintptr(window.hwnd), windows.SW_SHOWNORMAL)
	defer func() {
		if window.instance == nil {
			return
		}
		_ = window.instance.stop.Close()
		if !window.exitObserved {
			<-window.exited
		}
		_ = windows.CloseHandle(window.instance.process)
	}()

	var message windowMessage
	for {
		result, _, err := getMessage.Call(uintptr(unsafe.Pointer(&message)), 0, 0, 0)
		switch int32(result) {
		case -1:
			return fmt.Errorf("read desktop window message: %w", err)
		case 0:
			return window.err
		}
		translateMessage.Call(uintptr(unsafe.Pointer(&message)))
		dispatchMessage.Call(uintptr(unsafe.Pointer(&message)))
	}
}

func (w *desktopWindow) startServer() error {
	child, err := launchInstance(w.directory)
	if err != nil {
		return err
	}
	w.instance = child
	hwnd := w.hwnd
	go func() {
		address, err := child.readiness()
		w.ready <- readinessResult{address: address, err: err}
		postMessage.Call(uintptr(hwnd), wmServerReady, 0, 0)
	}()
	go func() {
		w.exited <- child.wait()
		postMessage.Call(uintptr(hwnd), wmServerExited, 0, 0)
	}()
	return nil
}

func (w *desktopWindow) create() error {
	var module windows.Handle
	if err := windows.GetModuleHandleEx(windows.GET_MODULE_HANDLE_EX_FLAG_UNCHANGED_REFCOUNT, nil, &module); err != nil {
		return err
	}
	name := wide(windowClass)
	cursor, _, _ := loadCursor.Call(0, 32512)       // IDC_ARROW
	icon, _, _ := loadIcon.Call(uintptr(module), 1) // Packaged application icon.
	class := windowClassEx{
		proc: windowProc, module: module, name: name,
		icon: icon, smallIcon: icon, cursor: cursor, background: 6, // COLOR_WINDOW + 1
	}
	class.size = uint32(unsafe.Sizeof(class))
	if result, _, err := registerClass.Call(uintptr(unsafe.Pointer(&class))); result == 0 {
		return fmt.Errorf("register RailKeeper window: %w", err)
	}
	title := wide(startingTitle)
	hwnd, _, err := createWindow.Call(0, uintptr(unsafe.Pointer(name)), uintptr(unsafe.Pointer(title)),
		wsOverlappedWindow, cwUseDefault, cwUseDefault, 1160, 800, 0, 0, uintptr(module), 0)
	if hwnd == 0 {
		return fmt.Errorf("create RailKeeper window: %w", err)
	}
	w.hwnd = windows.Handle(hwnd)
	return nil
}

func (w *desktopWindow) destroy() {
	if w.controller != nil {
		_ = callCOMMethod(unsafe.Pointer(w.controller), closeControllerSlot)
		if w.view != nil {
			w.view.Release()
			w.view = nil
		}
		w.controller.Release()
		w.controller = nil
	}
	if w.hwnd != 0 {
		destroyWindow.Call(uintptr(w.hwnd))
		w.hwnd = 0
	}
}

func handleWindowMessage(hwnd windows.Handle, message uint32, wParam, lParam uintptr) uintptr {
	w := activeWindow
	if w != nil && hwnd == w.hwnd {
		switch message {
		case wmDestroy:
			postQuitMessage.Call(0)
			return 0
		case wmSize:
			w.resize()
			return 0
		case wmMove:
			if w.controller != nil {
				_ = w.controller.NotifyParentWindowPositionChanged()
			}
		case wmSetFocus:
			if w.controller != nil && !w.stopping {
				_ = w.controller.MoveFocus(wv.COREWEBVIEW2_MOVE_FOCUS_REASON_PROGRAMMATIC)
			}
		case wmClose:
			w.requestClose()
			return 0
		case wmWebViewReady:
			if !w.stopping {
				if err := w.startServer(); err != nil {
					w.fail(err)
				}
			}
			return 0
		case wmServerReady:
			result := <-w.ready
			if !w.stopping {
				if result.err == nil {
					w.address = result.address
					w.navigate()
				} else if !errors.Is(result.err, io.EOF) {
					w.fail(fmt.Errorf("read RailKeeper readiness: %w", result.err))
				}
			}
			return 0
		case wmServerExited:
			err := <-w.exited
			w.exitObserved = true
			w.err = errors.Join(w.err, err)
			w.destroy()
			return 0
		case wmWebViewFailed:
			w.beginShutdown()
			return 0
		case wmTimer:
			w.confirmClose(wParam, true)
			return 0
		case wmConfirmClose:
			w.confirmClose(wParam, lParam != 0)
			return 0
		}
	}
	result, _, _ := defWindowProc.Call(uintptr(hwnd), uintptr(message), wParam, lParam)
	return result
}

func (w *desktopWindow) resize() {
	var bounds struct{ Left, Top, Right, Bottom int32 }
	if w.controller != nil {
		getClientRect.Call(uintptr(w.hwnd), uintptr(unsafe.Pointer(&bounds)))
		_ = callCOMMethod(unsafe.Pointer(w.controller), putBoundsSlot, uintptr(unsafe.Pointer(&bounds)))
	}
}

func (w *desktopWindow) setTitle(title string) {
	name := wide(title)
	setWindowText.Call(uintptr(w.hwnd), uintptr(unsafe.Pointer(name)))
}

func (w *desktopWindow) fail(err error) {
	if !w.stopping {
		w.err = err
		// Leave the WebView2 callback before closing the controller or opening an error dialog.
		postMessage.Call(uintptr(w.hwnd), wmWebViewFailed, 0, 0)
	}
}

func (w *desktopWindow) beginShutdown() {
	if w.stopping {
		return
	}
	w.stopping = true
	w.setTitle(stoppingTitle)
	if w.controller != nil {
		_ = w.controller.PutIsVisible(false)
	}
	if w.instance == nil {
		w.destroy()
		return
	}
	_ = w.instance.stop.Close()
}

func wide(text string) *uint16 {
	ptr, _ := windows.UTF16PtrFromString(text)
	return ptr
}
