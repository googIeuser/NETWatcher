//go:build windows

package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

type SystemProxyState struct {
	Enabled  bool
	Server   string
	Override string
}

var (
	accessStateMu          sync.Mutex
	wininetDLL             = syscall.NewLazyDLL("wininet.dll")
	procInternetSetOptionW = wininetDLL.NewProc("InternetSetOptionW")
)

const (
	internetOptionSettingsChanged = 39
	internetOptionRefresh         = 37
)

func queryRegistryValue(key, name string) (string, bool) {
	out, err := hiddenCommand("reg", "query", key, "/v", name).CombinedOutput()
	if err != nil {
		return "", false
	}
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) >= 3 && strings.EqualFold(fields[0], name) {
			return strings.Join(fields[2:], " "), true
		}
	}
	return "", false
}

func currentSystemProxyState() SystemProxyState {
	const key = `HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings`
	enabledText, _ := queryRegistryValue(key, "ProxyEnable")
	server, _ := queryRegistryValue(key, "ProxyServer")
	override, _ := queryRegistryValue(key, "ProxyOverride")
	enabledValue, _ := strconv.ParseInt(strings.TrimPrefix(strings.ToLower(strings.TrimSpace(enabledText)), "0x"), 16, 64)
	if enabledValue == 0 {
		decimal, _ := strconv.Atoi(strings.TrimSpace(enabledText))
		enabledValue = int64(decimal)
	}
	return SystemProxyState{Enabled: enabledValue != 0, Server: server, Override: override}
}

func notifyProxyChanged() {
	procInternetSetOptionW.Call(0, internetOptionSettingsChanged, 0, 0)
	procInternetSetOptionW.Call(0, internetOptionRefresh, 0, 0)
}

func applySystemProxyAddress(address string) error {
	accessStateMu.Lock()
	defer accessStateMu.Unlock()
	const key = `HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings`
	commands := [][]string{
		{"add", key, "/v", "ProxyEnable", "/t", "REG_DWORD", "/d", "1", "/f"},
		{"add", key, "/v", "ProxyServer", "/t", "REG_SZ", "/d", address, "/f"},
		{"add", key, "/v", "ProxyOverride", "/t", "REG_SZ", "/d", "<local>;localhost;127.*", "/f"},
	}
	for _, args := range commands {
		if output, err := hiddenCommand("reg", args...).CombinedOutput(); err != nil {
			return fmt.Errorf("system proxy update failed: %v (%s)", err, strings.TrimSpace(string(output)))
		}
	}
	notifyProxyChanged()
	return nil
}

func setSystemProxy(address string) (*SystemProxyState, error) {
	previous := currentSystemProxyState()
	if err := applySystemProxyAddress(address); err != nil {
		return nil, err
	}
	return &previous, nil
}

func restoreSystemProxy(state *SystemProxyState) error {
	if state == nil {
		return nil
	}
	accessStateMu.Lock()
	defer accessStateMu.Unlock()
	const key = `HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings`
	enabled := "0"
	if state.Enabled {
		enabled = "1"
	}
	if err := hiddenCommand("reg", "add", key, "/v", "ProxyEnable", "/t", "REG_DWORD", "/d", enabled, "/f").Run(); err != nil {
		return err
	}
	if state.Server != "" {
		_ = hiddenCommand("reg", "add", key, "/v", "ProxyServer", "/t", "REG_SZ", "/d", state.Server, "/f").Run()
	} else {
		_ = hiddenCommand("reg", "delete", key, "/v", "ProxyServer", "/f").Run()
	}
	if state.Override != "" {
		_ = hiddenCommand("reg", "add", key, "/v", "ProxyOverride", "/t", "REG_SZ", "/d", state.Override, "/f").Run()
	} else {
		_ = hiddenCommand("reg", "delete", key, "/v", "ProxyOverride", "/f").Run()
	}
	notifyProxyChanged()
	return nil
}

