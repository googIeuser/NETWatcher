//go:build windows

package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// quickInstallStrings contains only the messages shown by the one-click setup.
// The language is selected from the Windows UI language; there is no custom
// language selector and therefore no installer UI callback that can deadlock.
func quickInstallText(_ string, key string) string {
	en := map[string]string{
		"title":            "NetWatcher Setup",
		"confirm":          "NetWatcher will be installed on your computer.\n\n• Installs for your user account (no administrator access required)\n• Creates Start menu and desktop shortcuts\n• Launches NetWatcher after installation\n• Asks on first launch whether it should start with Windows and keep running in the notification area\n\nDo you want to start the installation?",
		"installing":       "NetWatcher is being installed. Please wait a few seconds.",
		"success":          "NetWatcher was installed successfully and is ready to use.\n\nMeasurements are stored under Documents\\NetWatcherLogs.",
		"failure":          "Setup could not be completed:\n\n%s",
		"shortcut_warning": "Setup completed, but some shortcuts could not be created:\n\n%s",
		"description":      "Internet connection and outage monitor",
	}
	return en[key]
}

func hasArg(name string) bool {
	for _, arg := range os.Args[1:] {
		if strings.EqualFold(arg, name) {
			return true
		}
	}
	return false
}

func appendSetupLog(line string) {
	dir := filepath.Join(os.Getenv("LOCALAPPDATA"), appName)
	if strings.TrimSpace(dir) == "" {
		return
	}
	_ = os.MkdirAll(dir, 0755)
	f, err := os.OpenFile(filepath.Join(dir, "setup.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = fmt.Fprintf(f, "%s  %s\r\n", time.Now().Format(time.RFC3339), line)
}

// performOneClickInstall performs a per-user install. Shortcut failures are
// returned separately because they must never make the installation unusable.
func performOneClickInstall(lang string) (dest string, warnings []string, err error) {
	installDir := defaultInstallDir()
	appendSetupLog("setup started; version=" + appVersion)

	if installDir == "" || !filepath.IsAbs(installDir) {
		return "", nil, fmt.Errorf("invalid installation directory")
	}
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return "", nil, err
	}

	self, err := os.Executable()
	if err != nil {
		return "", nil, err
	}
	dest = filepath.Join(installDir, appName+".exe")
	iconDest := ensureIconFile(installDir)

	selfAbs, _ := filepath.Abs(self)
	destAbs, _ := filepath.Abs(dest)
	if !strings.EqualFold(filepath.Clean(selfAbs), filepath.Clean(destAbs)) {
		if _, statErr := os.Stat(dest); statErr == nil {
			_, _ = runHiddenTimeout(6*time.Second, "taskkill", "/IM", appName+".exe", "/F")
			time.Sleep(300 * time.Millisecond)
		}

		tempDest := dest + ".installing"
		_ = os.Remove(tempDest)

		input, openErr := os.Open(self)
		if openErr != nil {
			return "", nil, openErr
		}
		output, createErr := os.Create(tempDest)
		if createErr != nil {
			_ = input.Close()
			return "", nil, createErr
		}
		_, copyErr := io.Copy(output, input)
		closeOutErr := output.Close()
		closeInErr := input.Close()
		if copyErr != nil {
			_ = os.Remove(tempDest)
			return "", nil, copyErr
		}
		if closeOutErr != nil {
			_ = os.Remove(tempDest)
			return "", nil, closeOutErr
		}
		if closeInErr != nil {
			_ = os.Remove(tempDest)
			return "", nil, closeInErr
		}
		_ = os.Remove(dest)
		if renameErr := os.Rename(tempDest, dest); renameErr != nil {
			_ = os.Remove(tempDest)
			return "", nil, renameErr
		}
	}

	_, configStatErr := os.Stat(configPath())
	isNewInstall := os.IsNotExist(configStatErr)
	cfg := loadConfig()
	cfg.Language = normalizeLanguage(lang)
	if isNewInstall {
		cfg.AutoStart = false
		cfg.StartMinimizedTray = true
		cfg.AutoMonitor = true
		cfg.AutoCheckUpdates = true
		cfg.OutageNotifications = true
		cfg.FirstRunComplete = false
	}
	if err := saveConfig(cfg); err != nil {
		return "", nil, err
	}

	// Keep the existing Windows-startup preference during upgrades and migrate
	// older Run entries from --app to the selected startup behavior.
	runKey := `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`
	if cfg.AutoStart {
		runValue := fmt.Sprintf("\"%s\" %s", dest, autoStartArgument(cfg.StartMinimizedTray))
		if _, regErr := runHiddenTimeout(8*time.Second, "reg", "add", runKey, "/v", appName, "/t", "REG_SZ", "/d", runValue, "/f"); regErr != nil {
			return "", warnings, regErr
		}
	} else {
		_, _ = runHiddenTimeout(8*time.Second, "reg", "delete", runKey, "/v", appName, "/f")
	}

	description := quickInstallText(lang, "description")
	appData := os.Getenv("APPDATA")
	startDir := filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", appName)
	if mkErr := os.MkdirAll(startDir, 0755); mkErr != nil {
		warnings = append(warnings, mkErr.Error())
	} else {
		if scErr := createShortcut(filepath.Join(startDir, appName+".lnk"), dest, "--app", description, iconDest); scErr != nil {
			warnings = append(warnings, scErr.Error())
		}
		if scErr := createShortcut(filepath.Join(startDir, "Uninstall "+appName+".lnk"), dest, "--uninstall", "Uninstall "+appName, iconDest); scErr != nil {
			warnings = append(warnings, scErr.Error())
		}
	}

	desktopOut, desktopErr := runHiddenTimeout(8*time.Second, "powershell", "-NoProfile", "-NonInteractive", "-Command", "[Environment]::GetFolderPath('Desktop')")
	desktop := strings.TrimSpace(string(desktopOut))
	if desktop == "" {
		desktop = filepath.Join(os.Getenv("USERPROFILE"), "Desktop")
	}
	if desktopErr != nil {
		warnings = append(warnings, desktopErr.Error())
	}
	if scErr := createShortcut(filepath.Join(desktop, appName+".lnk"), dest, "--app", description, iconDest); scErr != nil {
		warnings = append(warnings, scErr.Error())
	}

	uninstallKey := `HKCU\Software\Microsoft\Windows\CurrentVersion\Uninstall\NetWatcher`
	uninstallString := fmt.Sprintf("\"%s\" --uninstall", dest)
	installDate := time.Now().Format("20060102")
	sizeKB := "0"
	if info, statErr := os.Stat(dest); statErr == nil {
		sizeKB = fmt.Sprintf("%d", (info.Size()+1023)/1024)
	}
	regCommands := [][]string{
		{"add", uninstallKey, "/v", "DisplayName", "/t", "REG_SZ", "/d", appName, "/f"},
		{"add", uninstallKey, "/v", "DisplayVersion", "/t", "REG_SZ", "/d", appVersion, "/f"},
		{"add", uninstallKey, "/v", "Publisher", "/t", "REG_SZ", "/d", appName, "/f"},
		{"add", uninstallKey, "/v", "InstallLocation", "/t", "REG_SZ", "/d", installDir, "/f"},
		{"add", uninstallKey, "/v", "DisplayIcon", "/t", "REG_SZ", "/d", iconDest, "/f"},
		{"add", uninstallKey, "/v", "UninstallString", "/t", "REG_SZ", "/d", uninstallString, "/f"},
		{"add", uninstallKey, "/v", "QuietUninstallString", "/t", "REG_SZ", "/d", uninstallString, "/f"},
		{"add", uninstallKey, "/v", "InstallDate", "/t", "REG_SZ", "/d", installDate, "/f"},
		{"add", uninstallKey, "/v", "EstimatedSize", "/t", "REG_DWORD", "/d", sizeKB, "/f"},
		{"add", uninstallKey, "/v", "NoModify", "/t", "REG_DWORD", "/d", "1", "/f"},
		{"add", uninstallKey, "/v", "NoRepair", "/t", "REG_DWORD", "/d", "1", "/f"},
	}
	for _, args := range regCommands {
		if _, regErr := runHiddenTimeout(8*time.Second, "reg", args...); regErr != nil {
			return "", warnings, regErr
		}
	}

	appendSetupLog("setup completed; install_dir=" + installDir)
	return dest, warnings, nil
}

func runOneClickInstaller() {
	lang := defaultLanguage()
	title := quickInstallText(lang, "title")
	silent := hasArg("--silent") || hasArg("/silent") || hasArg("/s")

	if !silent {
		if messageBox(title, quickInstallText(lang, "confirm"), MB_YESNO|MB_ICONQUESTION) != IDYES {
			return
		}
	}

	dest, warnings, err := performOneClickInstall(lang)
	if err != nil {
		appendSetupLog("setup failed: " + err.Error())
		if !silent {
			messageBox(title, fmt.Sprintf(quickInstallText(lang, "failure"), err), MB_OK|MB_ICONERROR)
		}
		return
	}

	if !silent {
		if len(warnings) > 0 {
			messageBox(title, fmt.Sprintf(quickInstallText(lang, "shortcut_warning"), strings.Join(warnings, "\n")), MB_OK|MB_ICONWARNING)
		} else {
			messageBox(title, quickInstallText(lang, "success"), MB_OK|MB_ICONINFORMATION)
		}
	}

	// Do not block setup completion on application launch.
	cmd := exec.Command(dest, "--app")
	cmd.SysProcAttr = nil
	_ = cmd.Start()
}
