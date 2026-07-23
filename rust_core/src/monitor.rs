use std::{
    collections::HashMap,
    io,
    net::{SocketAddr, TcpStream, ToSocketAddrs},
    process::Command,
    sync::{
        atomic::{AtomicBool, Ordering},
        Arc, Mutex, RwLock,
    },
    thread,
    time::{Duration, Instant},
};

use chrono::{DateTime, Utc};

use crate::{
    models::{Config, Event, Measurement, Outage, Sample, Snapshot, Target, TargetStatus},
    storage::Store,
    targets,
};

#[derive(Default)]
struct RuntimeState {
    previous_latency: HashMap<String, f64>,
    history: HashMap<String, Vec<Sample>>,
    confirmed_state: String,
    pending_state: String,
    pending_count: u32,
    active_outage: Option<ActiveOutage>,
}

#[derive(Clone)]
struct ActiveOutage {
    start: DateTime<Utc>,
    category: String,
    details: String,
}

#[derive(Clone)]
pub struct Engine {
    config: Arc<RwLock<Config>>,
    snapshot: Arc<RwLock<Snapshot>>,
    running: Arc<AtomicBool>,
    worker: Arc<Mutex<Option<thread::JoinHandle<()>>>>,
    runtime: Arc<Mutex<RuntimeState>>,
    store: Store,
}

impl Engine {
    pub fn new(config: Config) -> Self {
        let store = Store::new();
        let outage_count = store
            .read_outages(Utc::now() - chrono::Duration::days(36_500))
            .map(|items| items.len() as u64)
            .unwrap_or(0);

        let mut history: HashMap<String, Vec<Sample>> = HashMap::new();
        let mut latest: HashMap<String, Measurement> = HashMap::new();
        if let Ok(measurements) =
            store.read_measurements(Utc::now() - chrono::Duration::hours(24))
        {
            for measurement in measurements {
                history
                    .entry(measurement.target_id.clone())
                    .or_default()
                    .push(Sample {
                        time: measurement.timestamp.clone(),
                        latency: measurement.latency,
                        success: measurement.success,
                    });
                latest.insert(measurement.target_id.clone(), measurement);
            }
        }

        let initial_targets: Vec<TargetStatus> = targets::all_targets(&config.custom_targets)
            .into_iter()
            .map(|target| {
                let samples = history.get(&target.id).cloned().unwrap_or_default();
                match latest.get(&target.id) {
                    Some(measurement) => TargetStatus {
                        target,
                        state: if measurement.success {
                            "online".into()
                        } else {
                            "offline".into()
                        },
                        latency: measurement.latency,
                        packet_loss: if measurement.success { 0.0 } else { 100.0 },
                        jitter: 0.0,
                        last_check: measurement.timestamp.clone(),
                        message: measurement.message.clone(),
                        history: history_for_range(&samples, config.graph_range_minutes),
                    },
                    None => TargetStatus {
                        target,
                        state: "waiting".into(),
                        latency: 0.0,
                        packet_loss: 0.0,
                        jitter: 0.0,
                        last_check: String::new(),
                        message: String::new(),
                        history: Vec::new(),
                    },
                }
            })
            .collect();

        let previous_latencies = initial_targets
            .iter()
            .filter(|status| status.state == "online")
            .map(|status| (status.target.id.clone(), status.latency))
            .collect();

        let online_latencies: Vec<f64> = initial_targets
            .iter()
            .filter(|status| status.state == "online")
            .map(|status| status.latency)
            .collect();

        let mut snapshot = Snapshot::default();
        snapshot.outages = outage_count;
        snapshot.targets = initial_targets;
        if !online_latencies.is_empty() {
            snapshot.average_latency =
                online_latencies.iter().sum::<f64>() / online_latencies.len() as f64;
            snapshot.connection_label = "Previous measurements".into();
        }

        let runtime = RuntimeState {
            previous_latency: previous_latencies,
            history,
            ..RuntimeState::default()
        };

        Self {
            config: Arc::new(RwLock::new(config)),
            snapshot: Arc::new(RwLock::new(snapshot)),
            running: Arc::new(AtomicBool::new(false)),
            worker: Arc::new(Mutex::new(None)),
            runtime: Arc::new(Mutex::new(runtime)),
            store,
        }
    }

    pub fn store(&self) -> Store {
        self.store.clone()
    }

    pub fn config(&self) -> Config {
        self.config.read().expect("config lock poisoned").clone()
    }

    pub fn update_config(&self, config: Config) {
        let range_minutes = config.graph_range_minutes;
        *self.config.write().expect("config lock poisoned") = config;

        let runtime = self.runtime.lock().expect("runtime lock poisoned");
        let mut snapshot = self.snapshot.write().expect("snapshot lock poisoned");
        for status in &mut snapshot.targets {
            status.history = runtime
                .history
                .get(&status.target.id)
                .map(|samples| history_for_range(samples, range_minutes))
                .unwrap_or_default();
        }
    }

