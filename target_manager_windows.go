//go:build windows

package main

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"
)

const (
	tmCombo   = 8201
	tmEdit    = 8202
	tmAdd     = 8203
	tmReplace = 8204
	tmRemove  = 8205
	tmClose   = 8206
	tmInfo    = 8207
)

type TargetManagerWindow struct {
	hwnd     syscall.Handle
	parent   *App
	controls map[int]syscall.Handle
}

var globalTargetManager *TargetManagerWindow
var targetManagerProcCallback = syscall.NewCallback(targetManagerProc)

func (a *App) replaceCustomTarget(oldValue, newValue string) bool {
	oldValue = strings.TrimSpace(oldValue)
	newValue = strings.TrimSpace(newValue)
	if oldValue == "" || newValue == "" {
		return false
	}
	a.mu.Lock()
	oldIndex := findCustomTargetIndex(a.config.CustomTargets, oldValue)
	if oldIndex < 0 {
		a.mu.Unlock()
		return false
	}
	if duplicate := findCustomTargetIndex(a.config.CustomTargets, newValue); duplicate >= 0 && duplicate != oldIndex {
		a.mu.Unlock()
		messageBox(appName, "That target already exists.", MB_OK|MB_ICONWARNING)
		return false
	}
	oldTarget := parseTargetSpec(oldValue)
	newTarget := parseTargetSpec(newValue)
	a.config.CustomTargets[oldIndex] = newValue
	for index, target := range a.targets {
		if target.Custom && strings.EqualFold(targetConfigValue(target), oldValue) {
			a.targets[index] = newTarget
			break
		}
	}
	delete(a.latest, oldTarget.Host)
	delete(a.history, oldTarget.Host)
	a.addEventLocked(fmt.Sprintf("Custom target updated: %s → %s", oldValue, newValue))
	cfg := a.config
	a.mu.Unlock()
	_ = saveConfig(cfg)
	return true
}

func (t *TargetManagerWindow) refresh() {
	combo := t.controls[tmCombo]
	procSendMessageW.Call(uintptr(combo), CB_RESETCONTENT, 0, 0)
	t.parent.mu.RLock()
	targets := append([]string(nil), t.parent.config.CustomTargets...)
	t.parent.mu.RUnlock()
	for _, target := range targets {
		comboAdd(combo, target)
	}
	if len(targets) > 0 {
		comboSet(combo, 0)
		setText(t.controls[tmEdit], targets[0])
	} else {
		setText(t.controls[tmEdit], "")
	}
	enable(t.controls[tmReplace], len(targets) > 0)
	enable(t.controls[tmRemove], len(targets) > 0)
}

func (t *TargetManagerWindow) buildControls() {
	t.controls[tmInfo] = createControl(t.hwnd, "STATIC", "Manage ping, TCP and HTTP/HTTPS targets. Examples: 1.1.1.1, tcp://example.com:443, https://example.com/health", WS_CHILD|WS_VISIBLE|SS_LEFT, tmInfo)
	t.controls[tmCombo] = createControl(t.hwnd, "COMBOBOX", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|WS_VSCROLL|CBS_DROPDOWNLIST, tmCombo)
	t.controls[tmEdit] = createControl(t.hwnd, "EDIT", "", WS_CHILD|WS_VISIBLE|WS_BORDER|WS_TABSTOP|ES_LEFT|ES_AUTOHSCROLL, tmEdit)
	t.controls[tmAdd] = createControl(t.hwnd, "BUTTON", "Add", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, tmAdd)
	t.controls[tmReplace] = createControl(t.hwnd, "BUTTON", "Save changes", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, tmReplace)
	t.controls[tmRemove] = createControl(t.hwnd, "BUTTON", "Remove", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, tmRemove)
	t.controls[tmClose] = createControl(t.hwnd, "BUTTON", "Close", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_DEFPUSHBUTTON, tmClose)
	t.refresh()
}

func (t *TargetManagerWindow) layout(width, height int32) {
	move(t.controls[tmInfo], 24, 20, width-48, 42)
	move(t.controls[tmCombo], 24, 72, width-48, 180)
	move(t.controls[tmEdit], 24, 112, width-48, 28)
	move(t.controls[tmAdd], 24, 160, 90, 30)
	move(t.controls[tmReplace], 122, 160, 120, 30)
	move(t.controls[tmRemove], 250, 160, 90, 30)
	move(t.controls[tmClose], width-114, height-52, 90, 30)
}

