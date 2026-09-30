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
    http_client: Arc<reqwest::blocking::Client>,
}

impl Engine {
    pub fn new(config: Config) -> Self {
        let store = Store::new();
        let completed_outages = store
            .read_outages(Utc::now() - chrono::Duration::days(36_500))
            .unwrap_or_default();
        let outage_count = completed_outages.len() as u64;
        let active_outage = store.read_active_outage().ok().flatten().and_then(|saved| {
            if completed_outages.iter().any(|item| item.start == saved.start) {
                let _ = store.clear_active_outage();
                return None;
            }
            DateTime::parse_from_rfc3339(&saved.start)
                .ok()
                .map(|start| ActiveOutage {
                    start: start.with_timezone(&Utc),
                    category: saved.category,
                    details: saved.details,
                })
        });

        let mut history: HashMap<String, Vec<Sample>> = HashMap::new();
        let mut latest: HashMap<String, Measurement> = HashMap::new();
        if let Ok(measurements) = store.read_measurements(
            Utc::now() - chrono::Duration::minutes(config.graph_range_minutes as i64),
        )
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
            confirmed_state: active_outage
                .as_ref()
                .map(|outage| outage.category.clone())
                .unwrap_or_default(),
            active_outage,
            ..RuntimeState::default()
        };

        let http_client = Arc::new(
            reqwest::blocking::Client::builder()
                .timeout(Duration::from_millis(config.timeout_ms))
                .redirect(reqwest::redirect::Policy::limited(5))
                .build()
                .expect("failed to build HTTP client"),
        );

