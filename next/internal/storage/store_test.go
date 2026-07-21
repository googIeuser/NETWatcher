package storage

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"netwatcher/next/internal/domain"
)

func TestMeasurementOutageAndExport(t *testing.T) {
	dir := t.TempDir()
	s, err := NewAt(dir)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Truncate(time.Second)
	if err := s.AppendMeasurement(domain.Measurement{Timestamp: now, TargetID: "a", TargetName: "A", Host: "1.1.1.1", Kind: "internet", Mode: "ping", Success: true, Latency: 12}); err != nil {
		t.Fatal(err)
	}
	items, err := s.ReadMeasurements(now.Add(-time.Minute))
	if err != nil || len(items) != 1 {
		t.Fatalf("measurements %d %v", len(items), err)
	}
	if err := s.AppendOutage(domain.Outage{Start: now.Format(time.RFC3339Nano), End: now.Add(time.Second).Format(time.RFC3339Nano), Category: "offline", DurationSeconds: 1}); err != nil {
		t.Fatal(err)
	}
	out, err := s.ReadOutages(now.Add(-time.Minute))
	if err != nil || len(out) != 1 {
		t.Fatalf("outages %d %v", len(out), err)
	}
	if _, err := os.Stat(filepath.Join(dir, "outages_v3.csv")); err != nil {
		t.Fatalf("v3 outage file was not created: %v", err)
	}
	path, err := s.ExportZIP("test.zip", map[string][]byte{"summary/test.txt": []byte("ok")})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	zf, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer zf.Close()
	found := false
	for _, item := range zf.File {
		if item.Name == "summary/test.txt" {
			found = true
		}
	}
	if !found {
		t.Fatal("diagnostics summary was not written to ZIP")
	}
}

func TestReadsLegacyVersion2LogsWithoutModifyingThem(t *testing.T) {
	dir := t.TempDir()
	now := time.Now().UTC().Truncate(time.Second)
	legacySamples := "\ufefftimestamp;name;host;target_type;success;latency_ms;message\r\n" +
		now.Format(time.RFC3339) + ";Cloudflare;1.1.1.1;internet;true;18,50;Reply received\r\n"
	legacyOutages := "\ufeffstart;end;duration_seconds;category;details\r\n" +
		now.Add(-time.Minute).Format(time.RFC3339) + ";" + now.Format(time.RFC3339) + ";60;ISP_OUTAGE;Legacy outage\r\n"
	if err := os.WriteFile(filepath.Join(dir, "samples_"+now.Format("2006-01-02")+".csv"), []byte(legacySamples), 0o644); err != nil {
		t.Fatal(err)
	}
	legacyOutagePath := filepath.Join(dir, "outages.csv")
	if err := os.WriteFile(legacyOutagePath, []byte(legacyOutages), 0o644); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(legacyOutagePath)

	s, err := NewAt(dir)
	if err != nil {
		t.Fatal(err)
	}
	measurements, err := s.ReadMeasurements(now.Add(-time.Hour))
	if err != nil || len(measurements) != 1 {
		t.Fatalf("legacy measurements: %d %v", len(measurements), err)
	}
	if measurements[0].Latency != 18.5 || measurements[0].Host != "1.1.1.1" || measurements[0].Mode != "ping" {
		t.Fatalf("unexpected legacy measurement: %#v", measurements[0])
	}
	outages, err := s.ReadOutages(now.Add(-time.Hour))
	if err != nil || len(outages) != 1 {
		t.Fatalf("legacy outages: %d %v", len(outages), err)
	}
	if outages[0].Category != "offline" || !strings.Contains(outages[0].Details, "Legacy") {
		t.Fatalf("unexpected legacy outage: %#v", outages[0])
	}
	if err := s.AppendOutage(domain.Outage{Start: now.Add(time.Minute).Format(time.RFC3339Nano), End: now.Add(2 * time.Minute).Format(time.RFC3339Nano), Category: "partial", DurationSeconds: 60}); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(legacyOutagePath)
	if string(before) != string(after) {
		t.Fatal("legacy outages.csv was modified")
	}
}

func TestReadsPreviewCommaOutageSchema(t *testing.T) {
	dir := t.TempDir()
	now := time.Now().UTC().Truncate(time.Second)
	content := "start,end,category,details,duration_seconds\n" +
		now.Add(-time.Minute).Format(time.RFC3339Nano) + "," + now.Format(time.RFC3339Nano) + ",degraded,Preview outage,60\n"
	if err := os.WriteFile(filepath.Join(dir, "outages.csv"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := NewAt(dir)
	if err != nil {
		t.Fatal(err)
	}
	outages, err := s.ReadOutages(now.Add(-time.Hour))
	if err != nil || len(outages) != 1 || outages[0].Category != "partial" {
		t.Fatalf("preview outages: %#v %v", outages, err)
	}
}
