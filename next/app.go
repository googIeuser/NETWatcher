package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	configstore "netwatcher/next/internal/config"
	"netwatcher/next/internal/domain"
	"netwatcher/next/internal/monitor"
	"netwatcher/next/internal/platform"
	"netwatcher/next/internal/reports"
	"netwatcher/next/internal/statistics"
	"netwatcher/next/internal/storage"
	updatecheck "netwatcher/next/internal/update"
)

const appVersion = "3.0.0"

type App struct {
	ctx                context.Context
	mu                 sync.RWMutex
	config             domain.Config
	engine             *monitor.Engine
	store              *storage.Store
	forceExit          bool
	startedFromWindows bool
}

func NewApp() *App {
	cfg, err := configstore.Load()
	if err != nil {
		cfg = configstore.Default()
	}
	store, _ := storage.New()
	app := &App{config: cfg, store: store, startedFromWindows: hasArg("--autostart")}
	app.engine = monitor.NewEngine(cfg, store, appVersion)
	app.engine.SetEventHandler(app.handleEngineEvent)
	return app
}
func hasArg(want string) bool {
	for _, arg := range os.Args[1:] {
		if strings.EqualFold(arg, want) {
			return true
		}
	}
	return false
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.mu.RLock()
	auto := a.config.AutoMonitor
	check := a.config.AutoCheckUpdates
	a.mu.RUnlock()
	if auto {
		a.engine.Start()
	}
	if check {
		go func() {
			time.Sleep(3 * time.Second)
			ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
			defer cancel()
			info, err := updatecheck.Check(ctx, appVersion)
			if err == nil && info.Available {
				a.mu.RLock()
				language := a.config.Language
				a.mu.RUnlock()
				title, message := "NetWatcher update available", "A newer NetWatcher release is available."
				if language == "tr" {
					title, message = "NetWatcher güncellemesi mevcut", "Daha yeni bir NetWatcher sürümü yayınlandı."
				}
				if !showTrayNotification(title, message, "info") {
					go func() { _ = platform.Notify(title, message) }()
				}
				if a.ctx != nil {
					wruntime.EventsEmit(a.ctx, "update:available", info)
				}
			}
		}()
	}
}
func (a *App) domReady(ctx context.Context) {
	a.ctx = ctx
	startTray(a)
	a.mu.RLock()
	hide := a.startedFromWindows && a.config.StartMinimizedTray
	a.mu.RUnlock()
	if hide {
		wruntime.WindowHide(ctx)
	}
}
func (a *App) shutdown(context.Context) { a.engine.Stop(); stopTray() }
func (a *App) beforeClose(ctx context.Context) bool {
	a.mu.RLock()
	hide := a.config.CloseToTray && !a.forceExit
	a.mu.RUnlock()
	if hide {
		wruntime.WindowHide(ctx)
		return true
	}
	return false
}

func (a *App) handleEngineEvent(event domain.Event) {
	a.mu.RLock()
	notify := a.config.OutageNotifications
	a.mu.RUnlock()
	if notify && (event.Category == "outage" || event.Category == "recovery") {
		title := "NetWatcher"
		if event.Category == "outage" {
			title = "NetWatcher connectivity alert"
		} else {
			title = "NetWatcher recovery"
		}
		if !showTrayNotification(title, event.Message, event.Level) {
			go func() { _ = platform.Notify(title, event.Message) }()
		}
	}
	if a.ctx != nil {
		wruntime.EventsEmit(a.ctx, "monitor:event", event)
		wruntime.EventsEmit(a.ctx, "monitor:snapshot", a.engine.Snapshot())
	}
}

func (a *App) GetSnapshot() domain.Snapshot { return a.engine.Snapshot() }
func (a *App) GetSettings() domain.Config   { a.mu.RLock(); defer a.mu.RUnlock(); return a.config }
func (a *App) StartMonitoring() domain.Snapshot {
	a.engine.Start()
	syncTrayMenu()
	return a.engine.Snapshot()
}
func (a *App) StopMonitoring() domain.Snapshot {
	a.engine.Stop()
	syncTrayMenu()
	return a.engine.Snapshot()
}