func proxyStateFromConfig(cfg Config) *SystemProxyState {
	if !cfg.AccessProxyOwned {
		return nil
	}
	return &SystemProxyState{Enabled: cfg.AccessPreviousEnabled, Server: cfg.AccessPreviousServer, Override: cfg.AccessPreviousOverride}
}

func clearStoredProxyState(cfg *Config) {
	cfg.AccessProxyOwned = false
	cfg.AccessPreviousEnabled = false
	cfg.AccessPreviousServer = ""
	cfg.AccessPreviousOverride = ""
}

func recoverStaleSystemProxy(cfg *Config) {
	if cfg == nil || !cfg.AccessProxyOwned || cfg.AccessAutoStart {
		return
	}
	_ = restoreSystemProxy(proxyStateFromConfig(*cfg))
	clearStoredProxyState(cfg)
	_ = saveConfig(*cfg)
}

func (a *App) startAccessMode(useSystemProxy bool) error {
	a.mu.Lock()
	if a.accessProxyEnabled {
		a.mu.Unlock()
		return nil
	}
	proxy := NewAccessProxy(a.config.AccessPort, a.config.AccessFragmentSize)
	storedPrevious := proxyStateFromConfig(a.config)
	a.mu.Unlock()

	if err := proxy.Start(); err != nil {
		return err
	}
	var previous *SystemProxyState
	if useSystemProxy {
		if storedPrevious != nil {
			previous = storedPrevious
			if err := applySystemProxyAddress(proxy.Address()); err != nil {
				proxy.Stop()
				return err
			}
		} else {
			captured, err := setSystemProxy(proxy.Address())
			if err != nil {
				proxy.Stop()
				return err
			}
			previous = captured
		}
	}

	a.mu.Lock()
	a.accessProxy = proxy
	a.accessProxyEnabled = true
	a.accessProxyPrevious = previous
	a.config.AccessPort = proxy.Port
	a.config.AccessUseSystemProxy = useSystemProxy
	if previous != nil {
		a.config.AccessProxyOwned = true
		a.config.AccessPreviousEnabled = previous.Enabled
		a.config.AccessPreviousServer = previous.Server
		a.config.AccessPreviousOverride = previous.Override
	} else {
		clearStoredProxyState(&a.config)
	}
	cfg := a.config
	a.addEventLocked("Access Mode started on " + proxy.Address() + ". HTTPS reachability uses DoH and a local fragmented proxy; no ICMP ping is used.")
	a.mu.Unlock()
	_ = saveConfig(cfg)
	postRefresh(a.hwnd)
	return nil
}

func (a *App) stopAccessMode() {
	a.mu.Lock()
	proxy := a.accessProxy
	previous := a.accessProxyPrevious
	if previous == nil {
		previous = proxyStateFromConfig(a.config)
	}
	wasEnabled := a.accessProxyEnabled
	a.accessProxy = nil
	a.accessProxyPrevious = nil
	a.accessProxyEnabled = false
	clearStoredProxyState(&a.config)
	cfg := a.config
	if wasEnabled {
		a.addEventLocked("Access Mode stopped.")
	}
	a.mu.Unlock()
	if proxy != nil {
		proxy.Stop()
	}
	_ = restoreSystemProxy(previous)
	_ = saveConfig(cfg)
	postRefresh(a.hwnd)
}

func (a *App) toggleAccessMode() {
	a.mu.RLock()
	enabled := a.accessProxyEnabled
	useSystem := a.config.AccessUseSystemProxy
	a.mu.RUnlock()
	if enabled {
		a.stopAccessMode()
		return
	}
	if err := a.startAccessMode(useSystem); err != nil {
		messageBox(appName, "Access Mode could not start:\n\n"+err.Error(), MB_OK|MB_ICONERROR)
	}
}

func openProxySettings() {
	_ = exec.Command("cmd", "/c", "start", "ms-settings:network-proxy").Start()
}
