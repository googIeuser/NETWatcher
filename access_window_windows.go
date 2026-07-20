//go:build windows

package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

const (
	awStatus        = 8301
	awPortLabel     = 8302
	awPort          = 8303
	awFragmentLabel = 8304
	awFragment      = 8305
	awSystemProxy   = 8306
	awAutoStart     = 8307
	awURLLabel      = 8308
	awURL           = 8309
	awStartStop     = 8310
	awTest          = 8311
	awProxySettings = 8312
	awClose         = 8313
	awInfo          = 8314
)

type AccessWindow struct {
	hwnd     syscall.Handle
	parent   *App
	controls map[int]syscall.Handle
}

var globalAccessWindow *AccessWindow
var accessWindowProcCallback = syscall.NewCallback(accessWindowProc)

func (w *AccessWindow) syncStatus() {
	w.parent.mu.RLock()
	enabled := w.parent.accessProxyEnabled
	proxy := w.parent.accessProxy
	w.parent.mu.RUnlock()
	if enabled && proxy != nil {
		setText(w.controls[awStatus], "Status: Running on "+proxy.Address())
		setText(w.controls[awStartStop], "Stop Access Mode")
		enable(w.controls[awPort], false)
		enable(w.controls[awFragment], false)
		enable(w.controls[awSystemProxy], false)
	} else {
		setText(w.controls[awStatus], "Status: Stopped")
		setText(w.controls[awStartStop], "Start Access Mode")
		enable(w.controls[awPort], true)
		enable(w.controls[awFragment], true)
		enable(w.controls[awSystemProxy], true)
	}
}

func (w *AccessWindow) buildControls() {
	cfg := w.parent.config
	w.controls[awStatus] = createControl(w.hwnd, "STATIC", "", WS_CHILD|WS_VISIBLE|SS_LEFT, awStatus)
	w.controls[awInfo] = createControl(w.hwnd, "STATIC", "Experimental browser/proxy-aware access mode. It uses encrypted DNS-over-HTTPS and fragments the first TLS bytes through a local proxy. It does not use ICMP ping, a VPN, a packet driver or administrator rights. It may not work against every type of block.", WS_CHILD|WS_VISIBLE|SS_LEFT, awInfo)
	w.controls[awPortLabel] = createControl(w.hwnd, "STATIC", "Local proxy port:", WS_CHILD|WS_VISIBLE|SS_LEFT, awPortLabel)
	w.controls[awPort] = createControl(w.hwnd, "EDIT", strconv.Itoa(cfg.AccessPort), WS_CHILD|WS_VISIBLE|WS_BORDER|WS_TABSTOP|ES_LEFT, awPort)
	w.controls[awFragmentLabel] = createControl(w.hwnd, "STATIC", "TLS fragment size:", WS_CHILD|WS_VISIBLE|SS_LEFT, awFragmentLabel)
	w.controls[awFragment] = createControl(w.hwnd, "COMBOBOX", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|WS_VSCROLL|CBS_DROPDOWNLIST, awFragment)
	for _, size := range []string{"1 byte", "2 bytes", "4 bytes", "8 bytes"} {
		comboAdd(w.controls[awFragment], size)
	}
	fragmentIndex := 0
	switch cfg.AccessFragmentSize {
	case 2:
		fragmentIndex = 1
	case 4:
		fragmentIndex = 2
	case 8:
		fragmentIndex = 3
	}
	comboSet(w.controls[awFragment], fragmentIndex)
	w.controls[awSystemProxy] = createControl(w.hwnd, "BUTTON", "Apply the local proxy to Windows (browsers and proxy-aware apps)", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_AUTOCHECKBOX, awSystemProxy)
	setCheck(w.controls[awSystemProxy], cfg.AccessUseSystemProxy)
	w.controls[awAutoStart] = createControl(w.hwnd, "BUTTON", "Start Access Mode automatically when NetWatcher opens", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_AUTOCHECKBOX, awAutoStart)
	setCheck(w.controls[awAutoStart], cfg.AccessAutoStart)
	w.controls[awURLLabel] = createControl(w.hwnd, "STATIC", "HTTPS test URL:", WS_CHILD|WS_VISIBLE|SS_LEFT, awURLLabel)
	w.controls[awURL] = createControl(w.hwnd, "EDIT", "https://example.com/", WS_CHILD|WS_VISIBLE|WS_BORDER|WS_TABSTOP|ES_LEFT|ES_AUTOHSCROLL, awURL)
	w.controls[awStartStop] = createControl(w.hwnd, "BUTTON", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_DEFPUSHBUTTON, awStartStop)
	w.controls[awTest] = createControl(w.hwnd, "BUTTON", "Test via proxy", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, awTest)
	w.controls[awProxySettings] = createControl(w.hwnd, "BUTTON", "Windows proxy settings", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, awProxySettings)
	w.controls[awClose] = createControl(w.hwnd, "BUTTON", "Close", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, awClose)
	w.syncStatus()
}