func (a *App) SaveSettings(cfg domain.Config) (domain.Config, error) {
	cfg = configstore.Normalize(cfg)
	a.mu.RLock()
	previous := a.config
	a.mu.RUnlock()
	// Refresh the Run entry even when the checkbox value did not change. This
	// migrates an existing 2.x startup entry to the current executable path.
	if err := platform.SetStartWithWindows(cfg.AutoStart); err != nil {
		return previous, err
	}
	if err := configstore.Save(cfg); err != nil {
		_ = platform.SetStartWithWindows(previous.AutoStart)
		return previous, err
	}
	a.mu.Lock()
	a.config = cfg
	a.mu.Unlock()
	a.engine.UpdateConfig(cfg)
	if a.ctx != nil {
		wruntime.EventsEmit(a.ctx, "settings:changed", cfg)
	}
	return cfg, nil
}

func cloneTargets(values []string) []string {
	return append([]string(nil), values...)
}

func (a *App) AddTarget(raw string) (domain.Config, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return a.GetSettings(), errors.New("target cannot be empty")
	}
	a.mu.RLock()
	current := a.config
	current.CustomTargets = cloneTargets(a.config.CustomTargets)
	a.mu.RUnlock()
	for _, value := range current.CustomTargets {
		if strings.EqualFold(strings.TrimSpace(value), raw) {
			return current, errors.New("target already exists")
		}
	}
	current.CustomTargets = append(current.CustomTargets, raw)
	if err := configstore.Save(current); err != nil {
		return a.GetSettings(), err
	}
	a.mu.Lock()
	a.config = current
	a.mu.Unlock()
	a.engine.UpdateConfig(current)
	return current, nil
}

func (a *App) EditTarget(oldValue, newValue string) (domain.Config, error) {
	oldValue = strings.TrimSpace(oldValue)
	newValue = strings.TrimSpace(newValue)
	if oldValue == "" || newValue == "" {
		return a.GetSettings(), errors.New("target values cannot be empty")
	}
	a.mu.RLock()
	current := a.config
	current.CustomTargets = cloneTargets(a.config.CustomTargets)
	a.mu.RUnlock()
	found := -1
	for i, value := range current.CustomTargets {
		trimmed := strings.TrimSpace(value)
		if strings.EqualFold(trimmed, oldValue) {
			found = i
			continue
		}
		if strings.EqualFold(trimmed, newValue) {
			return current, errors.New("target already exists")
		}
	}
	if found < 0 {
		return current, errors.New("target was not found")
	}
	current.CustomTargets[found] = newValue
	if err := configstore.Save(current); err != nil {
		return a.GetSettings(), err
	}
	a.mu.Lock()
	a.config = current
	a.mu.Unlock()
	a.engine.UpdateConfig(current)
	return current, nil
}

func (a *App) RemoveTarget(raw string) (domain.Config, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return a.GetSettings(), errors.New("target cannot be empty")
	}
	a.mu.RLock()
	current := a.config
	current.CustomTargets = cloneTargets(a.config.CustomTargets)
	a.mu.RUnlock()
	filtered := make([]string, 0, len(current.CustomTargets))
	removed := false
	for _, value := range current.CustomTargets {
		if strings.EqualFold(strings.TrimSpace(value), raw) {
			removed = true
			continue
		}
		filtered = append(filtered, value)
	}
	if !removed {
		return current, errors.New("target was not found")
	}
	current.CustomTargets = filtered
	if err := configstore.Save(current); err != nil {
		return a.GetSettings(), err
	}
	a.mu.Lock()
	a.config = current
	a.mu.Unlock()
	a.engine.UpdateConfig(current)
	return current, nil
}

