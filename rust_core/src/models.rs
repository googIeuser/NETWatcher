use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(default)]
pub struct Config {
    pub language: String,
    pub theme: String,
    #[serde(rename = "interval_seconds")]
    pub interval_seconds: f64,
    #[serde(rename = "timeout_ms")]
    pub timeout_ms: u64,
    #[serde(rename = "high_latency_ms")]
    pub high_latency_ms: f64,
    #[serde(rename = "confirm_cycles")]
    pub confirm_cycles: u32,
    #[serde(rename = "start_with_windows")]
    pub start_with_windows: bool,
    #[serde(rename = "start_minimized_to_notification_area")]
    pub start_minimized_to_notification_area: bool,
    #[serde(rename = "start_monitoring_automatically")]
    pub start_monitoring_automatically: bool,
    #[serde(rename = "keep_running_in_tray_on_close")]
    pub keep_running_in_tray_on_close: bool,
    #[serde(rename = "automatically_check_for_updates")]
    pub automatically_check_for_updates: bool,
    #[serde(rename = "show_outage_notifications")]
    pub show_outage_notifications: bool,
    #[serde(rename = "first_run_setup_completed")]
    pub first_run_setup_completed: bool,
    #[serde(rename = "custom_targets")]
    pub custom_targets: Vec<String>,
    #[serde(rename = "graph_range_minutes")]
    pub graph_range_minutes: u32,
    #[serde(rename = "log_retention_days")]
    pub log_retention_days: u32,
}

impl Default for Config {
    fn default() -> Self {
        Self {
            language: "en".into(),
            theme: "dark".into(),
            interval_seconds: 2.0,
            timeout_ms: 1500,
            high_latency_ms: 150.0,
            confirm_cycles: 2,
            start_with_windows: false,
            start_minimized_to_notification_area: true,
            start_monitoring_automatically: false,
            keep_running_in_tray_on_close: true,
            automatically_check_for_updates: true,
            show_outage_notifications: true,
            first_run_setup_completed: true,
            custom_targets: Vec::new(),
            graph_range_minutes: 5,
            log_retention_days: 30,
        }
    }
}

impl Config {
    pub fn normalised(mut self) -> Self {
        self.language = if self.language.eq_ignore_ascii_case("tr") {
            "tr".into()
        } else {
            "en".into()
        };
        self.theme = if self.theme.eq_ignore_ascii_case("light") {
            "light".into()
        } else {
            "dark".into()
        };
        if !(0.5..=3600.0).contains(&self.interval_seconds) {
            self.interval_seconds = 2.0;
        }
        if !(200..=60_000).contains(&self.timeout_ms) {
            self.timeout_ms = 1500;
        }
        if !(1.0..=60_000.0).contains(&self.high_latency_ms) {
            self.high_latency_ms = 150.0;
        }
        if !(1..=20).contains(&self.confirm_cycles) {
            self.confirm_cycles = 2;
        }
        if !matches!(self.graph_range_minutes, 5 | 30 | 60 | 1440) {
            self.graph_range_minutes = 5;
        }
        if self.log_retention_days > 3650 {
            self.log_retention_days = 30;
        }
        self
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Target {
    pub id: String,
    pub name: String,
    pub host: String,
    pub kind: String,
    pub mode: String,
    pub custom: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct TargetStatus {
    pub target: Target,
    pub state: String,
    pub latency: f64,
    pub packet_loss: f64,
    pub jitter: f64,
    pub last_check: String,
    pub message: String,
    pub history: Vec<Sample>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Sample {
    pub time: String,
    pub latency: f64,
    pub success: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Event {
    pub time: String,
    pub level: String,
    pub category: String,
    pub message: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Snapshot {
    pub monitoring: bool,
    pub connection_state: String,
    pub connection_label: String,
    pub quality_score: i32,
    pub average_latency: f64,
    pub packet_loss: f64,
    pub jitter: f64,
    pub samples: u64,
    pub outages: u64,
    pub targets: Vec<TargetStatus>,
    pub recent_events: Vec<Event>,
    pub updated_at: String,
    pub version: String,
}

impl Default for Snapshot {
    fn default() -> Self {
        Self {
            monitoring: false,
            connection_state: "waiting".into(),
            connection_label: "Waiting".into(),
            quality_score: 0,
            average_latency: 0.0,
            packet_loss: 0.0,
            jitter: 0.0,
            samples: 0,
            outages: 0,
            targets: Vec::new(),
            recent_events: Vec::new(),
            updated_at: chrono::Utc::now().to_rfc3339(),
            version: env!("CARGO_PKG_VERSION").into(),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Measurement {
    pub timestamp: String,
    pub target_id: String,
    pub target_name: String,
    pub host: String,
    pub kind: String,
    pub mode: String,
    pub success: bool,
    pub latency: f64,
    pub message: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Outage {
    pub start: String,
    pub end: String,
    pub category: String,
    pub details: String,
    pub duration_seconds: f64,
    pub active: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct TargetStatistics {
    pub target_id: String,
    pub target_name: String,
    pub host: String,
    pub mode: String,
    pub samples: usize,
    pub successful: usize,
    pub packet_loss: f64,
    pub average_latency: f64,
    pub minimum_latency: f64,
    pub maximum_latency: f64,
    pub p95_latency: f64,
    pub jitter: f64,
    pub uptime: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Statistics {
    pub range_hours: i64,
    pub from: String,
    pub to: String,
    pub samples: usize,
    pub successful: usize,
    pub packet_loss: f64,
    pub average_latency: f64,
    pub p95_latency: f64,
    pub jitter: f64,
    pub uptime: f64,
    pub outage_count: usize,
    pub outage_seconds: f64,
    pub target_breakdown: Vec<TargetStatistics>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ReportResult {
    pub kind: String,
    pub path: String,
    pub created_at: String,
    pub message: String,
}