        Self {
            config: Arc::new(RwLock::new(config)),
            snapshot: Arc::new(RwLock::new(snapshot)),
            running: Arc::new(AtomicBool::new(false)),
            worker: Arc::new(Mutex::new(None)),
            runtime: Arc::new(Mutex::new(runtime)),
            store,
            http_client,
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
        let previous_range = self.config().graph_range_minutes;
        let older_history = if range_minutes > previous_range {
            self.store
                .read_measurements(Utc::now() - chrono::Duration::minutes(range_minutes as i64))
                .ok()
        } else {
            None
        };
        *self.config.write().expect("config lock poisoned") = config;

        let mut runtime = self.runtime.lock().expect("runtime lock poisoned");
        let cutoff = (Utc::now() - chrono::Duration::minutes(range_minutes as i64)).to_rfc3339();
        if let Some(measurements) = older_history {
            let mut history: HashMap<String, Vec<Sample>> = HashMap::new();
            for measurement in measurements {
                history
                    .entry(measurement.target_id)
                    .or_default()
                    .push(Sample {
                        time: measurement.timestamp,
                        latency: measurement.latency,
                        success: measurement.success,
                    });
            }
            for (id, samples) in runtime.history.drain() {
                history.entry(id).or_default().extend(samples);
            }
            for samples in history.values_mut() {
                samples.sort_by(|a, b| a.time.cmp(&b.time));
                samples.dedup_by(|a, b| a.time == b.time);
                samples.retain(|sample| sample.time.as_str() >= cutoff.as_str());
            }
            runtime.history = history;
        } else {
            for samples in runtime.history.values_mut() {
                samples.retain(|sample| sample.time.as_str() >= cutoff.as_str());
            }
        }
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

    pub fn clear_outage_history(&self, days: i64) -> anyhow::Result<Vec<Outage>> {
        self.store.clear_outages()?;
        {
            let mut snapshot = self.snapshot.write().expect("snapshot lock poisoned");
            snapshot.outages = 0;
            snapshot.updated_at = Utc::now().to_rfc3339();
        }
        self.outage_history(days)
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
        if !self.running.swap(false, Ordering::SeqCst) {
            return self.snapshot();
        }
        if let Some(handle) = self.worker.lock().expect("worker lock poisoned").take() {
            let _ = handle.join();
        }
        let active = {
            let mut runtime = self.runtime.lock().expect("runtime lock poisoned");
            runtime.confirmed_state.clear();
            runtime.pending_state.clear();
            runtime.pending_count = 0;
            runtime.active_outage.take()
        };
        if let Some(active) = active {
            let now = Utc::now();
            let outage = Outage {
                start: active.start.to_rfc3339(),
                end: now.to_rfc3339(),
                category: active.category,
                details: active.details,
                duration_seconds: (now - active.start).num_milliseconds() as f64 / 1000.0,
                active: false,
            };
            if self.store.append_outage(&outage).is_ok() {
                let _ = self.store.clear_active_outage();
                let mut snapshot = self.snapshot.write().expect("snapshot lock poisoned");
                snapshot.outages += 1;
            }
        }
        {
            let mut snapshot = self.snapshot.write().expect("snapshot lock poisoned");
            snapshot.monitoring = false;
            snapshot.connection_state = "waiting".into();
            snapshot.connection_label = "Monitoring stopped".into();
            snapshot.updated_at = Utc::now().to_rfc3339();
        }
        self.push_event("info", "monitor", "Monitoring stopped.".into());
        self.snapshot()
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
                let active = ActiveOutage {
                    start: Utc::now(),
                    category: candidate.into(),
                    details: details.into(),
                };
                self.persist_active_outage(&active);
                runtime.active_outage = Some(active);
            }
            let confirmed = runtime.confirmed_state.clone();
            drop(runtime);
            if candidate != "online" {
                self.push_event(
                    "warning",
                    "outage",
                    format!("Connection state changed to {candidate}: {details}"),
                );
            }
            return confirmed;
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
                    let _ = self.store.clear_active_outage();
                    let mut snapshot = self.snapshot.write().expect("snapshot lock poisoned");
                    snapshot.outages += 1;
                }
            }
        }
        if candidate != "online" {
            let active = ActiveOutage {
                start: now,
                category: candidate.into(),
                details: details.into(),
            };
            self.persist_active_outage(&active);
            runtime.active_outage = Some(active);
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

    fn persist_active_outage(&self, active: &ActiveOutage) {
        let _ = self.store.write_active_outage(&Outage {
            start: active.start.to_rfc3339(),
            end: String::new(),
            category: active.category.clone(),
            details: active.details.clone(),
            duration_seconds: 0.0,
            active: true,
        });
    }

    pub fn monitor_once(&self) -> Snapshot {
        let config = self.config();
        let client = Arc::clone(&self.http_client);
        let mut statuses: Vec<_> = targets::all_targets(&config.custom_targets)
            .into_iter()
            .map(|target| check_target(target, config.timeout_ms, &client))
            .collect();

        {
            let mut runtime = self.runtime.lock().expect("runtime lock poisoned");
            let history_cutoff =
                Utc::now() - chrono::Duration::minutes(config.graph_range_minutes as i64);
            let cutoff_str = history_cutoff.to_rfc3339();

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
                target_history.retain(|sample| sample.time.as_str() >= cutoff_str.as_str());
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
    let cutoff_str = cutoff.to_rfc3339();
    let first = samples.partition_point(|sample| sample.time.as_str() < cutoff_str.as_str());
    let filtered = &samples[first..];

    const MAX_GRAPH_POINTS: usize = 600;
    if filtered.len() <= MAX_GRAPH_POINTS {
        return filtered.to_vec();
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

fn check_target(
    target: Target,
    timeout_ms: u64,
    http_client: &reqwest::blocking::Client,
) -> TargetStatus {
    let checked_at = Utc::now().to_rfc3339();
    let result: anyhow::Result<f64> = match target.mode.as_str() {
        "tcp" => check_tcp(&target.host, timeout_ms).map_err(anyhow::Error::from),
        "http" | "https" => check_http(&target.host, timeout_ms, http_client),
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

fn check_http(
    url: &str,
    timeout_ms: u64,
    client: &reqwest::blocking::Client,
) -> anyhow::Result<f64> {
    let started = Instant::now();
    let response = client
        .get(url)
        .timeout(Duration::from_millis(timeout_ms))
        .send()?;
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

#[cfg(test)]
mod tests {
    use std::{
        env, fs,
        io::{Read, Write},
        net::TcpListener,
        time::{SystemTime, UNIX_EPOCH},
    };

    use super::*;

    fn slow_http_server(delay: Duration) -> (String, thread::JoinHandle<()>) {
        let listener = TcpListener::bind("127.0.0.1:0").unwrap();
        let address = listener.local_addr().unwrap();
        let server = thread::spawn(move || {
            let (mut connection, _) = listener.accept().unwrap();
            let mut request = [0; 4096];
            connection.read(&mut request).unwrap();
            thread::sleep(delay);
            let _ = connection.write_all(
                b"HTTP/1.1 200 OK\r\nContent-Length: 0\r\nConnection: close\r\n\r\n",
            );
        });
        (format!("http://{address}/"), server)
    }

    #[test]
    fn http_check_uses_shorter_updated_timeout() {
        let (url, server) = slow_http_server(Duration::from_millis(350));
        let client = reqwest::blocking::Client::builder()
            .no_proxy()
            .timeout(Duration::from_secs(2))
            .build()
            .unwrap();
        let result = check_http(&url, 100, &client);
        server.join().unwrap();

        assert!(result.is_err(), "HTTP check ignored the updated timeout");
    }

    #[test]
    fn http_check_uses_longer_updated_timeout() {
        let (url, server) = slow_http_server(Duration::from_millis(350));
        let client = reqwest::blocking::Client::builder()
            .no_proxy()
            .timeout(Duration::from_millis(100))
            .build()
            .unwrap();
        let result = check_http(&url, 700, &client);
        server.join().unwrap();

        assert!(result.is_ok(), "HTTP check kept the original client timeout: {result:?}");
    }

    #[test]
    fn graph_history_loads_older_samples_only_for_wider_ranges() {
        let unique = SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_nanos();
        let dir = env::temp_dir().join(format!("netwatcher-history-range-{unique}"));
        let mut engine = Engine::new(Config::default());
        engine.store = Store::new_at(dir.clone());
        engine.runtime.lock().unwrap().history.clear();

        let old_time = (Utc::now() - chrono::Duration::minutes(30)).to_rfc3339();
        let recent_time = (Utc::now() - chrono::Duration::minutes(1)).to_rfc3339();
        for time in [&old_time, &recent_time] {
            engine.store.append_measurement(&Measurement {
                timestamp: time.clone(),
                target_id: "history-target".into(),
                target_name: "History target".into(),
                host: "127.0.0.1".into(),
                kind: "local".into(),
                mode: "ping".into(),
                success: true,
                latency: 1.0,
                message: "OK".into(),
            }).unwrap();
        }

        let mut config = Config::default();
        config.graph_range_minutes = 60;
        engine.update_config(config.clone());
        assert_eq!(engine.runtime.lock().unwrap().history["history-target"].len(), 2);

        config.graph_range_minutes = 5;
        engine.update_config(config);
        assert_eq!(engine.runtime.lock().unwrap().history["history-target"].len(), 1);
        let _ = fs::remove_dir_all(dir);
    }

    #[test]
    fn active_outage_is_saved_when_it_starts() {
        let unique = SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_nanos();
        let dir = env::temp_dir().join(format!("netwatcher-active-outage-{}-{unique}", std::process::id()));
        let mut engine = Engine::new(Config::default());
        engine.store = Store::new_at(dir.clone());

        engine.transition_state("offline", "test outage", 1);

        let active = dir.join("active_outage.json");
        let saved = fs::read_to_string(&active);
        let _ = fs::remove_dir_all(dir);
        assert!(saved.is_ok(), "active outage was not persisted");
        let saved: Outage = serde_json::from_str(&saved.unwrap()).unwrap();
        assert_eq!(saved.category, "offline");
        assert!(saved.active);
        assert_eq!(engine.snapshot().recent_events[0].category, "outage");
    }

    #[test]
    fn recovery_completes_the_saved_active_outage() {
        let unique = SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_nanos();
        let dir = env::temp_dir().join(format!("netwatcher-recovery-{}-{unique}", std::process::id()));
        let mut engine = Engine::new(Config::default());
        engine.store = Store::new_at(dir.clone());

        engine.transition_state("offline", "test outage", 1);
        let saved: Outage = serde_json::from_slice(&fs::read(dir.join("active_outage.json")).unwrap()).unwrap();
        engine.transition_state("online", "recovered", 1);

        let active_exists = dir.join("active_outage.json").exists();
        let completed = engine.store.read_outages(Utc::now() - chrono::Duration::days(1)).unwrap();
        let _ = fs::remove_dir_all(dir);
        assert!(!active_exists);
        assert_eq!(completed.len(), 1);
        assert_eq!(completed[0].start, saved.start);
        assert!(!completed[0].active);
    }

    #[test]
    fn changed_outage_category_replaces_the_active_record() {
        let unique = SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_nanos();
        let dir = env::temp_dir().join(format!("netwatcher-category-{unique}"));
        let mut engine = Engine::new(Config::default());
        engine.store = Store::new_at(dir.clone());

        engine.transition_state("offline", "initial failure", 1);
        engine.transition_state("local", "gateway failure", 1);

        let saved: Outage =
            serde_json::from_slice(&fs::read(dir.join("active_outage.json")).unwrap()).unwrap();
        let completed = engine
            .store
            .read_outages(Utc::now() - chrono::Duration::days(1))
            .unwrap();
        let _ = fs::remove_dir_all(dir);
        assert_eq!(saved.category, "local");
        assert_eq!(completed.len(), 1);
        assert_eq!(completed[0].category, "offline");
    }

    #[test]
    fn stopping_monitoring_finishes_the_active_outage() {
        let unique = SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_nanos();
        let dir = env::temp_dir().join(format!("netwatcher-stop-{unique}"));
        let mut engine = Engine::new(Config::default());
        engine.store = Store::new_at(dir.clone());

        engine.transition_state("offline", "connection lost", 1);
        engine.running.store(true, Ordering::SeqCst);
        engine.stop();

        let active_exists = dir.join("active_outage.json").exists();
        let completed = engine
            .store
            .read_outages(Utc::now() - chrono::Duration::days(1))
            .unwrap();
        let _ = fs::remove_dir_all(dir);
        assert!(!active_exists);
        assert_eq!(completed.len(), 1);
        assert_eq!(completed[0].category, "offline");
    }
}
