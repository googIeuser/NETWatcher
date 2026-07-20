package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestOutageHistoryAndEvidenceReport(t *testing.T) {
	dir := t.TempDir()
	outage := "start;end;duration_seconds;category;details\n" +
		time.Now().Add(-time.Hour).Format(time.RFC3339) + ";" + time.Now().Add(-59*time.Minute).Format(time.RFC3339) + ";60;ISP_OUTAGE;test outage\n"
	if err := os.WriteFile(filepath.Join(dir, "outages.csv"), []byte(outage), 0644); err != nil {
		t.Fatal(err)
	}
	history, err := generateOutageHistoryPage(dir)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(history)
	if !strings.Contains(string(data), "Outage History") {
		t.Fatal("history page missing title")
	}
	report, err := generateEvidenceReport(dir, 7)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(report); err != nil {
		t.Fatal(err)
	}
}

func TestCleanupOldLogs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "samples_2000-01-01.csv")
	if err := os.WriteFile(path, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-60 * 24 * time.Hour)
	_ = os.Chtimes(path, old, old)
	removed, err := cleanupOldLogs(dir, 30)
	if err != nil || removed != 1 {
		t.Fatalf("removed=%d err=%v", removed, err)
	}
}
