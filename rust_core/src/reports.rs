use std::{
    fs::{self, File},
    io::{Read, Write},
    path::Path,
};

use anyhow::{Context, Result};
use chrono::{Duration, Utc};
use zip::{write::SimpleFileOptions, CompressionMethod, ZipWriter};

use crate::{
    models::{Config, Outage, ReportResult, Snapshot, Statistics, TargetStatistics},
    statistics,
    storage::Store,
};

const PAGE_CSS: &str = r#":root{color-scheme:light dark;--bg:#0d1119;--card:#161d29;--line:#2a3547;--text:#eef3fb;--muted:#9aa8ba;--blue:#3175e7;--cyan:#61a1ff;--green:#42d99a;--yellow:#ffbd59;--red:#ff6d80}*{box-sizing:border-box}body{font:14px/1.55 "Segoe UI",Arial,sans-serif;margin:0;background:var(--bg);color:var(--text)}.wrap{max-width:1180px;margin:auto;padding:28px}.hero{background:linear-gradient(135deg,var(--blue),var(--cyan));padding:28px;border-radius:20px;color:#fff;box-shadow:0 18px 50px #0005}.hero-row{display:flex;align-items:flex-start;justify-content:space-between;gap:20px}.hero h1{margin:0 0 8px;font-size:32px}.hero p{margin:0;opacity:.94}.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(180px,1fr));gap:14px;margin-top:18px}.metric,.card{background:var(--card);border:1px solid var(--line);border-radius:16px}.metric{padding:18px}.metric span{display:block;color:var(--muted);font-size:12px;text-transform:uppercase;letter-spacing:.08em}.metric strong{display:block;margin-top:6px;font-size:25px}.card{padding:20px;margin-top:18px;overflow:auto}.card h2{margin:0 0 14px}.table-wrap{overflow:auto}table{width:100%;border-collapse:collapse;min-width:800px}th,td{padding:11px 13px;text-align:left;border-bottom:1px solid var(--line)}th{color:#8ec5ff;font-weight:650}.tag{display:inline-block;padding:4px 9px;border-radius:999px;background:#263546}.tag.bad{background:#4a2430;color:#ffc0ca}.tag.warn{background:#4b3b18;color:#ffe08a}.note{color:var(--muted);margin:18px 4px}.print-button{border:0;border-radius:10px;padding:10px 14px;background:#fff;color:#0758b8;font-weight:700;cursor:pointer}@media(max-width:700px){.wrap{padding:14px}.hero-row{display:block}.print-button{margin-top:14px}}@media(prefers-color-scheme:light){:root{--bg:#f3f6fa;--card:#fff;--line:#d9e2ec;--text:#18222d;--muted:#5f7185}.metric,.card{box-shadow:0 8px 28px #3452}}@media print{body{background:#fff;color:#111}.wrap{max-width:none;padding:0}.hero{box-shadow:none}.print-button{display:none}.metric,.card{break-inside:avoid}}"#;

fn escape(value: &str) -> String {
    value
        .replace('&', "&amp;")
        .replace('<', "&lt;")
        .replace('>', "&gt;")
        .replace('"', "&quot;")
        .replace('\'', "&#39;")
}

fn duration(seconds: f64) -> String {
    let total = seconds.max(0.0).round() as u64;
    let days = total / 86_400;
    let hours = total % 86_400 / 3_600;
    let minutes = total % 3_600 / 60;
    let seconds = total % 60;
    if days > 0 {
        format!("{days}d {hours}h {minutes}m")
    } else if hours > 0 {
        format!("{hours}h {minutes}m {seconds}s")
    } else if minutes > 0 {
        format!("{minutes}m {seconds}s")
    } else {
        format!("{seconds}s")
    }
}

fn quality(item: &TargetStatistics) -> (&'static str, i32) {
    let score = (100.0 - item.packet_loss * 0.75 - item.average_latency / 12.0 - item.jitter / 6.0)
        .clamp(0.0, 100.0)
        .round() as i32;
    let label = if score >= 90 {
        "Excellent"
    } else if score >= 75 {
        "Good"
    } else if score >= 55 {
        "Fair"
    } else {
        "Poor"
    };
    (label, score)
}

