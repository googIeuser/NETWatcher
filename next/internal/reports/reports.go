package reports

import (
	"bytes"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
	"time"

	"netwatcher/next/internal/domain"
	"netwatcher/next/internal/storage"
)

type labels struct{ connectionTitle, evidenceTitle, connectionSubtitle, evidenceSubtitle, print, rangeLabel, hours, uptime, loss, average, outages, quality, assessment, targetStats, target, mode, samples, p95, jitter, outageHistory, start, end, category, duration, details, noOutages, generated, local, active, lessSecond string }

func getLabels(language string) labels {
	if language == "tr" {
		return labels{"NetWatcher Bağlantı Raporu", "NetWatcher ISS Kanıt Raporu", "Yerel bağlantı ölçümleri ve hedef tanılamaları", "ISS veya düzenleyici kurum başvurusu için düzenli bağlantı kanıtı", "Yazdır / PDF Kaydet", "Aralık", "saat", "Çalışma oranı", "Paket kaybı", "Ortalama gecikme", "Kesinti", "Kalite puanı", "Otomatik değerlendirme", "Hedef istatistikleri", "Hedef", "Tür", "Ölçüm", "P95", "Jitter", "Kesinti geçmişi", "Başlangıç", "Bitiş", "Kategori", "Süre", "Açıklama", "Bu aralıkta kesinti kaydı yok.", "Oluşturulma", "Tüm ölçümler yerel olarak saklanır.", "Aktif", "< 1 saniye"}
	}
	return labels{"NetWatcher Connection Report", "NetWatcher ISP Evidence Report", "Local connection measurements and target diagnostics", "Structured evidence for an ISP or regulator submission", "Print / Save PDF", "Range", "hours", "Uptime", "Packet loss", "Average latency", "Outages", "Quality score", "Automated assessment", "Target statistics", "Target", "Mode", "Samples", "P95", "Jitter", "Outage history", "Start", "End", "Category", "Duration", "Details", "No outage records in this range.", "Generated", "All measurements remain local.", "Active", "< 1 second"}
}

func Generate(store *storage.Store, kind string, stats domain.Statistics, outages []domain.Outage, snapshot domain.Snapshot, language string) (domain.ReportResult, error) {
	l := getLabels(language)
	name := storage.SafeFileName("NetWatcher_Report") + ".html"
	if kind == "evidence" {
		name = storage.SafeFileName("NetWatcher_ISP_Evidence") + ".html"
	}
	path := filepath.Join(store.ReportsDir(), name)
	var b bytes.Buffer
	title, subtitle := l.connectionTitle, l.connectionSubtitle
	if kind == "evidence" {
		title, subtitle = l.evidenceTitle, l.evidenceSubtitle
	}
	fmt.Fprintf(&b, `<!doctype html><html lang="%s"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width"><title>%s</title><style>%s</style></head><body>`, language, html.EscapeString(title), css)
	fmt.Fprintf(&b, `<header><div><span>NETWATCHER</span><h1>%s</h1><p>%s</p></div><button onclick="print()">%s</button></header>`, html.EscapeString(title), html.EscapeString(subtitle), l.print)
	fmt.Fprintf(&b, `<section class="summary"><article><span>%s</span><strong>%d %s</strong></article><article><span>%s</span><strong>%.2f%%</strong></article><article><span>%s</span><strong>%.2f%%</strong></article><article><span>%s</span><strong>%.1f ms</strong></article><article><span>%s</span><strong>%d</strong></article><article><span>%s</span><strong>%d/100</strong></article></section>`, l.rangeLabel, stats.RangeHours, l.hours, l.uptime, stats.Uptime, l.loss, stats.PacketLoss, l.average, stats.AverageLatency, l.outages, stats.OutageCount, l.quality, snapshot.QualityScore)
	if kind == "evidence" {
		assessment := assessmentText(language, stats)
		note := "Measurements were collected locally by NetWatcher. Raw CSV files are available in the diagnostics export."
		if language == "tr" {
			note = "Ölçümler NetWatcher tarafından yerel olarak toplandı. Ham CSV kayıtları tanılama ZIP paketinde bulunur."
		}
		fmt.Fprintf(&b, `<section><h2>%s</h2><p class="assessment">%s</p><p class="note">%s</p></section>`, l.assessment, html.EscapeString(assessment), html.EscapeString(note))
	}
	fmt.Fprintf(&b, `<section><h2>%s</h2><table><thead><tr><th>%s</th><th>%s</th><th>%s</th><th>%s</th><th>%s</th><th>%s</th><th>%s</th><th>%s</th></tr></thead><tbody>`, l.targetStats, l.target, l.mode, l.samples, l.uptime, l.loss, l.average, l.p95, l.jitter)
	for _, s := range stats.TargetBreakdown {
		fmt.Fprintf(&b, `<tr><td><strong>%s</strong><small>%s</small></td><td>%s</td><td>%d</td><td>%.2f%%</td><td>%.2f%%</td><td>%.1f ms</td><td>%.1f ms</td><td>%.1f ms</td></tr>`, html.EscapeString(s.TargetName), html.EscapeString(s.Host), strings.ToUpper(html.EscapeString(s.Mode)), s.Samples, s.Uptime, s.PacketLoss, s.AverageLatency, s.P95Latency, s.Jitter)
	}
	fmt.Fprintf(&b, `</tbody></table></section><section><h2>%s</h2><table><thead><tr><th>%s</th><th>%s</th><th>%s</th><th>%s</th><th>%s</th></tr></thead><tbody>`, l.outageHistory, l.start, l.end, l.category, l.duration, l.details)
	if len(outages) == 0 {
		fmt.Fprintf(&b, `<tr><td colspan="5" class="empty">%s</td></tr>`, l.noOutages)
	}
	for _, o := range outages {
		end := o.End
		if o.Active {
			end = l.active
		}
		fmt.Fprintf(&b, `<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`, formatTime(o.Start), formatTime(end), html.EscapeString(categoryLabel(language, o.Category)), formatDuration(o.DurationSeconds, l.lessSecond), html.EscapeString(o.Details))
	}
	fmt.Fprintf(&b, `</tbody></table></section><footer>%s %s · NetWatcher %s · %s</footer></body></html>`, l.generated, time.Now().Format("2006-01-02 15:04:05"), html.EscapeString(snapshot.Version), l.local)
	if err := os.WriteFile(path, b.Bytes(), 0o644); err != nil {
		return domain.ReportResult{}, err
	}
	return domain.ReportResult{Kind: kind, Path: path, CreatedAt: time.Now().Format(time.RFC3339)}, nil
}
func categoryLabel(language, category string) string {
	labelsEN := map[string]string{"local": "Local network / gateway", "offline": "ISP / internet outage", "partial": "Partial access", "degraded": "High latency"}
	labelsTR := map[string]string{"local": "Yerel ağ / modem", "offline": "ISS / internet kesintisi", "partial": "Kısmi erişim", "degraded": "Yüksek gecikme"}
	key := strings.ToLower(strings.TrimSpace(category))
	if language == "tr" {
		if value, ok := labelsTR[key]; ok {
			return value
		}
	} else if value, ok := labelsEN[key]; ok {
		return value
	}
	return category
}

