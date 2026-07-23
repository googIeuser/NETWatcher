use std::{
    collections::{HashMap, HashSet},
    env,
    fs::{self, OpenOptions},
    path::{Path, PathBuf},
    sync::{Arc, Mutex},
};

use anyhow::{Context, Result};
use chrono::{DateTime, NaiveDateTime, TimeZone, Utc};
use csv::{ReaderBuilder, StringRecord, WriterBuilder};

use crate::models::{Event, Measurement, Outage};

#[derive(Clone)]
pub struct Store {
    dir: PathBuf,
    lock: Arc<Mutex<()>>,
}

impl Store {
    pub fn new() -> Self {
        let home = env::var_os("USERPROFILE")
            .or_else(|| env::var_os("HOME"))
            .map(PathBuf::from)
            .unwrap_or_else(|| PathBuf::from("."));
        Self::new_at(home.join("Documents").join("NetWatcherLogs"))
    }

    pub fn new_at(dir: PathBuf) -> Self {
        let _ = fs::create_dir_all(dir.join("Reports"));
        Self {
            dir,
            lock: Arc::new(Mutex::new(())),
        }
    }

    pub fn dir(&self) -> &Path {
        &self.dir
    }

    pub fn reports_dir(&self) -> PathBuf {
        self.dir.join("Reports")
    }

    fn append_csv(&self, path: PathBuf, header: &[&str], row: &[String]) -> Result<()> {
        let _guard = self.lock.lock().expect("storage lock poisoned");
        fs::create_dir_all(&self.dir)?;
        let needs_header = fs::metadata(&path).map(|meta| meta.len() == 0).unwrap_or(true);
        let file = OpenOptions::new().create(true).append(true).open(&path)?;
        let mut writer = WriterBuilder::new().from_writer(file);
        if needs_header {
            writer.write_record(header)?;
        }
        writer.write_record(row)?;
        writer.flush()?;
        Ok(())
    }

    pub fn append_measurement(&self, measurement: &Measurement) -> Result<()> {
        let date = DateTime::parse_from_rfc3339(&measurement.timestamp)
            .map(|value| value.with_timezone(&Utc))
            .unwrap_or_else(|_| Utc::now());
        let path = self
            .dir
            .join(format!("measurements_{}.csv", date.format("%Y-%m-%d")));
        self.append_csv(
            path,
            &[
                "timestamp",
                "target_id",
                "target_name",
                "host",
                "kind",
                "mode",
                "success",
                "latency_ms",
                "message",
            ],
            &[
                measurement.timestamp.clone(),
                measurement.target_id.clone(),
                measurement.target_name.clone(),
                measurement.host.clone(),
                measurement.kind.clone(),
                measurement.mode.clone(),
                measurement.success.to_string(),
                format!("{:.3}", measurement.latency),
                measurement.message.clone(),
            ],
        )
    }

    pub fn append_outage(&self, outage: &Outage) -> Result<()> {
        self.append_csv(
            self.dir.join("outages_v4.csv"),
            &["start", "end", "category", "details", "duration_seconds"],
            &[
                outage.start.clone(),
                outage.end.clone(),
                outage.category.clone(),
                outage.details.clone(),
                format!("{:.3}", outage.duration_seconds),
            ],
        )
    }

    pub fn append_event(&self, event: &Event) -> Result<()> {
        self.append_csv(
            self.dir.join("events_v4.csv"),
            &["time", "level", "category", "message"],
            &[
                event.time.clone(),
                event.level.clone(),
                event.category.clone(),
                event.message.clone(),
            ],
        )
    }

    pub fn clear_outages(&self) -> Result<()> {
        let _guard = self.lock.lock().expect("storage lock poisoned");
        for name in ["outages.csv", "outages_v3.csv", "outages_v4.csv"] {
            let path = self.dir.join(name);
            if path.exists() {
                fs::remove_file(&path)
                    .with_context(|| format!("failed to delete {}", path.display()))?;
            }
        }
        Ok(())
    }

    fn delimiter(path: &Path) -> u8 {
        fs::read_to_string(path)
            .ok()
            .and_then(|data| data.lines().next().map(str::to_owned))
            .map(|line| if line.matches(';').count() > line.matches(',').count() { b';' } else { b',' })
            .unwrap_or(b',')
    }

