//go:build windows

package main

import (
	"runtime"
	"sync"
	"syscall"
	"unsafe"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	trayCallbackMessage = 0x8000 + 77
	wmDestroy           = 0x0002
	wmClose             = 0x0010
	wmLButtonUp         = 0x0202
	wmLButtonDblClk     = 0x0203
	wmRButtonUp         = 0x0205
	wmNull              = 0x0000
	nimAdd              = 0x00000000
	nimModify           = 0x00000001
	nimDelete           = 0x00000002
	nifMessage          = 0x00000001
	nifIcon             = 0x00000002
	nifTip              = 0x00000004
	nifInfo             = 0x00000010
	niifInfo            = 0x00000001
	niifWarning         = 0x00000002
	mfString            = 0x00000000
	mfSeparator         = 0x00000800
	tpmRightButton      = 0x0002
	tpmReturnCmd        = 0x0100
	idiApplication      = 32512
	trayOpen            = 9101
	trayStart           = 9102
	trayStop            = 9103
	trayLogs            = 9104
	trayExit            = 9105
)

type trayPoint struct{ X, Y int32 }
type trayMessage struct {
	HWnd           syscall.Handle
	Message        uint32
	WParam, LParam uintptr
	Time           uint32
	Pt             trayPoint
	LPrivate       uint32
}
type trayWndClassEx struct {
	CbSize                 uint32
	Style                  uint32
	LpfnWndProc            uintptr
	CbClsExtra, CbWndExtra int32
	HInstance              syscall.Handle
	HIcon                  syscall.Handle
	HCursor                syscall.Handle
	HbrBackground          syscall.Handle
	LpszMenuName           *uint16
	LpszClassName          *uint16
	HIconSm                syscall.Handle
}
type notifyIconData struct {
	CbSize           uint32
	HWnd             syscall.Handle
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            syscall.Handle
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	UVersion         uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
	GuidItem         [16]byte
	HBalloonIcon     syscall.Handle
}

var (
	trayUser32              = syscall.NewLazyDLL("user32.dll")
	trayShell32             = syscall.NewLazyDLL("shell32.dll")
	trayKernel32            = syscall.NewLazyDLL("kernel32.dll")
	trayRegisterClassEx     = trayUser32.NewProc("RegisterClassExW")
	trayCreateWindowEx      = trayUser32.NewProc("CreateWindowExW")
	trayDefWindowProc       = trayUser32.NewProc("DefWindowProcW")
	trayDestroyWindow       = trayUser32.NewProc("DestroyWindow")
	trayGetMessage          = trayUser32.NewProc("GetMessageW")
	trayTranslateMessage    = trayUser32.NewProc("TranslateMessage")
	trayDispatchMessage     = trayUser32.NewProc("DispatchMessageW")
	trayPostQuitMessage     = trayUser32.NewProc("PostQuitMessage")
	trayPostMessage         = trayUser32.NewProc("PostMessageW")
	trayCreatePopupMenu     = trayUser32.NewProc("CreatePopupMenu")
	trayAppendMenu          = trayUser32.NewProc("AppendMenuW")
	trayTrackPopupMenu      = trayUser32.NewProc("TrackPopupMenu")
	trayDestroyMenu         = trayUser32.NewProc("DestroyMenu")
	trayGetCursorPos        = trayUser32.NewProc("GetCursorPos")
	traySetForegroundWindow = trayUser32.NewProc("SetForegroundWindow")
	trayLoadIcon            = trayUser32.NewProc("LoadIconW")
	trayGetModuleHandle     = trayKernel32.NewProc("GetModuleHandleW")
	trayShellNotifyIcon     = trayShell32.NewProc("Shell_NotifyIconW")
	trayProc                = syscall.NewCallback(trayWindowProc)
	trayMu                  sync.Mutex
	trayApp                 *App
	trayHwnd                syscall.Handle
	trayNID                 notifyIconData
)

func trayPtr(value string) *uint16 { p, _ := syscall.UTF16PtrFromString(value); return p }
func copyTrayText(dst []uint16, value string) {
	encoded := syscall.StringToUTF16(value)
	if len(encoded) > len(dst) {
		encoded = encoded[:len(dst)]
	}
	copy(dst, encoded)
}

func startTray(app *App) {
	trayMu.Lock()
	if trayHwnd != 0 {
		trayMu.Unlock()
		return
	}
	trayApp = app
	trayMu.Unlock()
	ready := make(chan struct{})
	go trayLoop(ready)
	<-ready
}
func stopTray() {
	trayMu.Lock()
	hwnd := trayHwnd
	trayMu.Unlock()
	if hwnd != 0 {
		trayPostMessage.Call(uintptr(hwnd), wmClose, 0, 0)
	}
}
func syncTrayMenu() {}

