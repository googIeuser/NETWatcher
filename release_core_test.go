package main

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		candidate string
		current   string
		want      bool
	}{
		{"v2.1.0", "2.0.7", true},
		{"2.1.0", "2.1.0", false},
		{"2.0.9", "2.1.0", false},
		{"v3.0.0-beta.1", "2.9.9", true},
	}
	for _, tt := range tests {
		if got := isNewerVersion(tt.candidate, tt.current); got != tt.want {
			t.Fatalf("isNewerVersion(%q, %q) = %v, want %v", tt.candidate, tt.current, got, tt.want)
		}
	}
}

func writeSampleCSV(t *testing.T, dir string) {
	t.Helper()
	path := filepath.Join(dir, "samples_"+time.Now().Format("2006-01-02")+".csv")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	w := csv.NewWriter(f)
	w.Comma = ';'
	rows := [][]string{
		{"timestamp", "name", "host", "target_type", "success", "latency_ms", "message"},
		{time.Now().Add(-time.Minute).Format(time.RFC3339), "Cloudflare", "1.1.1.1", "internet", "true", "15.5", "Reply received"},
		{time.Now().Add(-30 * time.Second).Format(time.RFC3339), "Cloudflare", "1.1.1.1", "internet", "false", "", "Timed out"},
	}
	if err := w.WriteAll(rows); err != nil {
		t.Fatal(err)
	}
	w.Flush()
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestStatisticsAndZipExport(t *testing.T) {
	dir := t.TempDir()
	writeSampleCSV(t, dir)

	stats, err := readDiskStatistics(dir, time.Now().Add(-24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	item := stats["Cloudflare (1.1.1.1)"]
	if item == nil || item.Samples != 2 || item.Successes != 1 {
		t.Fatalf("unexpected stats: %#v", item)
	}

	page, err := generateStatisticsPage(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(page); err != nil {
		t.Fatal(err)
	}

	archive, err := exportLogsZip(dir)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(archive)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Fatal("export archive is empty")
	}
}
