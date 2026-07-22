package main

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"html"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func normalizeVersion(v string) []int {
	v = strings.TrimSpace(strings.TrimPrefix(strings.ToLower(v), "v"))
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		v = v[:i]
	}
	parts := strings.Split(v, ".")
	out := make([]int, 3)
	for i := 0; i < len(out) && i < len(parts); i++ {
		n, _ := strconv.Atoi(parts[i])
		out[i] = n
	}
	return out
}

func isNewerVersion(candidate, current string) bool {
	a, b := normalizeVersion(candidate), normalizeVersion(current)
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return a[i] > b[i]
		}
	}
	return false
}

type periodStats struct {
	Name        string
	Samples     int
	Successes   int
	LatencySum  float64
	Latencies   []float64
	LastLatency float64
	HasLast     bool
	JitterSum   float64
	JitterPairs int
}

func (s *periodStats) add(success bool, latency float64) {
	s.Samples++
	if success {
		s.Successes++
		s.LatencySum += latency
		s.Latencies = append(s.Latencies, latency)
		if s.HasLast {
			delta := latency - s.LastLatency
			if delta < 0 {
				delta = -delta
			}
			s.JitterSum += delta
			s.JitterPairs++
		}
		s.LastLatency = latency
		s.HasLast = true
	}
}

func (s periodStats) uptime() float64 {
	if s.Samples == 0 {
		return 0
	}
	return float64(s.Successes) / float64(s.Samples) * 100
}
func (s periodStats) loss() float64 { return 100 - s.uptime() }
func (s periodStats) average() float64 {
	if s.Successes == 0 {
		return 0
	}
	return s.LatencySum / float64(s.Successes)
}
func (s periodStats) p95() float64 {
	if len(s.Latencies) == 0 {
		return 0
	}
	values := append([]float64(nil), s.Latencies...)
	sort.Float64s(values)
	return values[int(float64(len(values)-1)*0.95)]
}

func (s periodStats) jitter() float64 {
	if s.JitterPairs == 0 {
		return 0
	}
	return s.JitterSum / float64(s.JitterPairs)
}

func readDiskStatistics(logDir string, since time.Time) (map[string]*periodStats, error) {
	paths, err := filepath.Glob(filepath.Join(logDir, "samples_*.csv"))
	if err != nil {
		return nil, err
	}
	stats := map[string]*periodStats{}
	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		reader := csv.NewReader(f)
		reader.Comma = ';'
		reader.FieldsPerRecord = -1
		rows, err := reader.ReadAll()
		_ = f.Close()
		if err != nil {
			continue
		}
		for idx, row := range rows {
			if idx == 0 || len(row) < 7 {
				continue
			}
			row[0] = strings.TrimPrefix(row[0], "\ufeff")
			timestamp, err := time.Parse(time.RFC3339, row[0])
			if err != nil || timestamp.Before(since) {
				continue
			}
			key := row[1] + " (" + row[2] + ")"
			item := stats[key]
			if item == nil {
				item = &periodStats{Name: key}
				stats[key] = item
			}
			success, _ := strconv.ParseBool(row[4])
			latency, _ := strconv.ParseFloat(strings.ReplaceAll(row[5], ",", "."), 64)
			item.add(success, latency)
		}
	}
	return stats, nil
}

func statsTable(stats map[string]*periodStats) string {
	keys := make([]string, 0, len(stats))
	for key := range stats {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		return `<tr><td colspan="9">No samples are available for this period.</td></tr>`
	}
	var rows strings.Builder
	for _, key := range keys {
		item := stats[key]
		score, label := connectionQuality(RollingMetrics{Samples: item.Samples, Successes: item.Successes, Failures: item.Samples - item.Successes, PacketLoss: item.loss(), Average: item.average(), P95: item.p95(), Jitter: item.jitter()})
		fmt.Fprintf(&rows, `<tr><td>%s</td><td>%d</td><td>%.3f%%</td><td>%.3f%%</td><td>%.2f ms</td><td>%.2f ms</td><td>%.2f ms</td><td>%s (%d)</td><td>%d</td></tr>`,
			html.EscapeString(item.Name), item.Samples, item.uptime(), item.loss(), item.average(), item.p95(), item.jitter(), label, score, item.Samples-item.Successes)
	}
	return rows.String()
}