fn category_label(value: &str) -> &'static str {
    match value.to_ascii_lowercase().as_str() {
        "local" | "local_network" => "Local network / modem",
        "offline" | "isp_outage" => "ISP / internet",
        "partial" => "Partial access",
        "degraded" | "high_latency" => "High latency",
        _ => "Connectivity incident",
    }
}

fn statistics_rows(stats: &Statistics, evidence: bool) -> String {
    if stats.target_breakdown.is_empty() {
        let colspan = if evidence { 8 } else { 7 };
        return format!("<tr><td colspan=\"{colspan}\">No measurements are available for this period.</td></tr>");
    }
    stats
        .target_breakdown
        .iter()
        .map(|item| {
            if evidence {
                let (label, score) = quality(item);
                format!(
                    "<tr><td>{}</td><td>{}</td><td>{}</td><td>{:.3}%</td><td>{:.2} ms</td><td>{:.2} ms</td><td>{:.2} ms</td><td>{} ({})</td></tr>",
                    escape(&item.target_name),
                    item.samples,
                    format!("{:.3}%", item.uptime),
                    item.packet_loss,
                    item.average_latency,
                    item.p95_latency,
                    item.jitter,
                    label,
                    score
                )
            } else {
                format!(
                    "<tr><td>{}</td><td>{}</td><td>{}</td><td>{}</td><td>{:.3}%</td><td>{:.2} ms</td><td>{:.2} ms</td></tr>",
                    escape(&item.target_name),
                    escape(&item.host),
                    item.samples,
                    item.samples.saturating_sub(item.successful),
                    item.packet_loss,
                    item.average_latency,
                    item.p95_latency
                )
            }
        })
        .collect::<Vec<_>>()
        .join("")
}

fn outage_rows(outages: &[Outage]) -> String {
    if outages.is_empty() {
        return "<tr><td colspan=\"5\">No completed outages are present in this period.</td></tr>".into();
    }
    outages
        .iter()
        .map(|item| {
            let class = if matches!(item.category.to_ascii_lowercase().as_str(), "offline" | "isp_outage" | "local" | "local_network") {
                "tag bad"
            } else {
                "tag warn"
            };
            format!(
                "<tr><td>{}</td><td>{}</td><td>{}</td><td><span class=\"{}\">{}</span></td><td>{}</td></tr>",
                escape(&item.start),
                escape(&item.end),
                duration(item.duration_seconds),
                class,
                category_label(&item.category),
                escape(&item.details)
            )
        })
        .collect::<Vec<_>>()
        .join("")
}

fn load(store: &Store, hours: i64) -> Result<(Statistics, Vec<Outage>)> {
    let now = Utc::now();
    let hours = hours.max(1);
    let since = now - Duration::hours(hours);
    let measurements = store.read_measurements(since)?;
    let outages = store.read_outages(since)?;
    let stats = statistics::build(hours, since, now, &measurements, &outages);
    Ok((stats, outages))
}

fn result(kind: &str, path: &Path, message: &str) -> ReportResult {
    ReportResult {
        kind: kind.into(),
        path: path.to_string_lossy().into_owned(),
        created_at: Utc::now().to_rfc3339(),
        message: message.into(),
    }
}