    pub fn snapshot(&self) -> Snapshot {
        self.snapshot.read().expect("snapshot lock poisoned").clone()
    }

    pub fn outage_history(&self, days: i64) -> anyhow::Result<Vec<Outage>> {
        let range_days = days.clamp(1, 36_500);
        let since = Utc::now() - chrono::Duration::days(range_days);
        let mut outages = self.store.read_outages(since)?;

        let active = self
            .runtime
            .lock()
            .expect("runtime lock poisoned")
            .active_outage
            .clone();
        if let Some(active) = active {
            let now = Utc::now();
            outages.push(Outage {
                start: active.start.to_rfc3339(),
                end: String::new(),
                category: active.category,
                details: active.details,
                duration_seconds: (now - active.start).num_milliseconds() as f64 / 1000.0,
                active: true,
            });
        }

        outages.sort_by(|a, b| b.start.cmp(&a.start));
        Ok(outages)
    }

    fn push_event(&self, level: &str, category: &str, message: String) {
        let event = Event {
            time: Utc::now().to_rfc3339(),
            level: level.into(),
            category: category.into(),
            message,
        };
        let _ = self.store.append_event(&event);
        let mut snapshot = self.snapshot.write().expect("snapshot lock poisoned");
        snapshot.recent_events.insert(0, event);
        snapshot.recent_events.truncate(50);
    }

    pub fn start(&self) -> Snapshot {
        if self.running.swap(true, Ordering::SeqCst) {
            return self.snapshot();
        }
        {
            let mut snapshot = self.snapshot.write().expect("snapshot lock poisoned");
            snapshot.monitoring = true;
            snapshot.connection_state = "waiting".into();
            snapshot.connection_label = "Starting".into();
        }
        self.push_event("info", "monitor", "Monitoring started.".into());
        let engine = self.clone();
        let handle = thread::spawn(move || {
            while engine.running.load(Ordering::SeqCst) {
                engine.monitor_once();
                let interval = engine.config().interval_seconds.max(0.5);
                let slices = (interval * 10.0).round() as u64;
                for _ in 0..slices.max(1) {
                    if !engine.running.load(Ordering::SeqCst) {
                        break;
                    }
                    thread::sleep(Duration::from_millis(100));
                }
            }
        });
        *self.worker.lock().expect("worker lock poisoned") = Some(handle);
        self.snapshot()
    }

    pub fn stop(&self) -> Snapshot {
        self.running.store(false, Ordering::SeqCst);
        if let Some(handle) = self.worker.lock().expect("worker lock poisoned").take() {
            let _ = handle.join();
        }
        let mut snapshot = self.snapshot.write().expect("snapshot lock poisoned");
        snapshot.monitoring = false;
        snapshot.connection_state = "waiting".into();
        snapshot.connection_label = "Monitoring stopped".into();
        snapshot.updated_at = Utc::now().to_rfc3339();
        snapshot.clone()
    }

    fn classify(statuses: &[TargetStatus], high_latency: f64) -> (&'static str, String) {
        let gateway_failed = statuses
            .iter()
            .any(|item| item.target.kind == "local" && item.state != "online");
        let internet: Vec<_> = statuses
            .iter()
            .filter(|item| item.target.kind != "local")
            .collect();
        let failed = internet.iter().filter(|item| item.state != "online").count();
        let online_latencies: Vec<f64> = internet
            .iter()
            .filter(|item| item.state == "online")
            .map(|item| item.latency)
            .collect();
        let average = if online_latencies.is_empty() {
            0.0
        } else {
            online_latencies.iter().sum::<f64>() / online_latencies.len() as f64
        };
        if gateway_failed {
            ("local", "Default gateway is unreachable.".into())
        } else if !internet.is_empty() && failed == internet.len() {
            ("offline", "Gateway responds but all internet targets failed.".into())
        } else if failed > 0 {
            ("partial", "Some internet targets failed.".into())
        } else if average > high_latency {
            ("degraded", format!("Average latency is {:.1} ms.", average))
        } else {
            ("online", "Connection is normal.".into())
        }
    }