func targetManagerProc(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	manager := globalTargetManager
	if manager == nil {
		result, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
		return result
	}
	switch msg {
	case WM_CREATE:
		manager.hwnd = hwnd
		manager.buildControls()
		dark := isDarkTheme(manager.parent.config.Theme)
		applyWindowDarkMode(hwnd, dark)
		for _, control := range manager.controls {
			applyControlTheme(control, dark)
		}
		return 0
	case WM_SIZE:
		manager.layout(int32(loword(lParam)), int32(hiword(lParam)))
		return 0
	case WM_ERASEBKGND:
		fillClientBackground(hwnd, syscall.Handle(wParam), isDarkTheme(manager.parent.config.Theme))
		return 1
	case WM_CTLCOLORSTATIC, WM_CTLCOLORBTN:
		return themeControlColor(syscall.Handle(wParam), isDarkTheme(manager.parent.config.Theme), false)
	case WM_CTLCOLOREDIT:
		return themeControlColor(syscall.Handle(wParam), isDarkTheme(manager.parent.config.Theme), true)
	case WM_COMMAND:
		id, notify := int(loword(wParam)), hiword(wParam)
		if id == tmCombo && notify == CBN_SELCHANGE {
			setText(manager.controls[tmEdit], getText(manager.controls[tmCombo]))
			return 0
		}
		if notify == BN_CLICKED {
			switch id {
			case tmAdd:
				if manager.parent.addCustomTarget(getText(manager.controls[tmEdit])) {
					manager.refresh()
					manager.parent.refreshCustomTargetCombo("")
					manager.parent.refreshUI()
				}
			case tmReplace:
				oldValue := getText(manager.controls[tmCombo])
				if manager.parent.replaceCustomTarget(oldValue, getText(manager.controls[tmEdit])) {
					manager.refresh()
					manager.parent.refreshCustomTargetCombo("")
					manager.parent.refreshUI()
				}
			case tmRemove:
				if manager.parent.removeCustomTarget(getText(manager.controls[tmCombo])) {
					manager.refresh()
					manager.parent.refreshCustomTargetCombo("")
					manager.parent.refreshUI()
				}
			case tmClose:
				procDestroyWindow.Call(uintptr(hwnd))
			}
		}
		return 0
	case WM_CLOSE:
		procDestroyWindow.Call(uintptr(hwnd))
		return 0
	case WM_DESTROY:
		enable(manager.parent.hwnd, true)
		procSetForegroundWindow.Call(uintptr(manager.parent.hwnd))
		globalTargetManager = nil
		return 0
	}
	result, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return result
}

func openTargetManager(a *App) {
	if globalTargetManager != nil && globalTargetManager.hwnd != 0 {
		procSetForegroundWindow.Call(uintptr(globalTargetManager.hwnd))
		return
	}
	instance, _, _ := procGetModuleHandleW.Call(0)
	className := ptr("NetWatcherTargetManager")
	icon := loadIconFromFile(runtimeIconPath())
	cursor, _, _ := procLoadCursorW.Call(0, IDC_ARROW)
	windowClass := WNDCLASSEX{CbSize: uint32(unsafe.Sizeof(WNDCLASSEX{})), LpfnWndProc: targetManagerProcCallback, HInstance: syscall.Handle(instance), HIcon: icon, HCursor: syscall.Handle(cursor), HbrBackground: syscall.Handle(COLOR_WINDOW + 1), LpszClassName: className, HIconSm: icon}
	atom, _, registerErr := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&windowClass)))
	if atom == 0 {
		if errno, ok := registerErr.(syscall.Errno); !ok || errno != syscall.Errno(1410) {
			messageBox(appName, "Target Manager could not be opened: "+registerErr.Error(), MB_OK|MB_ICONERROR)
			return
		}
	}
	manager := &TargetManagerWindow{parent: a, controls: map[int]syscall.Handle{}}
	globalTargetManager = manager
	style := uint32(WS_CAPTION | WS_SYSMENU | WS_CLIPCHILDREN | WS_VISIBLE)
	hwnd, _, createErr := procCreateWindowExW.Call(WS_EX_DLGMODALFRAME|WS_EX_CONTROLPARENT, uintptr(unsafe.Pointer(className)), uintptr(unsafe.Pointer(ptr("NetWatcher Target Manager"))), uintptr(style), 0, 0, 640, 280, uintptr(a.hwnd), 0, instance, 0)
	if hwnd == 0 {
		globalTargetManager = nil
		messageBox(appName, "Target Manager window error: "+createErr.Error(), MB_OK|MB_ICONERROR)
		return
	}
	procSendMessageW.Call(hwnd, WM_SETICON, ICON_BIG, uintptr(icon))
	procSendMessageW.Call(hwnd, WM_SETICON, ICON_SMALL, uintptr(icon))
	var parentRect RECT
	if ok, _, _ := procGetWindowRect.Call(uintptr(a.hwnd), uintptr(unsafe.Pointer(&parentRect))); ok != 0 {
		x := parentRect.Left + (parentRect.Right-parentRect.Left-640)/2
		y := parentRect.Top + (parentRect.Bottom-parentRect.Top-280)/2
		procMoveWindow.Call(hwnd, uintptr(x), uintptr(y), 640, 280, 0)
	}
	enable(a.hwnd, false)
	procShowWindow.Call(hwnd, SW_SHOW)
	procUpdateWindow.Call(hwnd)
	procSetForegroundWindow.Call(hwnd)
}
