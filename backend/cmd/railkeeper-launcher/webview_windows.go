package main

import (
	"crypto/sha256"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unsafe"

	"github.com/wailsapp/wails/webview2/pkg/combridge"
	wv "github.com/wailsapp/wails/webview2/pkg/edge"
	"github.com/wailsapp/wails/webview2/webviewloader"
	"golang.org/x/sys/windows"
)

const runtimeDownload = "https://developer.microsoft.com/microsoft-edge/webview2/#download-section"

// Track edits until their controls leave the document or the form is reset.
// Some views save asynchronously, so a still-mounted editor is treated conservatively.
const watchEdits = `(() => {
  if (window.top !== window) return;
  const edited = new Set();
  const mark = (event) => {
    const field = event.target;
    if (field instanceof Element &&
        field.matches('input:not([type=search]), textarea, select, [contenteditable]')) {
      edited.add(field);
    }
  };
  document.addEventListener('input', mark, true);
  document.addEventListener('change', mark, true);
  document.addEventListener('reset', (event) => {
    for (const field of edited) {
      if (field.form === event.target) edited.delete(field);
    }
  }, true);
  new MutationObserver(() => {
    for (const field of edited) {
      if (!field.isConnected) edited.delete(field);
    }
  }).observe(document, { childList: true, subtree: true });
  Object.defineProperty(window, '__railKeeperHasEdits', { value: () => edited.size > 0 });
})();`