    fn transition_state(&self, candidate: &str, details: &str, confirm_cycles: u32) -> String {
        let mut runtime = self.runtime.lock().expect("runtime lock poisoned");
        if runtime.confirmed_state.is_empty() {
            runtime.confirmed_state = candidate.into();
            if candidate != "online" {
                runtime.active_outage = Some(ActiveOutage {
                    start: Utc::now(),
                    category: candidate.into(),
                    details: details.into(),
                });
            }
            return runtime.confirmed_state.clone();
        }
        if candidate == runtime.confirmed_state {
            runtime.pending_state.clear();
            runtime.pending_count = 0;
            return runtime.confirmed_state.clone();
        }
        if runtime.pending_state == candidate {
            runtime.pending_count += 1;
        } else {
            runtime.pending_state = candidate.into();
            runtime.pending_count = 1;
        }
        if runtime.pending_count < confirm_cycles.max(1) {
            return runtime.confirmed_state.clone();
        }

        let previous = runtime.confirmed_state.clone();
        let now = Utc::now();
        if previous != "online" {
            if let Some(active) = runtime.active_outage.take() {
                let outage = Outage {
                    start: active.start.to_rfc3339(),
                    end: now.to_rfc3339(),
                    category: active.category,
                    details: active.details,
                    duration_seconds: (now - active.start).num_milliseconds() as f64 / 1000.0,
                    active: false,
                };
                if self.store.append_outage(&outage).is_ok() {
                    let mut snapshot = self.snapshot.write().expect("snapshot lock poisoned");
                    snapshot.outages += 1;
                }
            }
        }
        if candidate != "online" {
            runtime.active_outage = Some(ActiveOutage {
                start: now,
                category: candidate.into(),
                details: details.into(),
            });
        }
        runtime.confirmed_state = candidate.into();
        runtime.pending_state.clear();
        runtime.pending_count = 0;
        drop(runtime);

        let message = if candidate == "online" {
            "Connection recovered.".to_owned()
        } else {
            format!("Connection state changed to {candidate}: {details}")
        };
        self.push_event(
            if candidate == "online" { "success" } else { "warning" },
            if candidate == "online" { "recovery" } else { "outage" },
            message,
        );
        candidate.into()
    }

    pub fn monitor_once(&self) -> Snapshot {
        let config = self.config();
        let mut statuses: Vec<_> = targets::all_targets(&config.custom_targets)
            .into_iter()
            .map(|target| check_target(target, config.timeout_ms))
            .collect();

        {
            let mut runtime = self.runtime.lock().expect("runtime lock poisoned");
            let history_cutoff = Utc::now() - chrono::Duration::hours(24);
            for status in &mut statuses {
                if status.state == "online" {
                    if let Some(previous) = runtime
                        .previous_latency
                        .insert(status.target.id.clone(), status.latency)
                    {
                        status.jitter = (status.latency - previous).abs();
                    }
                }

                let target_history = runtime
                    .history
                    .entry(status.target.id.clone())
                    .or_default();
                target_history.extend(status.history.iter().cloned());
                target_history.retain(|sample| {
                    DateTime::parse_from_rfc3339(&sample.time)
                        .map(|value| value.with_timezone(&Utc) >= history_cutoff)
                        .unwrap_or(true)
                });
                if target_history.len() > 50_000 {
                    let remove = target_history.len() - 50_000;
                    target_history.drain(0..remove);
                }
                status.history =
                    history_for_range(target_history, config.graph_range_minutes);
            }
        }

        for status in &statuses {
            let _ = self.store.append_measurement(&Measurement {
                timestamp: status.last_check.clone(),
                target_id: status.target.id.clone(),
                target_name: status.target.name.clone(),
                host: status.target.host.clone(),
                kind: status.target.kind.clone(),
                mode: status.target.mode.clone(),
                success: status.state == "online",
                latency: status.latency,
                message: status.message.clone(),
            });
        }

        let successes = statuses.iter().filter(|status| status.state == "online").count();
        let total = statuses.len();
        let online_latencies: Vec<_> = statuses
            .iter()
            .filter(|status| status.state == "online")
            .map(|status| status.latency)
            .collect();
        let average_latency = if online_latencies.is_empty() {
            0.0
        } else {
            online_latencies.iter().sum::<f64>() / online_latencies.len() as f64
        };
        let packet_loss = if total == 0 {
            0.0
        } else {
            (total - successes) as f64 / total as f64 * 100.0
        };
        let (candidate, details) = Self::classify(&statuses, config.high_latency_ms);
        let state = self.transition_state(candidate, &details, config.confirm_cycles);
        let label = match state.as_str() {
            "online" => "Online",
            "offline" => "Offline",
            "local" => "Local network failure",
            "partial" => "Partial access",
            "degraded" => "High latency",
            _ => "Waiting",
        };
        let max_jitter = statuses.iter().map(|status| status.jitter).fold(0.0, f64::max);
        let quality = (100.0 - packet_loss * 0.75 - average_latency / 12.0 - max_jitter / 6.0)
            .clamp(0.0, 100.0);

        let mut snapshot = self.snapshot.write().expect("snapshot lock poisoned");
        snapshot.monitoring = self.running.load(Ordering::SeqCst);
        snapshot.connection_state = state;
        snapshot.connection_label = label.into();
        snapshot.quality_score = quality.round() as i32;
        snapshot.average_latency = average_latency;
        snapshot.packet_loss = packet_loss;
        snapshot.jitter = max_jitter;
        snapshot.samples += total as u64;
        snapshot.targets = statuses;
        snapshot.updated_at = Utc::now().to_rfc3339();
        snapshot.clone()
    }
}