pub fn generate_html(store: &Store, hours: i64) -> Result<ReportResult> {
    fs::create_dir_all(store.reports_dir())?;
    let (stats, outages) = load(store, hours)?;
    let page = format!(
        r#"<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>NetWatcher Connection Report</title><style>{}</style></head><body><div class="wrap"><section class="hero"><div class="hero-row"><div><h1>NetWatcher Connection Report</h1><p>{} — {}. Generated locally from NetWatcher measurements.</p></div><button class="print-button" onclick="window.print()">Print / Save PDF</button></div></section><section class="grid"><div class="metric"><span>Total samples</span><strong>{}</strong></div><div class="metric"><span>Availability</span><strong>{:.3}%</strong></div><div class="metric"><span>Average latency</span><strong>{:.2} ms</strong></div><div class="metric"><span>Packet loss</span><strong>{:.3}%</strong></div><div class="metric"><span>Completed outages</span><strong>{}</strong></div><div class="metric"><span>Total outage time</span><strong>{}</strong></div></section><section class="card"><h2>Target summary</h2><div class="table-wrap"><table><thead><tr><th>Target</th><th>Address</th><th>Samples</th><th>Failed</th><th>Packet loss</th><th>Average</th><th>P95</th></tr></thead><tbody>{}</tbody></table></div></section><section class="card"><h2>Outage events</h2><div class="table-wrap"><table><thead><tr><th>Start</th><th>End</th><th>Duration</th><th>Class</th><th>Description</th></tr></thead><tbody>{}</tbody></table></div></section><p class="note">Raw CSV records are stored locally in Documents\\NetWatcherLogs.</p></div></body></html>"#,
        PAGE_CSS,
        escape(&stats.from),
        escape(&stats.to),
        stats.samples,
        stats.uptime,
        stats.average_latency,
        stats.packet_loss,
        stats.outage_count,
        duration(stats.outage_seconds),
        statistics_rows(&stats, false),
        outage_rows(&outages),
    );
    let path = store.reports_dir().join(format!(
        "netwatcher_report_{}h_{}.html",
        hours.max(1),
        Utc::now().format("%Y%m%d_%H%M%S")
    ));
    fs::write(&path, page)?;
    Ok(result("html", &path, "HTML report created and opened."))
}

pub fn generate_evidence(store: &Store, days: i64) -> Result<ReportResult> {
    let days = if matches!(days, 1 | 7 | 30) { days } else { 7 };
    fs::create_dir_all(store.reports_dir())?;
    let (stats, outages) = load(store, days * 24)?;
    let page = format!(
        r#"<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>ISP Evidence Report</title><style>{}</style></head><body><div class="wrap"><section class="hero"><div class="hero-row"><div><h1>ISP Evidence Report — Last {} day(s)</h1><p>{} — {}. Generated locally from NetWatcher CSV logs.</p></div><button class="print-button" onclick="window.print()">Print / Save PDF</button></div></section><section class="grid"><div class="metric"><span>Total samples</span><strong>{}</strong></div><div class="metric"><span>Weighted availability</span><strong>{:.3}%</strong></div><div class="metric"><span>Completed outages</span><strong>{}</strong></div><div class="metric"><span>Total outage time</span><strong>{}</strong></div></section><section class="card"><h2>Target measurements</h2><div class="table-wrap"><table><thead><tr><th>Target</th><th>Samples</th><th>Availability</th><th>Packet loss</th><th>Average</th><th>P95</th><th>Jitter</th><th>Quality</th></tr></thead><tbody>{}</tbody></table></div></section><section class="card"><h2>Outage events</h2><div class="table-wrap"><table><thead><tr><th>Start</th><th>End</th><th>Duration</th><th>Class</th><th>Description</th></tr></thead><tbody>{}</tbody></table></div></section><p class="note">This diagnostic report is not a contractual SLA measurement. Keep the original CSV files when submitting evidence to an ISP or regulator.</p></div></body></html>"#,
        PAGE_CSS,
        days,
        escape(&stats.from),
        escape(&stats.to),
        stats.samples,
        stats.uptime,
        stats.outage_count,
        duration(stats.outage_seconds),
        statistics_rows(&stats, true),
        outage_rows(&outages),
    );
    let path = store.reports_dir().join(format!(
        "netwatcher_evidence_{}d_{}.html",
        days,
        Utc::now().format("%Y%m%d_%H%M%S")
    ));
    fs::write(&path, page)?;
    Ok(result(
        "evidence",
        &path,
        "ISP Evidence Report created and opened.",
    ))
}

