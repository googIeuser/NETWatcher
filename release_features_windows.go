//go:build windows

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"
	"unsafe"
)

// githubRepository is injected by the GitHub Actions release workflow using:
// -ldflags "-X main.githubRepository=${{ github.repository }}".
// Local builds use the public NetWatcher repository; release builds can still override it.
var githubRepository = "googIeuser/NetWatcher"

type TrayNotification struct {
	Title string
	Body  string
	Flags uint32
	URL   string
}

const (
	NIIF_NONE    = 0x00000000
	NIIF_INFO    = 0x00000001
	NIIF_WARNING = 0x00000002
	NIIF_ERROR   = 0x00000003
	NIF_INFO     = 0x00000010

	NIN_BALLOONUSERCLICK = WM_USER + 5
)

func (a *App) queueNotificationLocked(title, body string, flags uint32, url string) {
	if !a.config.OutageNotifications && url == "" {
		return
	}
	a.notificationQueue = append(a.notificationQueue, TrayNotification{Title: title, Body: body, Flags: flags, URL: url})
}

func (a *App) drainNotifications() {
	a.mu.Lock()
	items := append([]TrayNotification(nil), a.notificationQueue...)
	a.notificationQueue = nil
	a.mu.Unlock()
	for _, item := range items {
		a.showTrayNotification(item)
	}
}

func (a *App) showTrayNotification(item TrayNotification) {
	if a == nil || a.hwnd == 0 {
		return
	}
	a.ensureTrayIcon()
	a.trayMu.Lock()
	defer a.trayMu.Unlock()
	if !a.trayAdded {
		return
	}
	nid := NOTIFYICONDATA{
		CbSize:      uint32(unsafe.Sizeof(NOTIFYICONDATA{})),
		HWnd:        a.hwnd,
		UID:         trayIconID,
		UFlags:      NIF_INFO,
		DwInfoFlags: item.Flags,
	}
	copyUTF16(nid.SzInfoTitle[:], item.Title)
	copyUTF16(nid.SzInfo[:], item.Body)
	if item.URL != "" {
		a.mu.Lock()
		a.pendingUpdateURL = item.URL
		a.mu.Unlock()
	}
	procShellNotifyIconW.Call(NIM_MODIFY, uintptr(unsafe.Pointer(&nid)))
}

type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Name    string `json:"name"`
}

func latestGitHubRelease(repo string) (githubRelease, error) {
	if strings.TrimSpace(repo) == "" {
		return githubRelease{}, errors.New("update service is not configured for this build")
	}
	client := &http.Client{Timeout: 8 * time.Second}
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/repos/"+repo+"/releases/latest", nil)
	if err != nil {
		return githubRelease{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "NetWatcher/"+appVersion)
	response, err := client.Do(req)
	if err != nil {
		return githubRelease{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return githubRelease{}, fmt.Errorf("GitHub returned %s", response.Status)
	}
	var release githubRelease
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&release); err != nil {
		return githubRelease{}, err
	}
	if strings.TrimSpace(release.TagName) == "" || strings.TrimSpace(release.HTMLURL) == "" {
		return githubRelease{}, errors.New("release response is incomplete")
	}
	return release, nil
}

func (a *App) checkForUpdates(manual bool) {
	go func() {
		release, err := latestGitHubRelease(githubRepository)
		if err != nil {
			if manual {
				a.addEvent("Update check failed: " + err.Error())
				a.showTrayNotification(TrayNotification{Title: "NetWatcher update", Body: "The update check could not be completed.", Flags: NIIF_WARNING})
				postRefresh(a.hwnd)
			}
			return
		}
		if !isNewerVersion(release.TagName, appVersion) {
			if manual {
				a.showTrayNotification(TrayNotification{Title: "NetWatcher update", Body: "You are using the latest version.", Flags: NIIF_INFO})
			}
			return
		}
		name := release.TagName
		if strings.TrimSpace(release.Name) != "" {
			name = release.Name
		}
		a.addEvent("A new NetWatcher version is available: " + name)
		a.showTrayNotification(TrayNotification{
			Title: "NetWatcher update available",
			Body:  name + " is ready. Click this notification to open the download page.",
			Flags: NIIF_INFO,
			URL:   release.HTMLURL,
		})
		postRefresh(a.hwnd)
	}()
}

func revealFile(path string) {
	_ = exec.Command("explorer", "/select,"+path).Start()
}