func generateStatisticsPage(logDir string) (string, error) {
	now := time.Now()
	stats24, err := readDiskStatistics(logDir, now.Add(-24*time.Hour))
	if err != nil {
		return "", err
	}
	stats7, err := readDiskStatistics(logDir, now.Add(-7*24*time.Hour))
	if err != nil {
		return "", err
	}
	page := fmt.Sprintf(`<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>NetWatcher Statistics</title><style>:root{color-scheme:light dark}body{font:14px/1.45 "Segoe UI",Arial,sans-serif;margin:0;background:#111820;color:#eaf0f7}.wrap{max-width:1100px;margin:auto;padding:28px}.hero{background:linear-gradient(135deg,#0b6cff,#38bdf8);padding:26px;border-radius:18px;color:white;box-shadow:0 18px 50px #0005}.card{background:#18222d;border:1px solid #2b3948;border-radius:16px;padding:18px;margin-top:18px;overflow:auto}table{width:100%%;border-collapse:collapse;min-width:720px}th,td{padding:10px 12px;text-align:left;border-bottom:1px solid #2b3948}th{color:#8ec5ff}small{opacity:.8}@media(prefers-color-scheme:light){body{background:#f3f6fa;color:#18222d}.card{background:white;border-color:#d9e2ec}th,td{border-color:#e3e9ef}}</style></head><body><div class="wrap"><section class="hero"><h1>NetWatcher Statistics</h1><p>Generated %s. Statistics are calculated locally from your CSV logs; no telemetry is uploaded.</p></section><section class="card"><h2>Last 24 hours</h2><table><thead><tr><th>Target</th><th>Samples</th><th>Availability</th><th>Packet loss</th><th>Average</th><th>P95</th><th>Jitter</th><th>Quality</th><th>Failures</th></tr></thead><tbody>%s</tbody></table></section><section class="card"><h2>Last 7 days</h2><table><thead><tr><th>Target</th><th>Samples</th><th>Availability</th><th>Packet loss</th><th>Average</th><th>P95</th><th>Jitter</th><th>Quality</th><th>Failures</th></tr></thead><tbody>%s</tbody></table></section><p><small>Availability is the percentage of successful checks for each target. Jitter is the average change between consecutive successful latency samples. This is diagnostic evidence, not a contractual SLA measurement.</small></p></div></body></html>`, now.Format("2006-01-02 15:04:05"), statsTable(stats24), statsTable(stats7))
	path := filepath.Join(logDir, "netwatcher_statistics_"+now.Format("20060102_150405")+".html")
	return path, os.WriteFile(path, []byte(page), 0644)
}

func exportLogsZip(logDir string) (string, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return "", err
	}
	output := filepath.Join(logDir, "NetWatcher_Export_"+time.Now().Format("20060102_150405")+".zip")
	f, err := os.Create(output)
	if err != nil {
		return "", err
	}
	zw := zip.NewWriter(f)
	entries, err := os.ReadDir(logDir)
	if err != nil {
		_ = zw.Close()
		_ = f.Close()
		return "", err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		lower := strings.ToLower(name)
		if filepath.Clean(filepath.Join(logDir, name)) == filepath.Clean(output) || strings.HasSuffix(lower, ".zip") {
			continue
		}
		if !(strings.HasSuffix(lower, ".csv") || strings.HasSuffix(lower, ".html") || strings.HasSuffix(lower, ".log")) {
			continue
		}
		input, err := os.Open(filepath.Join(logDir, name))
		if err != nil {
			continue
		}
		info, statErr := input.Stat()
		if statErr != nil {
			_ = input.Close()
			continue
		}
		header, headerErr := zip.FileInfoHeader(info)
		if headerErr != nil {
			_ = input.Close()
			continue
		}
		header.Name = name
		header.Method = zip.Deflate
		writer, createErr := zw.CreateHeader(header)
		if createErr == nil {
			_, _ = io.Copy(writer, input)
		}
		_ = input.Close()
	}
	if err := zw.Close(); err != nil {
		_ = f.Close()
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return output, nil
}
