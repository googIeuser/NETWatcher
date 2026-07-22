//go:build windows

package platform

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func hidden(name string, args ...string) *exec.Cmd {
	c := exec.Command(name, args...)
	c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	return c
}
func SetStartWithWindows(enabled bool) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	key := `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`
	if enabled {
		value := fmt.Sprintf(`"%s" --autostart`, exe)
		out, err := hidden("reg", "add", key, "/v", "NetWatcher", "/t", "REG_SZ", "/d", value, "/f").CombinedOutput()
		if err != nil {
			return fmt.Errorf("startup registration failed: %v (%s)", err, strings.TrimSpace(string(out)))
		}
		return nil
	}
	_ = hidden("reg", "delete", key, "/v", "NetWatcher", "/f").Run()
	return nil
}
func OpenPath(path string) error { return exec.Command("explorer", path).Start() }
func OpenFile(path string) error {
	return exec.Command("rundll32", "url.dll,FileProtocolHandler", path).Start()
}
func OpenURL(url string) error {
	return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
}
func Notify(title, message string) error {
	esc := func(v string) string {
		return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\"", "&quot;", "'", "&apos;").Replace(v)
	}
	script := `$ErrorActionPreference='SilentlyContinue';[Windows.UI.Notifications.ToastNotificationManager,Windows.UI.Notifications,ContentType=WindowsRuntime]|Out-Null;$xml=New-Object Windows.Data.Xml.Dom.XmlDocument;$xml.LoadXml('<toast><visual><binding template="ToastGeneric"><text>` + esc(title) + `</text><text>` + esc(message) + `</text></binding></visual></toast>');$toast=New-Object Windows.UI.Notifications.ToastNotification $xml;[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('NetWatcher').Show($toast)`
	return hidden("powershell", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script).Run()
}
func ProcessID() string { return strconv.Itoa(os.Getpid()) }