func trayLoop(ready chan struct{}) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	instance, _, _ := trayGetModuleHandle.Call(0)
	className := trayPtr("NetWatcherTrayWindowV3")
	wc := trayWndClassEx{CbSize: uint32(unsafe.Sizeof(trayWndClassEx{})), LpfnWndProc: trayProc, HInstance: syscall.Handle(instance), LpszClassName: className}
	trayRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc)))
	hwnd, _, _ := trayCreateWindowEx.Call(0, uintptr(unsafe.Pointer(className)), uintptr(unsafe.Pointer(trayPtr("NetWatcher Tray"))), 0, 0, 0, 0, 0, 0, 0, instance, 0)
	if hwnd == 0 {
		close(ready)
		return
	}
	icon, _, _ := trayLoadIcon.Call(0, idiApplication)
	nid := notifyIconData{CbSize: uint32(unsafe.Sizeof(notifyIconData{})), HWnd: syscall.Handle(hwnd), UID: 1, UFlags: nifMessage | nifIcon | nifTip, UCallbackMessage: trayCallbackMessage, HIcon: syscall.Handle(icon)}
	copyTrayText(nid.SzTip[:], "NetWatcher")
	trayShellNotifyIcon.Call(nimAdd, uintptr(unsafe.Pointer(&nid)))
	trayMu.Lock()
	trayHwnd = syscall.Handle(hwnd)
	trayNID = nid
	trayMu.Unlock()
	close(ready)
	var msg trayMessage
	for {
		r, _, _ := trayGetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		trayTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		trayDispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

func trayWindowProc(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case trayCallbackMessage:
		switch uint32(lParam) {
		case wmLButtonUp, wmLButtonDblClk:
			showMainWindow()
		case wmRButtonUp:
			showTrayMenu(hwnd)
		}
		return 0
	case wmClose:
		trayDestroyWindow.Call(uintptr(hwnd))
		return 0
	case wmDestroy:
		trayMu.Lock()
		nid := trayNID
		trayHwnd = 0
		trayMu.Unlock()
		trayShellNotifyIcon.Call(nimDelete, uintptr(unsafe.Pointer(&nid)))
		trayPostQuitMessage.Call(0)
		return 0
	}
	r, _, _ := trayDefWindowProc.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return r
}
func showMainWindow() {
	trayMu.Lock()
	app := trayApp
	trayMu.Unlock()
	if app != nil && app.ctx != nil {
		wruntime.WindowShow(app.ctx)
		wruntime.WindowUnminimise(app.ctx)
	}
}
func appendTrayItem(menu uintptr, id uintptr, text string) {
	trayAppendMenu.Call(menu, mfString, id, uintptr(unsafe.Pointer(trayPtr(text))))
}
func showTrayMenu(hwnd syscall.Handle) {
	trayMu.Lock()
	app := trayApp
	trayMu.Unlock()
	if app == nil {
		return
	}
	snapshot := app.GetSnapshot()
	menu, _, _ := trayCreatePopupMenu.Call()
	appendTrayItem(menu, trayOpen, "Open NetWatcher")
	trayAppendMenu.Call(menu, mfSeparator, 0, 0)
	if snapshot.Monitoring {
		appendTrayItem(menu, trayStop, "Stop monitoring")
	} else {
		appendTrayItem(menu, trayStart, "Start monitoring")
	}
	appendTrayItem(menu, trayLogs, "Open log folder")
	trayAppendMenu.Call(menu, mfSeparator, 0, 0)
	appendTrayItem(menu, trayExit, "Exit")
	var point trayPoint
	trayGetCursorPos.Call(uintptr(unsafe.Pointer(&point)))
	traySetForegroundWindow.Call(uintptr(hwnd))
	cmd, _, _ := trayTrackPopupMenu.Call(menu, tpmRightButton|tpmReturnCmd, uintptr(point.X), uintptr(point.Y), 0, uintptr(hwnd), 0)
	trayDestroyMenu.Call(menu)
	trayPostMessage.Call(uintptr(hwnd), wmNull, 0, 0)
	switch cmd {
	case trayOpen:
		showMainWindow()
	case trayStart:
		app.StartMonitoring()
	case trayStop:
		app.StopMonitoring()
	case trayLogs:
		_ = app.OpenLogsFolder()
	case trayExit:
		app.Quit()
	}
}

func showTrayNotification(title, message, level string) bool {
	trayMu.Lock()
	if trayHwnd == 0 {
		trayMu.Unlock()
		return false
	}
	nid := trayNID
	trayMu.Unlock()
	nid.UFlags = nifInfo
	nid.DwInfoFlags = niifInfo
	if level == "warning" || level == "error" {
		nid.DwInfoFlags = niifWarning
	}
	copyTrayText(nid.SzInfoTitle[:], title)
	copyTrayText(nid.SzInfo[:], message)
	result, _, _ := trayShellNotifyIcon.Call(nimModify, uintptr(unsafe.Pointer(&nid)))
	return result != 0
}
