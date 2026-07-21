//go:build windows

package main

import (
	"encoding/json"
	"os"
	"strings"
	"sync"
	"syscall"
)

// This file contains only a one-time upgrade safeguard. Access Mode itself was
// removed in NetWatcher 2.2.7. If an older version owned the Windows proxy, the
// original setting is restored before the current configuration is loaded.
type legacyAccessConfig struct {
	Owned            bool   `json:"access_proxy_owned"`
	PreviousEnabled  bool   `json:"access_previous_proxy_enabled"`
	PreviousServer   string `json:"access_previous_proxy_server"`
	PreviousOverride string `json:"access_previous_proxy_override"`
}

type legacyProxyState struct {
	Enabled  bool
	Server   string
	Override string
}

var (
	legacyProxyStateMu     sync.Mutex
	wininetDLL             = syscall.NewLazyDLL("wininet.dll")
	procInternetSetOptionW = wininetDLL.NewProc("InternetSetOptionW")
)

const (
	internetOptionSettingsChanged = 39
	internetOptionRefresh         = 37
)

func notifyLegacyProxyChanged() {
	procInternetSetOptionW.Call(0, internetOptionSettingsChanged, 0, 0)
	procInternetSetOptionW.Call(0, internetOptionRefresh, 0, 0)
}

func restoreLegacySystemProxy(state legacyProxyState) {
	legacyProxyStateMu.Lock()
	defer legacyProxyStateMu.Unlock()

	const key = `HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings`
	enabled := "0"
	if state.Enabled {
		enabled = "1"
	}
	_ = hiddenCommand("reg", "add", key, "/v", "ProxyEnable", "/t", "REG_DWORD", "/d", enabled, "/f").Run()
	if strings.TrimSpace(state.Server) != "" {
		_ = hiddenCommand("reg", "add", key, "/v", "ProxyServer", "/t", "REG_SZ", "/d", state.Server, "/f").Run()
	} else {
		_ = hiddenCommand("reg", "delete", key, "/v", "ProxyServer", "/f").Run()
	}
	if strings.TrimSpace(state.Override) != "" {
		_ = hiddenCommand("reg", "add", key, "/v", "ProxyOverride", "/t", "REG_SZ", "/d", state.Override, "/f").Run()
	} else {
		_ = hiddenCommand("reg", "delete", key, "/v", "ProxyOverride", "/f").Run()
	}
	notifyLegacyProxyChanged()
}

func recoverStaleSystemProxy() {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return
	}
	var legacy legacyAccessConfig
	if json.Unmarshal(data, &legacy) != nil || !legacy.Owned {
		return
	}

	restoreLegacySystemProxy(legacyProxyState{
		Enabled:  legacy.PreviousEnabled,
		Server:   legacy.PreviousServer,
		Override: legacy.PreviousOverride,
	})

	var values map[string]any
	if json.Unmarshal(data, &values) != nil {
		return
	}
	for _, key := range []string{
		"access_proxy_port",
		"access_fragment_size",
		"access_use_system_proxy",
		"access_auto_start",
		"access_proxy_owned",
		"access_previous_proxy_enabled",
		"access_previous_proxy_server",
		"access_previous_proxy_override",
	} {
		delete(values, key)
	}
	cleaned, err := json.MarshalIndent(values, "", "  ")
	if err == nil {
		_ = os.WriteFile(configPath(), cleaned, 0644)
	}
}