impl Drop for Engine {
    fn drop(&mut self) {
        self.running.store(false, Ordering::SeqCst);
    }
}

fn history_for_range(samples: &[Sample], range_minutes: u32) -> Vec<Sample> {
    let cutoff = Utc::now() - chrono::Duration::minutes(range_minutes as i64);
    let filtered: Vec<Sample> = samples
        .iter()
        .filter(|sample| {
            DateTime::parse_from_rfc3339(&sample.time)
                .map(|value| value.with_timezone(&Utc) >= cutoff)
                .unwrap_or(true)
        })
        .cloned()
        .collect();

    const MAX_GRAPH_POINTS: usize = 600;
    if filtered.len() <= MAX_GRAPH_POINTS {
        return filtered;
    }

    let step = (filtered.len() + MAX_GRAPH_POINTS - 1) / MAX_GRAPH_POINTS;
    let mut downsampled: Vec<Sample> = filtered.iter().step_by(step).cloned().collect();
    if let Some(last) = filtered.last() {
        if downsampled
            .last()
            .map(|sample| sample.time.as_str())
            != Some(last.time.as_str())
        {
            downsampled.push(last.clone());
        }
    }
    downsampled
}

fn check_target(target: Target, timeout_ms: u64) -> TargetStatus {
    let checked_at = Utc::now().to_rfc3339();
    let result: anyhow::Result<f64> = match target.mode.as_str() {
        "tcp" => check_tcp(&target.host, timeout_ms).map_err(anyhow::Error::from),
        "http" | "https" => check_http(&target.host, timeout_ms),
        _ => check_ping(&target.host, timeout_ms).map_err(anyhow::Error::from),
    };
    match result {
        Ok(latency) => TargetStatus {
            target,
            state: "online".into(),
            latency,
            packet_loss: 0.0,
            jitter: 0.0,
            last_check: checked_at.clone(),
            message: "OK".into(),
            history: vec![Sample {
                time: checked_at,
                latency,
                success: true,
            }],
        },
        Err(error) => TargetStatus {
            target,
            state: "offline".into(),
            latency: 0.0,
            packet_loss: 100.0,
            jitter: 0.0,
            last_check: checked_at.clone(),
            message: error.to_string(),
            history: vec![Sample {
                time: checked_at,
                latency: 0.0,
                success: false,
            }],
        },
    }
}

fn check_tcp(host: &str, timeout_ms: u64) -> io::Result<f64> {
    let address = resolve_address(host)?;
    let started = Instant::now();
    TcpStream::connect_timeout(&address, Duration::from_millis(timeout_ms))?;
    Ok(started.elapsed().as_secs_f64() * 1000.0)
}

fn resolve_address(host: &str) -> io::Result<SocketAddr> {
    host.to_socket_addrs()?
        .next()
        .ok_or_else(|| io::Error::new(io::ErrorKind::AddrNotAvailable, "address not found"))
}

fn check_http(url: &str, timeout_ms: u64) -> anyhow::Result<f64> {
    let client = reqwest::blocking::Client::builder()
        .timeout(Duration::from_millis(timeout_ms))
        .redirect(reqwest::redirect::Policy::limited(5))
        .build()?;
    let started = Instant::now();
    let response = client.get(url).send()?;
    if !response.status().is_success() && !response.status().is_redirection() {
        anyhow::bail!("HTTP {}", response.status());
    }
    Ok(started.elapsed().as_secs_f64() * 1000.0)
}

fn check_ping(host: &str, timeout_ms: u64) -> io::Result<f64> {
    let started = Instant::now();
    let output = if cfg!(windows) {
        Command::new("ping")
            .args(["-n", "1", "-w", &timeout_ms.to_string(), host])
            .output()?
    } else {
        let seconds = ((timeout_ms as f64 / 1000.0).ceil() as u64).max(1);
        Command::new("ping")
            .args(["-c", "1", "-W", &seconds.to_string(), host])
            .output()?
    };
    if !output.status.success() {
        return Err(io::Error::new(io::ErrorKind::TimedOut, "ping failed"));
    }
    Ok(started.elapsed().as_secs_f64() * 1000.0)
}