func webViewProfile() (string, error) {
	if version, err := webviewloader.GetAvailableCoreWebView2BrowserVersionString(""); err != nil || version == "" {
		return "", fmt.Errorf("Microsoft Edge WebView2 Runtime fehlt oder ist nicht verfügbar. Installieren Sie die Evergreen Runtime oder verwenden Sie start-railkeeper.bat.\nMicrosoft Edge WebView2 Runtime is missing or unavailable. Install the Evergreen Runtime or use start-railkeeper.bat.\n\n%s", runtimeDownload)
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	data := strings.TrimSpace(os.Getenv("RAILKEEPER_DATA_DIR"))
	if data == "" {
		data = filepath.Join(cache, "RailKeeper", "data")
	}
	profile := sha256.Sum256([]byte(strings.ToLower(filepath.Clean(data))))
	return filepath.Join(cache, "RailKeeper", "WebView2", fmt.Sprintf("%x", profile[:16])), nil
}

func (w *desktopWindow) createWebView(profile string) error {
	if err := webviewloader.CreateCoreWebView2EnvironmentWithOptions(w, webviewloader.WithUserDataFolder(profile)); err != nil {
		return fmt.Errorf("create WebView2 environment: %w", err)
	}
	return nil
}

func (w *desktopWindow) EnvironmentCompleted(code webviewloader.HRESULT, env *webviewloader.ICoreWebView2Environment) webviewloader.HRESULT {
	if w.stopping {
		return 0
	}
	if code < 0 || env == nil {
		w.fail(fmt.Errorf("create WebView2 environment: HRESULT 0x%08x", uint32(code)))
		return 0
	}
	handler := combridge.New[controllerCompleted](w)
	defer handler.Close()
	if err := callCOMMethod(unsafe.Pointer(env), createControllerSlot, uintptr(w.hwnd), handler.Ref()); err != nil {
		w.fail(fmt.Errorf("create WebView2 controller: %w", err))
	}
	return 0
}

func (w *desktopWindow) controllerCompleted(code int32, controller *wv.ICoreWebView2Controller) {
	if w.stopping {
		if controller != nil {
			_ = callCOMMethod(unsafe.Pointer(controller), closeControllerSlot)
		}
		return
	}
	if code < 0 || controller == nil {
		w.fail(fmt.Errorf("create WebView2 controller: HRESULT 0x%08x", uint32(code)))
		return
	}
	controller.AddRef()
	w.controller = controller
	view, err := controller.GetCoreWebView2()
	if err != nil {
		w.fail(err)
		return
	}
	w.view = view
	if err := w.configureWebView(); err != nil {
		w.fail(fmt.Errorf("configure WebView2: %w", err))
		return
	}
	w.resize()
	_ = controller.MoveFocus(wv.COREWEBVIEW2_MOVE_FOCUS_REASON_PROGRAMMATIC)
}

func (w *desktopWindow) configureWebView() error {
	settings, err := w.view.GetSettings()
	if err != nil {
		return err
	}
	defer settings.Release()
	if err := settings.PutAreHostObjectsAllowed(false); err != nil {
		return err
	}
	if err := settings.PutIsWebMessageEnabled(false); err != nil {
		return err
	}

	start := combridge.New[navigationStarting](w)
	defer start.Close()
	if err := w.addWebViewEvent(addNavigationStartingSlot, start.Ref()); err != nil {
		return err
	}

	completed := combridge.New[navigationCompleted](w)
	defer completed.Close()
	if err := w.addWebViewEvent(addNavigationCompletedSlot, completed.Ref()); err != nil {
		return err
	}

	popup := combridge.New[newWindowRequested](w)
	defer popup.Close()
	if err := w.addWebViewEvent(addNewWindowRequestedSlot, popup.Ref()); err != nil {
		return err
	}

	script := combridge.New[documentScriptAdded](w)
	defer script.Close()
	text := wide(watchEdits)
	return callCOMMethod(unsafe.Pointer(w.view), addDocumentScriptSlot, uintptr(unsafe.Pointer(text)), script.Ref())
}

func (w *desktopWindow) documentScriptAdded(code int32) {
	if code < 0 {
		w.fail(fmt.Errorf("install WebView2 close guard: HRESULT 0x%08x", uint32(code)))
		return
	}
	postMessage.Call(uintptr(w.hwnd), wmWebViewReady, 0, 0)
}

func (w *desktopWindow) navigate() {
	if w.address != "" && !w.stopping {
		if err := w.view.Navigate(w.address); err != nil {
			w.fail(fmt.Errorf("open RailKeeper in WebView2: %w", err))
		}
	}
}

func (w *desktopWindow) navigationCompleted(args *navigationCompletedArgs) {
	if w.stopping || w.pageLoaded {
		return
	}
	ok, err := comBool(unsafe.Pointer(args), args.vtbl.getIsSuccess)
	if err != nil || !ok {
		var status int32
		_ = callCOM(args.vtbl.getErrorStatus, unsafe.Pointer(args), uintptr(unsafe.Pointer(&status)))
		w.fail(fmt.Errorf("load RailKeeper in WebView2: status %d", status))
		return
	}
	w.pageLoaded = true
	w.setTitle(windowTitle)
}

func (w *desktopWindow) sameOrigin(target *url.URL) bool {
	return target.User == nil && target.Scheme+"://"+target.Host == w.address
}

func (w *desktopWindow) navigationStarting(args *navigationStartingArgs) {
	uri, err := comString(unsafe.Pointer(args), args.vtbl.getURI)
	target, parseErr := url.Parse(uri)
	if !w.stopping && err == nil && parseErr == nil && w.sameOrigin(target) {
		return
	}
	if cancelErr := callCOM(args.vtbl.putCancel, unsafe.Pointer(args), 1); cancelErr != nil {
		w.fail(cancelErr)
		return
	}
	userInitiated, _ := comBool(unsafe.Pointer(args), args.vtbl.getIsUserInitiated)
	redirect, _ := comBool(unsafe.Pointer(args), args.vtbl.getIsRedirected)
	if !w.stopping && err == nil && parseErr == nil && userInitiated && !redirect {
		w.openExternal(target)
	}
}

func (w *desktopWindow) newWindowRequested(args *newWindowArgs) {
	uri, err := comString(unsafe.Pointer(args), args.vtbl.getURI)
	target, parseErr := url.Parse(uri)
	userInitiated, _ := comBool(unsafe.Pointer(args), args.vtbl.getIsUserInitiated)
	// The print views open a blank document and populate it through window.open.
	if !w.stopping && err == nil && parseErr == nil && userInitiated && (uri == "about:blank" || w.sameOrigin(target)) {
		return
	}
	if handleErr := callCOM(args.vtbl.putHandled, unsafe.Pointer(args), 1); handleErr != nil {
		w.fail(handleErr)
		return
	}
	if !w.stopping && err == nil && parseErr == nil && userInitiated {
		w.openExternal(target)
	}
}

func (w *desktopWindow) openExternal(target *url.URL) {
	if (target.Scheme != "https" && target.Scheme != "http") || target.Hostname() == "" || target.User != nil {
		return
	}
	address, err := windows.UTF16PtrFromString(target.String())
	if err == nil {
		hwnd := w.hwnd
		go func() {
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()
			// The default protocol handler may be implemented as a shell extension.
			if err := initializeSTA(); err != nil {
				return
			}
			defer windows.CoUninitialize()
			_ = windows.ShellExecute(hwnd, wide("open"), address, nil, nil, windows.SW_SHOWNORMAL)
		}()
	}
}

func (w *desktopWindow) requestClose() {
	if w.stopping || w.closePending {
		return
	}
	if !w.pageLoaded {
		w.beginShutdown()
		return
	}
	w.closePending = true
	w.closeRequest++
	request := &closeRequest{window: w, id: w.closeRequest}
	script := combridge.New[scriptExecuted](request)
	defer script.Close()
	query := wide("typeof window.__railKeeperHasEdits !== 'function' || window.__railKeeperHasEdits()")
	err := callCOMMethod(unsafe.Pointer(w.view), executeScriptSlot, uintptr(unsafe.Pointer(query)), script.Ref())
	if err != nil {
		w.confirmClose(request.id, true)
		return
	}
	if timer, _, _ := setTimer.Call(uintptr(w.hwnd), request.id, 2000, 0); timer == 0 {
		w.confirmClose(request.id, true)
	}
}

type closeRequest struct {
	window *desktopWindow
	id     uintptr
}

func (r *closeRequest) scriptExecuted(code int32, result *uint16) {
	// WebView2 callbacks must return before entering a nested modal message loop.
	confirm := uintptr(1)
	if code >= 0 && windows.UTF16PtrToString(result) == "false" {
		confirm = 0
	}
	postMessage.Call(uintptr(r.window.hwnd), wmConfirmClose, r.id, confirm)
}

func (w *desktopWindow) confirmClose(request uintptr, confirm bool) {
	if !w.closePending || request != w.closeRequest || w.stopping {
		return
	}
	killTimer.Call(uintptr(w.hwnd), request)
	defer func() { w.closePending = false }()
	if confirm {
		answer, _ := windows.MessageBox(windows.HWND(w.hwnd), wide("Ungespeicherte Eingaben können verloren gehen. RailKeeper beenden?\nUnsaved changes may be lost. Close RailKeeper?"),
			wide(windowTitle), windows.MB_YESNO|windows.MB_ICONWARNING|windows.MB_DEFBUTTON2)
		const idYes = 6
		if answer != idYes {
			return
		}
	}
	w.beginShutdown()
}
