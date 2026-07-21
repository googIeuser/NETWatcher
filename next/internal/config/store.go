package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"netwatcher/next/internal/domain"
)

func Default() domain.Config {
	return domain.Config{
		Language:            "en",
		Theme:               "dark",
		Interval:            2,
		TimeoutMS:           1500,
		HighLatencyMS:       150,
		ConfirmCycles:       2,
		StartMinimizedTray:  true,
		CloseToTray:         true,
		AutoCheckUpdates:    true,
		OutageNotifications: true,
		FirstRunComplete:    true,
		GraphRangeMinutes:   5,
		LogRetentionDays:    30,
	}
}

func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "NetWatcher")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "settings.json"), nil
}

func Load() (domain.Config, error) {
	cfg := Default()
	path, err := Path()
	if err != nil {
		return cfg, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Normalize(cfg), nil
	}
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Default(), err
	}
	return Normalize(cfg), nil
}

func Save(cfg domain.Config) error {
	cfg = Normalize(cfg)
	path, err := Path()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func Normalize(cfg domain.Config) domain.Config {
	if strings.EqualFold(cfg.Theme, "light") {
		cfg.Theme = "light"
	} else {
		cfg.Theme = "dark"
	}
	if !strings.EqualFold(cfg.Language, "tr") {
		cfg.Language = "en"
	} else {
		cfg.Language = "tr"
	}
	if cfg.Interval < 0.5 || cfg.Interval > 3600 {
		cfg.Interval = 2
	}
	if cfg.TimeoutMS < 200 || cfg.TimeoutMS > 60000 {
		cfg.TimeoutMS = 1500
	}
	if cfg.HighLatencyMS < 1 || cfg.HighLatencyMS > 60000 {
		cfg.HighLatencyMS = 150
	}
	if cfg.ConfirmCycles < 1 || cfg.ConfirmCycles > 20 {
		cfg.ConfirmCycles = 2
	}
	if cfg.GraphRangeMinutes != 5 && cfg.GraphRangeMinutes != 30 && cfg.GraphRangeMinutes != 60 && cfg.GraphRangeMinutes != 1440 {
		cfg.GraphRangeMinutes = 5
	}
	if cfg.LogRetentionDays < 0 || cfg.LogRetentionDays > 3650 {
		cfg.LogRetentionDays = 30
	}
	if cfg.CustomTargets == nil {
		cfg.CustomTargets = []string{}
	}
	return cfg
}
