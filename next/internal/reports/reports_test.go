package reports

import (
	"os"
	"strings"
	"testing"
	"time"

	"netwatcher/next/internal/domain"
	"netwatcher/next/internal/storage"
)

func TestGenerateEvidenceReportEscapesDataAndLocalizesCategory(t *testing.T) {
	store, err := storage.NewAt(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	stats := domain.Statistics{
		RangeHours: 24, Samples: 2, Successful: 1, PacketLoss: 50, Uptime: 50,
		TargetBreakdown: []domain.TargetStatistics{{TargetID: "x", TargetName: "<Cloudflare>", Host: "1.1.1.1", Mode: "ping", Samples: 2, Successful: 1, PacketLoss: 50, Uptime: 50}},
	}
	now := time.Now().UTC()
	outages := []domain.Outage{{Start: now.Add(-time.Minute).Format(time.RFC3339Nano), End: now.Format(time.RFC3339Nano), Category: "offline", Details: "ISS <test>", DurationSeconds: 60}}
	result, err := Generate(store, "evidence", stats, outages, domain.Snapshot{QualityScore: 42, Version: "test"}, "tr")
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(result.Path)
	if err != nil {
		t.Fatal(err)
	}
	html := string(content)
	for _, expected := range []string{"NetWatcher ISS Kanıt Raporu", "ISS / internet kesintisi", "&lt;Cloudflare&gt;", "ISS &lt;test&gt;"} {
		if !strings.Contains(html, expected) {
			t.Fatalf("report missing %q", expected)
		}
	}
	if strings.Contains(html, "<Cloudflare>") || strings.Contains(html, "ISS <test>") {
		t.Fatal("unescaped report content")
	}
}