func assessmentText(language string, stats domain.Statistics) string {
	if language == "tr" {
		if stats.PacketLoss >= 5 {
			return "Belirgin paket kaybı ölçüldü. Bu durum görüşme, oyun ve yayın deneyimini etkileyebilir."
		}
		if stats.OutageCount > 0 {
			return "Seçilen aralıkta bir veya daha fazla bağlantı kesintisi kaydedildi."
		}
		if stats.P95Latency >= 150 {
			return "Yüksek uç gecikme ölçüldü; yoğunluk veya yönlendirme kararsızlığı ihtimali bulunuyor."
		}
		return "Seçilen aralıkta sürekli bir bağlantı sorunu tespit edilmedi."
	}
	if stats.PacketLoss >= 5 {
		return "Significant packet loss was measured. This can affect calls, games and streaming."
	}
	if stats.OutageCount > 0 {
		return "One or more connectivity interruptions were recorded in the selected range."
	}
	if stats.P95Latency >= 150 {
		return "High tail latency was measured and may indicate congestion or routing instability."
	}
	return "No persistent issue was detected in the selected range."
}
func formatTime(v string) string {
	if v == "" || v == "Active" || v == "Aktif" {
		return v
	}
	t, err := time.Parse(time.RFC3339Nano, v)
	if err != nil {
		return html.EscapeString(v)
	}
	return t.Local().Format("2006-01-02 15:04:05")
}
func formatDuration(v float64, less string) string {
	d := time.Duration(v * float64(time.Second))
	if d < time.Second {
		return less
	}
	return d.Round(time.Second).String()
}

const css = `:root{font-family:Inter,Segoe UI,Arial,sans-serif;color:#172033;background:#f4f7fb}*{box-sizing:border-box}body{max-width:1180px;margin:0 auto;padding:42px}header{display:flex;justify-content:space-between;align-items:center;padding:32px;border-radius:20px;background:linear-gradient(135deg,#176fea,#54a1ff);color:white}header span{font-size:11px;letter-spacing:2px;font-weight:800}h1{margin:8px 0 5px;font-size:30px}header p{margin:0;color:#e6f1ff}button{border:0;background:white;color:#176fea;padding:12px 16px;border-radius:10px;font-weight:700}.summary{display:grid;grid-template-columns:repeat(3,1fr);gap:12px;margin:18px 0}.summary article,section{background:white;border:1px solid #dce4ef;border-radius:15px;padding:20px}.summary span{display:block;color:#6e7a8d;font-size:12px}.summary strong{font-size:22px;display:block;margin-top:7px}section{margin-top:16px}h2{font-size:17px;margin:0 0 16px}table{width:100%;border-collapse:collapse;font-size:12px}th{text-align:left;color:#6e7a8d;font-size:10px;text-transform:uppercase;letter-spacing:.7px;padding:10px;border-bottom:1px solid #dce4ef}td{padding:11px 10px;border-bottom:1px solid #edf1f6}td small{display:block;color:#7d899b;margin-top:4px}.assessment{font-size:16px;line-height:1.6}.note,.empty{color:#6e7a8d}footer{text-align:center;color:#7d899b;font-size:11px;padding:28px}@media print{button{display:none}body{padding:0;background:white}section,.summary article{break-inside:avoid}}`