    fn rows(path: &Path) -> Result<(HashMap<String, usize>, Vec<StringRecord>)> {
        let mut reader = ReaderBuilder::new()
            .delimiter(Self::delimiter(path))
            .flexible(true)
            .from_path(path)?;
        let headers = reader.headers()?.clone();
        let indexes = headers
            .iter()
            .enumerate()
            .map(|(index, value)| {
                (
                    value.trim_start_matches('\u{feff}').trim().to_ascii_lowercase(),
                    index,
                )
            })
            .collect();
        let rows = reader.records().filter_map(|record| record.ok()).collect();
        Ok((indexes, rows))
    }

    fn cell<'a>(row: &'a StringRecord, indexes: &HashMap<String, usize>, key: &str) -> &'a str {
        indexes
            .get(key)
            .and_then(|index| row.get(*index))
            .unwrap_or("")
            .trim()
    }

    fn parse_time(value: &str) -> Option<DateTime<Utc>> {
        let value = value.trim().trim_start_matches('\u{feff}');
        if value.is_empty() {
            return None;
        }
        if let Ok(parsed) = DateTime::parse_from_rfc3339(value) {
            return Some(parsed.with_timezone(&Utc));
        }
        for format in ["%Y-%m-%d %H:%M:%S", "%Y-%m-%dT%H:%M:%S"] {
            if let Ok(parsed) = NaiveDateTime::parse_from_str(value, format) {
                return Some(Utc.from_utc_datetime(&parsed));
            }
        }
        None
    }

    fn normalise_outage_category(value: &str) -> String {
        let upper = value.trim().to_ascii_uppercase();
        match upper.as_str() {
            "LOCAL_NETWORK" | "LOCAL" => "local".to_owned(),
            "ISP_OUTAGE" | "OFFLINE" => "offline".to_owned(),
            "DEGRADED" | "HIGH_LATENCY" => "degraded".to_owned(),
            "PARTIAL" => "partial".to_owned(),
            other => other.to_ascii_lowercase(),
        }
    }

    pub fn read_measurements(&self, since: DateTime<Utc>) -> Result<Vec<Measurement>> {
        let _guard = self.lock.lock().expect("storage lock poisoned");
        if !self.dir.exists() {
            return Ok(Vec::new());
        }
        let mut output = Vec::new();
        for entry in fs::read_dir(&self.dir)? {
            let entry = entry?;
            if !entry.file_type()?.is_file() {
                continue;
            }
            let name = entry.file_name().to_string_lossy().to_ascii_lowercase();
            if !name.ends_with(".csv")
                || !(name.starts_with("measurements_") || name.starts_with("samples_"))
            {
                continue;
            }
            let path = entry.path();
            let (indexes, rows) = Self::rows(&path)
                .with_context(|| format!("failed to read {}", path.display()))?;
            let current = indexes.contains_key("target_id");
            for row in rows {
                let timestamp = Self::cell(&row, &indexes, "timestamp");
                let Some(parsed) = Self::parse_time(timestamp) else {
                    continue;
                };
                if parsed < since {
                    continue;
                }
                let target_name = if current {
                    Self::cell(&row, &indexes, "target_name")
                } else {
                    Self::cell(&row, &indexes, "name")
                };
                let host = Self::cell(&row, &indexes, "host");
                let target_id = if current {
                    Self::cell(&row, &indexes, "target_id").to_owned()
                } else {
                    format!("legacy:{}:{}", target_name, host)
                };
                let kind = if current {
                    Self::cell(&row, &indexes, "kind")
                } else {
                    Self::cell(&row, &indexes, "target_type")
                };
                let mode = if current {
                    Self::cell(&row, &indexes, "mode").to_owned()
                } else if host.starts_with("tcp://") {
                    "tcp".to_owned()
                } else if host.starts_with("https://") {
                    "https".to_owned()
                } else if host.starts_with("http://") {
                    "http".to_owned()
                } else {
                    "ping".to_owned()
                };
                let success = Self::cell(&row, &indexes, "success")
                    .parse::<bool>()
                    .unwrap_or(false);
                let latency = Self::cell(&row, &indexes, "latency_ms")
                    .replace(',', ".")
                    .parse::<f64>()
                    .unwrap_or(0.0);
                output.push(Measurement {
                    timestamp: parsed.to_rfc3339(),
                    target_id,
                    target_name: target_name.to_owned(),
                    host: host.to_owned(),
                    kind: kind.to_owned(),
                    mode,
                    success,
                    latency,
                    message: Self::cell(&row, &indexes, "message").to_owned(),
                });
            }
        }
        output.sort_by(|a, b| a.timestamp.cmp(&b.timestamp));
        Ok(output)
    }

    pub fn read_outages(&self, since: DateTime<Utc>) -> Result<Vec<Outage>> {
        let _guard = self.lock.lock().expect("storage lock poisoned");
        let mut output = Vec::new();
        let mut seen = HashSet::new();

        for name in ["outages.csv", "outages_v3.csv", "outages_v4.csv"] {
            let path = self.dir.join(name);
            if !path.exists() {
                continue;
            }
            let (indexes, rows) = Self::rows(&path)?;
            for row in rows {
                let start_text = Self::cell(&row, &indexes, "start");
                let end_text = Self::cell(&row, &indexes, "end");
                let Some(start) = Self::parse_time(start_text) else {
                    continue;
                };
                let end = Self::parse_time(end_text);
                let effective_end = end.as_ref().unwrap_or(&start);
                if effective_end < &since {
                    continue;
                }

                let duration_text = {
                    let current = Self::cell(&row, &indexes, "duration_seconds");
                    if current.is_empty() {
                        Self::cell(&row, &indexes, "duration")
                    } else {
                        current
                    }
                };
                let mut duration = duration_text
                    .replace(',', ".")
                    .parse::<f64>()
                    .unwrap_or(0.0);
                if duration <= 0.0 {
                    if let Some(end_value) = end.as_ref() {
                        duration = (end_value.clone() - start.clone()).num_milliseconds() as f64 / 1000.0;
                    }
                }

                let category = Self::normalise_outage_category(
                    Self::cell(&row, &indexes, "category"),
                );
                let details = Self::cell(&row, &indexes, "details").to_owned();
                let end_value = end
                    .as_ref()
                    .map(|value| value.to_rfc3339())
                    .unwrap_or_default();
                let key = format!(
                    "{}|{}|{}|{}",
                    start.to_rfc3339(),
                    end_value,
                    category,
                    details
                );
                if !seen.insert(key) {
                    continue;
                }

                output.push(Outage {
                    start: start.to_rfc3339(),
                    end: end_value,
                    category,
                    details,
                    duration_seconds: duration.max(0.0),
                    active: end.is_none(),
                });
            }
        }
        output.sort_by(|a, b| b.start.cmp(&a.start));
        Ok(output)
    }

    pub fn csv_files(&self) -> Result<Vec<PathBuf>> {
        let _guard = self.lock.lock().expect("storage lock poisoned");
        if !self.dir.exists() {
            return Ok(Vec::new());
        }
        let mut files = Vec::new();
        for entry in fs::read_dir(&self.dir)? {
            let entry = entry?;
            if entry.file_type()?.is_file()
                && entry
                    .path()
                    .extension()
                    .and_then(|ext| ext.to_str())
                    .is_some_and(|ext| ext.eq_ignore_ascii_case("csv"))
            {
                files.push(entry.path());
            }
        }
        files.sort();
        Ok(files)
    }
}

#[cfg(test)]
mod tests {
    use super::Store;
    use std::{
        env, fs,
        time::{SystemTime, UNIX_EPOCH},
    };

    #[test]
    fn clear_outages_removes_current_and_legacy_files() {
        let unique = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .expect("clock before unix epoch")
            .as_nanos();
        let dir = env::temp_dir().join(format!(
            "netwatcher-storage-test-{}-{unique}",
            std::process::id()
        ));
        fs::create_dir_all(&dir).expect("create test directory");
        for name in ["outages.csv", "outages_v3.csv", "outages_v4.csv"] {
            fs::write(dir.join(name), "start,end,category,details,duration_seconds\n")
                .expect("write outage file");
        }

        let store = Store::new_at(dir.clone());
        store.clear_outages().expect("clear outage history");

        for name in ["outages.csv", "outages_v3.csv", "outages_v4.csv"] {
            assert!(!dir.join(name).exists(), "{name} should be deleted");
        }
        let _ = fs::remove_dir_all(dir);
    }
}
