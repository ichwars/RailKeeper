package main

import (
	"syscall"
	"unsafe"

	"github.com/wailsapp/wails/webview2/pkg/combridge"
	wv "github.com/wailsapp/wails/webview2/pkg/edge"
	"golang.org/x/sys/windows"
)

// Adapter for the Win32 WebView2 callbacks and methods that aren't wrapped by
// pkg/edge. Interface identifiers and vtable order come from Microsoft's WebView2.h.
type controllerCompleted interface {
	controllerCompleted(int32, *wv.ICoreWebView2Controller)
}

type navigationStarting interface {
	navigationStarting(*navigationStartingArgs)
}

type navigationCompleted interface {
	navigationCompleted(*navigationCompletedArgs)
}

type newWindowRequested interface {
	newWindowRequested(*newWindowArgs)
}

type documentScriptAdded interface {
	documentScriptAdded(int32)
}

type scriptExecuted interface {
	scriptExecuted(int32, *uint16)
}

func init() {
	combridge.RegisterVTable[combridge.IUnknown, controllerCompleted](
		"{6c4819f3-c9b7-4260-8127-c9f5bde7f68c}",
		func(this uintptr, code int32, controller *wv.ICoreWebView2Controller) uintptr {
			combridge.Resolve[controllerCompleted](this).controllerCompleted(code, controller)
			return 0
		})
	combridge.RegisterVTable[combridge.IUnknown, navigationStarting](
		"{9adbe429-f36d-432b-9ddc-f8881fbd76e3}",
		func(this uintptr, _ *wv.ICoreWebView2, args *navigationStartingArgs) uintptr {
			combridge.Resolve[navigationStarting](this).navigationStarting(args)
			return 0
		})
	combridge.RegisterVTable[combridge.IUnknown, navigationCompleted](
		"{d33a35bf-1c49-4f98-93ab-006e0533fe1c}",
		func(this uintptr, _ *wv.ICoreWebView2, args *navigationCompletedArgs) uintptr {
			combridge.Resolve[navigationCompleted](this).navigationCompleted(args)
			return 0
		})
	combridge.RegisterVTable[combridge.IUnknown, newWindowRequested](
		"{d4c185fe-c81c-4989-97af-2d3fa7ab5651}",
		func(this uintptr, _ *wv.ICoreWebView2, args *newWindowArgs) uintptr {
			combridge.Resolve[newWindowRequested](this).newWindowRequested(args)
			return 0
		})
	combridge.RegisterVTable[combridge.IUnknown, documentScriptAdded](
		"{b99369f3-9b11-47b5-bc6f-8e7895fcea17}",
		func(this uintptr, code int32, _ *uint16) uintptr {
			combridge.Resolve[documentScriptAdded](this).documentScriptAdded(code)
			return 0
		})
	combridge.RegisterVTable[combridge.IUnknown, scriptExecuted](
		"{49511172-cc67-4bca-9923-137112f4c4cc}",
		func(this uintptr, code int32, result *uint16) uintptr {
			combridge.Resolve[scriptExecuted](this).scriptExecuted(code, result)
			return 0
		})
}

type navigationStartingArgs struct {
	vtbl *struct {
		combridge.IUnknownVtbl
		getURI, getIsUserInitiated, getIsRedirected wv.ComProc
		getRequestHeaders, getCancel, putCancel     wv.ComProc
	}
}

type navigationCompletedArgs struct {
	vtbl *struct {
		combridge.IUnknownVtbl
		getIsSuccess, getErrorStatus wv.ComProc
	}
}

type newWindowArgs struct {
	vtbl *struct {
		combridge.IUnknownVtbl
		getURI, putNewWindow, getNewWindow         wv.ComProc
		putHandled, getHandled, getIsUserInitiated wv.ComProc
	}
}

func (w *desktopWindow) addWebViewEvent(method int, handler uintptr) error {
	var token int64
	return callCOMMethod(unsafe.Pointer(w.view), method, handler, uintptr(unsafe.Pointer(&token)))
}

// Slots in the base ICoreWebView2Environment, ICoreWebView2Controller and
// ICoreWebView2 vtables, counting the three IUnknown methods.
const (
	createControllerSlot       = 3
	putBoundsSlot              = 6
	closeControllerSlot        = 24
	addNavigationStartingSlot  = 7
	addNavigationCompletedSlot = 15
	addDocumentScriptSlot      = 27
	executeScriptSlot          = 29
	addNewWindowRequestedSlot  = 44
)

//go:uintptrescapes
func callCOMMethod(self unsafe.Pointer, method int, args ...uintptr) error {
	vtbl := *(*unsafe.Pointer)(self)
	methodAddress := unsafe.Add(vtbl, method*int(unsafe.Sizeof(uintptr(0))))
	return callCOM(*(*wv.ComProc)(methodAddress), self, args...)
}

//go:uintptrescapes
func callCOM(proc wv.ComProc, self unsafe.Pointer, args ...uintptr) error {
	result, _, _ := proc.Call(append([]uintptr{uintptr(self)}, args...)...)
	if int32(result) < 0 {
		return syscall.Errno(uint32(result))
	}
	return nil
}

func comBool(self unsafe.Pointer, proc wv.ComProc) (bool, error) {
	var value int32
	err := callCOM(proc, self, uintptr(unsafe.Pointer(&value)))
	return value != 0, err
}

func comString(self unsafe.Pointer, proc wv.ComProc) (string, error) {
	var value *uint16
	if err := callCOM(proc, self, uintptr(unsafe.Pointer(&value))); err != nil {
		return "", err
	}
	defer windows.CoTaskMemFree(unsafe.Pointer(value))
	return windows.UTF16PtrToString(value), nil
}