func (w *AccessWindow) layout(width, height int32) {
	move(w.controls[awStatus], 24, 18, width-48, 24)
	move(w.controls[awInfo], 24, 48, width-48, 72)
	move(w.controls[awPortLabel], 24, 132, 180, 24)
	move(w.controls[awPort], 210, 128, 100, 28)
	move(w.controls[awFragmentLabel], 332, 132, 150, 24)
	move(w.controls[awFragment], 480, 128, 120, 150)
	move(w.controls[awSystemProxy], 24, 172, width-48, 28)
	move(w.controls[awAutoStart], 24, 206, width-48, 28)
	move(w.controls[awURLLabel], 24, 246, 130, 24)
	move(w.controls[awURL], 155, 242, width-179, 28)
	move(w.controls[awStartStop], 24, 292, 145, 32)
	move(w.controls[awTest], 177, 292, 115, 32)
	move(w.controls[awProxySettings], 300, 292, 160, 32)
	move(w.controls[awClose], width-114, height-52, 90, 30)
}

func accessFragmentFromCombo(index int) int {
	switch index {
	case 1:
		return 2
	case 2:
		return 4
	case 3:
		return 8
	default:
		return 1
	}
}

func (w *AccessWindow) saveOptions() bool {
	port, err := strconv.Atoi(strings.TrimSpace(getText(w.controls[awPort])))
	if err != nil || port < 1 || port > 65535 {
		messageBoxOwned(w.hwnd, appName, "Enter a valid proxy port between 1 and 65535.", MB_OK|MB_ICONERROR)
		return false
	}
	w.parent.mu.Lock()
	w.parent.config.AccessPort = port
	w.parent.config.AccessFragmentSize = accessFragmentFromCombo(comboGet(w.controls[awFragment]))
	w.parent.config.AccessUseSystemProxy = isChecked(w.controls[awSystemProxy])
	w.parent.config.AccessAutoStart = isChecked(w.controls[awAutoStart])
	cfg := w.parent.config
	w.parent.mu.Unlock()
	_ = saveConfig(cfg)
	return true
}

func testThroughAccessProxy(proxyAddress, targetURL string) error {
	parsedProxy, err := url.Parse("http://" + proxyAddress)
	if err != nil {
		return err
	}
	client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(parsedProxy), DisableKeepAlives: true}, Timeout: 12 * time.Second}
	request, err := http.NewRequest(http.MethodHead, strings.TrimSpace(targetURL), nil)
	if err != nil {
		return err
	}
	request.Header.Set("User-Agent", "NetWatcher/"+appVersion)
	started := time.Now()
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode >= 500 {
		return fmt.Errorf("server returned HTTP %d", response.StatusCode)
	}
	_ = started
	return nil
}

