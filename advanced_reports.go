package main

import (
	"encoding/csv"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type diskOutage struct {
	Start    time.Time
	End      time.Time
	Duration time.Duration
	Category string
	Details  string
}

func readOutageHistory(logDir string, since time.Time) ([]diskOutage, error) {
	path := filepath.Join(logDir, "outages.csv")
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.Comma = ';'
	reader.FieldsPerRecord = -1
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	outages := make([]diskOutage, 0, len(rows))
	for index, row := range rows {
		if index == 0 || len(row) < 5 {
			continue
		}
		row[0] = strings.TrimPrefix(row[0], "\ufeff")
		start, startErr := time.Parse(time.RFC3339, row[0])
		end, endErr := time.Parse(time.RFC3339, row[1])
		if startErr != nil || endErr != nil || (!since.IsZero() && end.Before(since)) {
			continue
		}
		seconds, _ := strconv.Atoi(row[2])
		duration := time.Duration(seconds) * time.Second
		if duration <= 0 {
			duration = end.Sub(start)
		}
		outages = append(outages, diskOutage{Start: start, End: end, Duration: duration, Category: row[3], Details: row[4]})
	}
	sort.Slice(outages, func(i, j int) bool { return outages[i].Start.After(outages[j].Start) })
	return outages, nil
}

func outageCategoryLabel(category string) string {
	switch category {
	case "LOCAL_NETWORK":
		return "Local network / modem"
	case "ISP_OUTAGE":
		return "ISP / internet"
	case "DEGRADED":
		return "Partial access"
	case "HIGH_LATENCY":
		return "High latency"
	default:
		return category
	}
}

const modernPageCSS = `:root{color-scheme:light dark;--bg:#0e151d;--card:#18222d;--line:#2b3948;--text:#eaf0f7;--muted:#a9b6c5;--blue:#1d7cff;--cyan:#38bdf8;--green:#34d399;--yellow:#fbbf24;--red:#fb7185}*{box-sizing:border-box}body{font:14px/1.5 "Segoe UI",Arial,sans-serif;margin:0;background:var(--bg);color:var(--text)}.wrap{max-width:1180px;margin:auto;padding:28px}.hero{background:linear-gradient(135deg,var(--blue),var(--cyan));padding:28px;border-radius:20px;color:#fff;box-shadow:0 18px 50px #0005}.hero h1{margin:0 0 8px;font-size:32px}.hero p{margin:0;opacity:.94}.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(190px,1fr));gap:14px;margin-top:18px}.metric,.card{background:var(--card);border:1px solid var(--line);border-radius:16px}.metric{padding:18px}.metric span{display:block;color:var(--muted);font-size:12px;text-transform:uppercase;letter-spacing:.08em}.metric strong{display:block;margin-top:6px;font-size:26px}.card{padding:20px;margin-top:18px;overflow:auto}.card h2{margin:0 0 14px}.table-wrap{overflow:auto}table{width:100%;border-collapse:collapse;min-width:780px}th,td{padding:11px 13px;text-align:left;border-bottom:1px solid var(--line)}th{color:#8ec5ff;font-weight:650}.tag{display:inline-block;padding:4px 9px;border-radius:999px;background:#263546}.tag.bad{background:#4a2430;color:#ffc0ca}.tag.warn{background:#4b3b18;color:#ffe08a}.note{color:var(--muted);margin:18px 4px}.print-button{border:0;border-radius:10px;padding:10px 14px;background:#fff;color:#0758b8;font-weight:700;cursor:pointer}.hero-row{display:flex;align-items:flex-start;justify-content:space-between;gap:20px}@media(max-width:700px){.wrap{padding:14px}.hero-row{display:block}.print-button{margin-top:14px}}@media(prefers-color-scheme:light){:root{--bg:#f3f6fa;--card:#fff;--line:#d9e2ec;--text:#18222d;--muted:#5f7185}.metric,.card{box-shadow:0 8px 28px #3452  }}@media print{body{background:#fff;color:#111}.wrap{max-width:none;padding:0}.hero{box-shadow:none}.print-button{display:none}.metric,.card{break-inside:avoid}}`

func generateOutageHistoryPage(logDir string) (string, error) {
	outages, err := readOutageHistory(logDir, time.Time{})
	if err != nil {
		return "", err
	}
	var rows strings.Builder
	total := time.Duration(0)
	for _, outage := range outages {
		total += outage.Duration
		class := "tag"
		if outage.Category == "ISP_OUTAGE" || outage.Category == "LOCAL_NETWORK" {
			class += " bad"
		} else {
			class += " warn"
		}
		fmt.Fprintf(&rows, `<tr><td>%s</td><td>%s</td><td>%s</td><td><span class="%s">%s</span></td><td>%s</td></tr>`,
			outage.Start.Format("2006-01-02 15:04:05"), outage.End.Format("2006-01-02 15:04:05"), formatDuration(outage.Duration, "en"), class, html.EscapeString(outageCategoryLabel(outage.Category)), html.EscapeString(outage.Details))
	}
	if rows.Len() == 0 {
		rows.WriteString(`<tr><td colspan="5">No completed outages have been recorded.</td></tr>`)
	}
	page := fmt.Sprintf(`<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>NetWatcher Outage History</title><style>%s</style></head><body><div class="wrap"><section class="hero"><div class="hero-row"><div><h1>Outage History</h1><p>Completed connectivity incidents recorded locally by NetWatcher.</p></div><button class="print-button" onclick="window.print()">Print / Save PDF</button></div></section><section class="grid"><div class="metric"><span>Completed incidents</span><strong>%d</strong></div><div class="metric"><span>Total outage time</span><strong>%s</strong></div><div class="metric"><span>Generated</span><strong style="font-size:18px">%s</strong></div></section><section class="card"><h2>Incident timeline</h2><div class="table-wrap"><table><thead><tr><th>Start</th><th>End</th><th>Duration</th><th>Class</th><th>Description</th></tr></thead><tbody>%s</tbody></table></div></section><p class="note">The data comes from the local outages.csv file. No telemetry is uploaded.</p></div></body></html>`, modernPageCSS, len(outages), formatDuration(total, "en"), time.Now().Format("2006-01-02 15:04"), rows.String())
	path := filepath.Join(logDir, "netwatcher_outage_history_"+time.Now().Format("20060102_150405")+".html")
	return path, os.WriteFile(path, []byte(page), 0644)
}

func evidenceStatsRows(stats map[string]*periodStats) (string, int, float64) {
	keys := make([]string, 0, len(stats))
	for key := range stats {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var rows strings.Builder
	totalSamples := 0
	weightedAvailability := 0.0
	for _, key := range keys {
		item := stats[key]
		totalSamples += item.Samples
		weightedAvailability += item.uptime() * float64(item.Samples)
		score, label := connectionQuality(RollingMetrics{Samples: item.Samples, Successes: item.Successes, Failures: item.Samples - item.Successes, PacketLoss: item.loss(), Average: item.average(), P95: item.p95(), Jitter: item.jitter()})
		fmt.Fprintf(&rows, `<tr><td>%s</td><td>%d</td><td>%.3f%%</td><td>%.3f%%</td><td>%.2f ms</td><td>%.2f ms</td><td>%.2f ms</td><td>%s (%d)</td></tr>`, html.EscapeString(item.Name), item.Samples, item.uptime(), item.loss(), item.average(), item.p95(), item.jitter(), label, score)
	}
	if len(keys) == 0 {
		rows.WriteString(`<tr><td colspan="8">No samples are available for this period.</td></tr>`)
	}
	availability := 0.0
	if totalSamples > 0 {
		availability = weightedAvailability / float64(totalSamples)
	}
	return rows.String(), totalSamples, availability
}

func generateEvidenceReport(logDir string, days int) (string, error) {
	if days != 1 && days != 7 && days != 30 {
		days = 7
	}
	now := time.Now()
	since := now.Add(-time.Duration(days) * 24 * time.Hour)
	stats, err := readDiskStatistics(logDir, since)
	if err != nil {
		return "", err
	}
	outages, err := readOutageHistory(logDir, since)
	if err != nil {
		return "", err
	}
	rows, totalSamples, availability := evidenceStatsRows(stats)
	var outageRows strings.Builder
	totalOutage := time.Duration(0)
	for _, outage := range outages {
		totalOutage += outage.Duration
		fmt.Fprintf(&outageRows, `<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`, outage.Start.Format("2006-01-02 15:04:05"), outage.End.Format("2006-01-02 15:04:05"), formatDuration(outage.Duration, "en"), html.EscapeString(outageCategoryLabel(outage.Category)), html.EscapeString(outage.Details))
	}
	if outageRows.Len() == 0 {
		outageRows.WriteString(`<tr><td colspan="5">No completed outages are present in this period.</td></tr>`)
	}
	title := fmt.Sprintf("ISP Evidence Report — Last %d day(s)", days)
	page := fmt.Sprintf(`<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>%s</title><style>%s</style></head><body><div class="wrap"><section class="hero"><div class="hero-row"><div><h1>%s</h1><p>%s — %s. Generated locally from NetWatcher CSV logs.</p></div><button class="print-button" onclick="window.print()">Print / Save PDF</button></div></section><section class="grid"><div class="metric"><span>Total samples</span><strong>%d</strong></div><div class="metric"><span>Weighted availability</span><strong>%.3f%%</strong></div><div class="metric"><span>Completed outages</span><strong>%d</strong></div><div class="metric"><span>Total outage time</span><strong>%s</strong></div></section><section class="card"><h2>Target measurements</h2><div class="table-wrap"><table><thead><tr><th>Target</th><th>Samples</th><th>Availability</th><th>Packet loss</th><th>Average</th><th>P95</th><th>Jitter</th><th>Quality</th></tr></thead><tbody>%s</tbody></table></div></section><section class="card"><h2>Outage events</h2><div class="table-wrap"><table><thead><tr><th>Start</th><th>End</th><th>Duration</th><th>Class</th><th>Description</th></tr></thead><tbody>%s</tbody></table></div></section><p class="note">This diagnostic report is not a contractual SLA measurement. Keep the original CSV files when submitting evidence to an ISP or regulator.</p></div></body></html>`, html.EscapeString(title), modernPageCSS, html.EscapeString(title), since.Format("2006-01-02 15:04"), now.Format("2006-01-02 15:04"), totalSamples, availability, len(outages), formatDuration(totalOutage, "en"), rows, outageRows.String())
	path := filepath.Join(logDir, fmt.Sprintf("netwatcher_evidence_%dd_%s.html", days, now.Format("20060102_150405")))
	return path, os.WriteFile(path, []byte(page), 0644)
}

func cleanupOldLogs(logDir string, retentionDays int) (int, error) {
	if retentionDays == 0 {
		return 0, nil
	}
	if retentionDays < 1 {
		retentionDays = 30
	}
	entries, err := os.ReadDir(logDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	cutoff := time.Now().Add(-time.Duration(retentionDays) * 24 * time.Hour)
	removed := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		lower := strings.ToLower(entry.Name())
		if !(strings.HasPrefix(lower, "samples_") || strings.HasPrefix(lower, "netwatcher_") || strings.HasPrefix(lower, "netwatcher-export_") || strings.HasPrefix(lower, "netwatcher_export_")) {
			continue
		}
		info, infoErr := entry.Info()
		if infoErr != nil || !info.ModTime().Before(cutoff) {
			continue
		}
		if removeErr := os.Remove(filepath.Join(logDir, entry.Name())); removeErr == nil {
			removed++
		}
	}
	return removed, nil
}