func (a *App) GetStatistics(hours int) (domain.Statistics, error) {
	if a.store == nil {
		return domain.Statistics{}, errors.New("log storage is unavailable")
	}
	if hours <= 0 {
		hours = 24
	}
	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	measurements, err := a.store.ReadMeasurements(since)
	if err != nil {
		return domain.Statistics{}, err
	}
	days := hours/24 + 1
	outages := a.engine.Outages(days)
	return statistics.Build(hours, measurements, outages), nil
}
func (a *App) GetOutages(days int) []domain.Outage { return a.engine.Outages(days) }
func (a *App) GenerateHTMLReport(hours int) (domain.ReportResult, error) {
	return a.generateReport("standard", hours)
}
func (a *App) GenerateEvidenceReport(days int) (domain.ReportResult, error) {
	if days <= 0 {
		days = 7
	}
	return a.generateReport("evidence", days*24)
}
func (a *App) generateReport(kind string, hours int) (domain.ReportResult, error) {
	if a.store == nil {
		return domain.ReportResult{}, errors.New("log storage is unavailable")
	}
	stats, err := a.GetStatistics(hours)
	if err != nil {
		return domain.ReportResult{}, err
	}
	outages := a.engine.Outages(hours/24 + 1)
	result, err := reports.Generate(a.store, kind, stats, outages, a.engine.Snapshot(), a.GetSettings().Language)
	if err != nil {
		return result, err
	}
	_ = platform.OpenFile(result.Path)
	return result, nil
}

func (a *App) ExportDiagnostics(hours int) (domain.ReportResult, error) {
	if a.store == nil {
		return domain.ReportResult{}, errors.New("log storage is unavailable")
	}
	if hours <= 0 {
		hours = 168
	}
	stats, err := a.GetStatistics(hours)
	if err != nil {
		return domain.ReportResult{}, err
	}
	outages := a.engine.Outages(hours/24 + 1)
	snapshot := a.engine.Snapshot()
	cfg := a.GetSettings()
	statsJSON, _ := json.MarshalIndent(stats, "", "  ")
	outagesJSON, _ := json.MarshalIndent(outages, "", "  ")
	snapshotJSON, _ := json.MarshalIndent(snapshot, "", "  ")
	cfgJSON, _ := json.MarshalIndent(cfg, "", "  ")
	readme := []byte("NetWatcher diagnostics export\r\nGenerated: " + time.Now().Format(time.RFC3339) + "\r\nAll measurements were collected locally.\r\n")
	path, err := a.store.ExportZIP(storage.SafeFileName("NetWatcher_Diagnostics")+".zip", map[string][]byte{"summary/statistics.json": statsJSON, "summary/outages.json": outagesJSON, "summary/snapshot.json": snapshotJSON, "summary/settings.json": cfgJSON, "README.txt": readme})
	if err != nil {
		return domain.ReportResult{}, err
	}
	_ = platform.OpenPath(a.store.ReportsDir())
	return domain.ReportResult{Kind: "diagnostics", Path: path, CreatedAt: time.Now().Format(time.RFC3339Nano)}, nil
}
func (a *App) OpenLogsFolder() error {
	if a.store == nil {
		return errors.New("log storage is unavailable")
	}
	return platform.OpenPath(a.store.Dir())
}
func (a *App) OpenFile(path string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("path is empty")
	}
	return platform.OpenFile(path)
}
func (a *App) CheckForUpdates() (domain.UpdateInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	return updatecheck.Check(ctx, appVersion)
}
func (a *App) OpenReleaseURL(url string) error { return platform.OpenURL(url) }
func (a *App) ShowWindow() {
	if a.ctx != nil {
		wruntime.WindowShow(a.ctx)
		wruntime.WindowUnminimise(a.ctx)
	}
}
func (a *App) HideWindow() {
	if a.ctx != nil {
		wruntime.WindowHide(a.ctx)
	}
}
func (a *App) Quit() {
	a.mu.Lock()
	a.forceExit = true
	a.mu.Unlock()
	if a.ctx != nil {
		wruntime.Quit(a.ctx)
	}
}
func (a *App) Version() string { return appVersion }
func (a *App) SortOutages(items []domain.Outage) []domain.Outage {
	sort.Slice(items, func(i, j int) bool { return items[i].Start > items[j].Start })
	return items
}
func (a *App) DebugInfo() string {
	return fmt.Sprintf("NetWatcher %s · process %s", appVersion, platform.ProcessID())
}