func accessWindowProc(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	window := globalAccessWindow
	if window == nil {
		result, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
		return result
	}
	switch msg {
	case WM_CREATE:
		window.hwnd = hwnd
		window.buildControls()
		dark := isDarkTheme(window.parent.config.Theme)
		applyWindowDarkMode(hwnd, dark)
		for _, control := range window.controls {
			applyControlTheme(control, dark)
		}
		return 0
	case WM_SIZE:
		window.layout(int32(loword(lParam)), int32(hiword(lParam)))
		return 0
	case WM_ERASEBKGND:
		fillClientBackground(hwnd, syscall.Handle(wParam), isDarkTheme(window.parent.config.Theme))
		return 1
	case WM_CTLCOLORSTATIC, WM_CTLCOLORBTN:
		return themeControlColor(syscall.Handle(wParam), isDarkTheme(window.parent.config.Theme), false)
	case WM_CTLCOLOREDIT:
		return themeControlColor(syscall.Handle(wParam), isDarkTheme(window.parent.config.Theme), true)
	case WM_COMMAND:
		if hiword(wParam) == BN_CLICKED {
			switch int(loword(wParam)) {
			case awStartStop:
				window.parent.mu.RLock()
				enabled := window.parent.accessProxyEnabled
				window.parent.mu.RUnlock()
				if enabled {
					window.parent.stopAccessMode()
				} else if window.saveOptions() {
					window.parent.mu.RLock()
					useSystem := window.parent.config.AccessUseSystemProxy
					window.parent.mu.RUnlock()
					if err := window.parent.startAccessMode(useSystem); err != nil {
						messageBoxOwned(hwnd, appName, "Access Mode could not start:\n\n"+err.Error(), MB_OK|MB_ICONERROR)
					}
				}
				window.syncStatus()
			case awTest:
				window.parent.mu.RLock()
				proxy := window.parent.accessProxy
				enabled := window.parent.accessProxyEnabled
				window.parent.mu.RUnlock()
				if !enabled || proxy == nil {
					messageBoxOwned(hwnd, appName, "Start Access Mode before running the test.", MB_OK|MB_ICONINFORMATION)
					return 0
				}
				testURL := getText(window.controls[awURL])
				proxyAddress := proxy.Address()
				go func(owner syscall.Handle) {
					err := testThroughAccessProxy(proxyAddress, testURL)
					if err != nil {
						messageBoxOwned(owner, appName, "HTTPS proxy test failed:\n\n"+err.Error(), MB_OK|MB_ICONERROR)
					} else {
						messageBoxOwned(owner, appName, "HTTPS request completed successfully through Access Mode. No ICMP ping was used.", MB_OK|MB_ICONINFORMATION)
					}
				}(hwnd)
			case awProxySettings:
				openProxySettings()
			case awClose:
				_ = window.saveOptions()
				procDestroyWindow.Call(uintptr(hwnd))
			}
		}
		return 0
	case WM_CLOSE:
		_ = window.saveOptions()
		procDestroyWindow.Call(uintptr(hwnd))
		return 0
	case WM_DESTROY:
		enable(window.parent.hwnd, true)
		procSetForegroundWindow.Call(uintptr(window.parent.hwnd))
		globalAccessWindow = nil
		return 0
	}
	result, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return result
}

func openAccessWindow(a *App) {
	if globalAccessWindow != nil && globalAccessWindow.hwnd != 0 {
		procSetForegroundWindow.Call(uintptr(globalAccessWindow.hwnd))
		return
	}
	instance, _, _ := procGetModuleHandleW.Call(0)
	className := ptr("NetWatcherAccessWindow")
	icon := loadIconFromFile(runtimeIconPath())
	cursor, _, _ := procLoadCursorW.Call(0, IDC_ARROW)
	windowClass := WNDCLASSEX{CbSize: uint32(unsafe.Sizeof(WNDCLASSEX{})), LpfnWndProc: accessWindowProcCallback, HInstance: syscall.Handle(instance), HIcon: icon, HCursor: syscall.Handle(cursor), HbrBackground: syscall.Handle(COLOR_WINDOW + 1), LpszClassName: className, HIconSm: icon}
	atom, _, registerErr := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&windowClass)))
	if atom == 0 {
		if errno, ok := registerErr.(syscall.Errno); !ok || errno != syscall.Errno(1410) {
			messageBox(appName, "Access Mode window could not open: "+registerErr.Error(), MB_OK|MB_ICONERROR)
			return
		}
	}
	window := &AccessWindow{parent: a, controls: map[int]syscall.Handle{}}
	globalAccessWindow = window
	style := uint32(WS_CAPTION | WS_SYSMENU | WS_CLIPCHILDREN | WS_VISIBLE)
	hwnd, _, createErr := procCreateWindowExW.Call(WS_EX_DLGMODALFRAME|WS_EX_CONTROLPARENT, uintptr(unsafe.Pointer(className)), uintptr(unsafe.Pointer(ptr("NetWatcher Access Mode"))), uintptr(style), 0, 0, 720, 410, uintptr(a.hwnd), 0, instance, 0)
	if hwnd == 0 {
		globalAccessWindow = nil
		messageBox(appName, "Access Mode window error: "+createErr.Error(), MB_OK|MB_ICONERROR)
		return
	}
	procSendMessageW.Call(hwnd, WM_SETICON, ICON_BIG, uintptr(icon))
	procSendMessageW.Call(hwnd, WM_SETICON, ICON_SMALL, uintptr(icon))
	var parentRect RECT
	if ok, _, _ := procGetWindowRect.Call(uintptr(a.hwnd), uintptr(unsafe.Pointer(&parentRect))); ok != 0 {
		x := parentRect.Left + (parentRect.Right-parentRect.Left-720)/2
		y := parentRect.Top + (parentRect.Bottom-parentRect.Top-410)/2
		procMoveWindow.Call(hwnd, uintptr(x), uintptr(y), 720, 410, 0)
	}
	enable(a.hwnd, false)
	procShowWindow.Call(hwnd, SW_SHOW)
	procUpdateWindow.Call(hwnd)
	procSetForegroundWindow.Call(hwnd)
}
