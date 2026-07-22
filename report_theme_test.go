package main

import (
	"strings"
	"testing"
)

func TestReportThemeMatchesStatisticsDesign(t *testing.T) {
	required := []string{
		"linear-gradient(135deg,#0b6cff,#38bdf8)",
		".summary-grid",
		".card",
		"@media(prefers-color-scheme:light)",
		"@media print",
	}
	for _, token := range required {
		if !strings.Contains(reportCSS, token) {
			t.Fatalf("report theme is missing %q", token)
		}
	}
}
