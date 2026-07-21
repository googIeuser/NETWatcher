package monitor

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"netwatcher/next/internal/config"
	"netwatcher/next/internal/storage"
)

func TestRestoresLegacyCloudflareHistoryIntoCurrentTarget(t *testing.T) {
	dir := t.TempDir()
	now := time.Now().UTC().Truncate(time.Second)
	legacy := "\ufefftimestamp;name;host;target_type;success;latency_ms;message\r\n" +
		now.Format(time.RFC3339) + ";Cloudflare;1.1.1.1;internet;true;21.25;Reply received\r\n"
	if err := os.WriteFile(filepath.Join(dir, "samples_"+now.Format("2006-01-02")+".csv"), []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}
	store, err := storage.NewAt(dir)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.GraphRangeMinutes = 1440
	engine := NewEngine(cfg, store, "test")
	snapshot := engine.Snapshot()
	for _, target := range snapshot.Targets {
		if target.Target.Name == "Cloudflare" {
			if len(target.History) != 1 || target.Latency != 21.25 || target.State != "online" {
				t.Fatalf("legacy history was not restored: %#v", target)
			}
			return
		}
	}
	t.Fatal("Cloudflare target not found")
}
