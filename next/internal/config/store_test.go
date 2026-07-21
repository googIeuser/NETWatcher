package config

import "testing"

func TestNormalizePreservesSupportedLanguageAndRanges(t *testing.T) {
	cfg := Default()
	cfg.Language = "tr"
	cfg.Theme = "light"
	cfg.Interval = 0.1
	cfg.GraphRangeMinutes = 999
	got := Normalize(cfg)
	if got.Language != "tr" || got.Theme != "light" {
		t.Fatalf("language/theme not preserved: %+v", got)
	}
	if got.Interval != 2 || got.GraphRangeMinutes != 5 {
		t.Fatalf("invalid values not normalised: %+v", got)
	}
}