fn add_bytes(zip: &mut ZipWriter<File>, name: &str, data: &[u8]) -> Result<()> {
    let options = SimpleFileOptions::default().compression_method(CompressionMethod::Deflated);
    zip.start_file(name.replace('\\', "/"), options)?;
    zip.write_all(data)?;
    Ok(())
}

pub fn export_diagnostics(
    store: &Store,
    config: &Config,
    snapshot: &Snapshot,
    hours: i64,
) -> Result<ReportResult> {
    fs::create_dir_all(store.reports_dir())?;
    let (stats, outages) = load(store, hours.max(1))?;
    let path = store.reports_dir().join(format!(
        "netwatcher_diagnostics_{}h_{}.zip",
        hours.max(1),
        Utc::now().format("%Y%m%d_%H%M%S")
    ));
    let file = File::create(&path)?;
    let mut zip = ZipWriter::new(file);
    add_bytes(
        &mut zip,
        "README.txt",
        format!(
            "NetWatcher diagnostics export\r\nGenerated: {}\r\nAll measurements were collected locally.\r\n",
            Utc::now().to_rfc3339()
        )
        .as_bytes(),
    )?;
    add_bytes(
        &mut zip,
        "summary/statistics.json",
        &serde_json::to_vec_pretty(&stats)?,
    )?;
    add_bytes(
        &mut zip,
        "summary/outages.json",
        &serde_json::to_vec_pretty(&outages)?,
    )?;
    add_bytes(
        &mut zip,
        "summary/snapshot.json",
        &serde_json::to_vec_pretty(snapshot)?,
    )?;
    add_bytes(
        &mut zip,
        "summary/settings.json",
        &serde_json::to_vec_pretty(config)?,
    )?;
    for source in store.csv_files()? {
        let Some(name) = source.file_name().and_then(|value| value.to_str()) else {
            continue;
        };
        let mut data = Vec::new();
        File::open(&source)
            .with_context(|| format!("failed to open {}", source.display()))?
            .read_to_end(&mut data)?;
        add_bytes(&mut zip, &format!("raw/{name}"), &data)?;
    }
    zip.finish()?;
    Ok(result(
        "diagnostics",
        &path,
        "Diagnostics ZIP created in the reports folder.",
    ))
}

#[cfg(test)]
mod tests {
    use std::{fs, path::{Path, PathBuf}};

    use chrono::Duration;
    use crate::models::{Measurement, Outage};

    use super::*;

    fn test_store() -> Store {
        let dir = std::env::temp_dir().join(format!(
            "netwatcher-report-test-{}",
            Utc::now().timestamp_nanos_opt().unwrap_or_default()
        ));
        Store::new_at(PathBuf::from(dir))
    }

    #[test]
    fn generates_html_evidence_and_zip() {
        let store = test_store();
        let now = Utc::now();
        store
            .append_measurement(&Measurement {
                timestamp: now.to_rfc3339(),
                target_id: "cloudflare".into(),
                target_name: "Cloudflare".into(),
                host: "1.1.1.1".into(),
                kind: "internet".into(),
                mode: "ping".into(),
                success: true,
                latency: 15.5,
                message: "OK".into(),
            })
            .unwrap();
        store
            .append_outage(&Outage {
                start: (now - Duration::minutes(2)).to_rfc3339(),
                end: (now - Duration::minutes(1)).to_rfc3339(),
                category: "offline".into(),
                details: "test outage".into(),
                duration_seconds: 60.0,
                active: false,
            })
            .unwrap();
        let html = generate_html(&store, 24).unwrap();
        assert!(Path::new(&html.path).exists());
        let evidence = generate_evidence(&store, 7).unwrap();
        let evidence_text = fs::read_to_string(evidence.path).unwrap();
        assert!(evidence_text.contains("ISP Evidence Report"));
        let diagnostics = export_diagnostics(
            &store,
            &Config::default(),
            &Snapshot::default(),
            168,
        )
        .unwrap();
        assert!(Path::new(&diagnostics.path).exists());
    }
}
