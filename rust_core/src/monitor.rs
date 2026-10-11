use std::{
    collections::HashMap,
    io,
    net::{IpAddr, SocketAddr, TcpStream, ToSocketAddrs},
    process::Command,
    sync::{
        atomic::{AtomicBool, Ordering},
        mpsc, Arc, Mutex, RwLock,
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
    /// Chart buckets for `history`, updated as new samples arrive.
    charts: HashMap<String, Vec<Sample>>,
    confirmed_state: String,
    pending_state: String,
    pending_count: u32,
    active_outage: Option<ActiveOutage>,
    history_epoch: u64,
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
        let saved_active = store.read_active_outage().ok().flatten();
        let mut outage_count = 0u64;
        let mut saved_active_completed = false;
        let _ = store.for_each_outage(Utc::now() - chrono::Duration::days(36_500), |outage| {
            outage_count += 1;
            if saved_active.as_ref().is_some_and(|saved| saved.start == outage.start) {
                saved_active_completed = true;
            }
        });
        let active_outage = saved_active.and_then(|saved| {
            if saved_active_completed {
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
                    .push(Sample::at(
                        &measurement.timestamp,
                        measurement.latency,
                        measurement.success,
                    ));
                latest.insert(measurement.target_id.clone(), measurement);
            }
        }
        let charts = charts_for_range(&history, config.graph_range_minutes);

        let initial_targets: Vec<TargetStatus> = targets::all_targets(&config.custom_targets)
            .into_iter()
            .map(|target| {
                let samples = charts.get(&target.id).cloned().unwrap_or_default();
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
                        history: samples,
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
        set_history_metadata(&mut snapshot, 1, config.graph_range_minutes);
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
            charts,
            confirmed_state: active_outage
                .as_ref()
                .map(|outage| outage.category.clone())
                .unwrap_or_default(),
            active_outage,
            history_epoch: 1,
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
        let cutoff_ms =
            (Utc::now() - chrono::Duration::minutes(range_minutes as i64)).timestamp_millis();
        if let Some(measurements) = older_history {
            let mut history: HashMap<String, Vec<Sample>> = HashMap::new();
            for measurement in measurements {
                history
                    .entry(measurement.target_id)
                    .or_default()
                    .push(Sample::at(
                        &measurement.timestamp,
                        measurement.latency,
                        measurement.success,
                    ));
            }
            for (id, samples) in runtime.history.drain() {
                history.entry(id).or_default().extend(samples);
            }
            for samples in history.values_mut() {
                samples.sort_by_key(|sample| sample.epoch_ms);
                samples.dedup_by_key(|sample| sample.epoch_ms);
                drop_before(samples, cutoff_ms);
            }
            runtime.history = history;
        } else {
            for samples in runtime.history.values_mut() {
                drop_before(samples, cutoff_ms);
            }
        }
        runtime.charts = charts_for_range(&runtime.history, range_minutes);
        runtime.history_epoch += 1;
        let history_epoch = runtime.history_epoch;
        let mut snapshot = self.snapshot.write().expect("snapshot lock poisoned");
        set_history_metadata(&mut snapshot, history_epoch, range_minutes);
        for status in &mut snapshot.targets {
            status.history = runtime
                .charts
                .get(&status.target.id)
                .cloned()
                .unwrap_or_default();
        }
    }

    pub fn snapshot(&self) -> Snapshot {
        self.snapshot.read().expect("snapshot lock poisoned").clone()
    }

    /// Returns the snapshot with only the data the client does not have yet.
    ///
    /// `history` is the newest chart sample time the client holds and its history epoch.
    /// The bucket containing that time is resent because its sample can still change.
    /// `events_seen` is the event count of the client's last snapshot; only newer events
    /// are sent when all of them are still in the recent list.
    pub fn snapshot_since(&self, history: Option<(i64, u64)>, events_seen: Option<u64>) -> Snapshot {
        let source = self.snapshot.read().expect("snapshot lock poisoned");
        let history_after = history
            .filter(|(_, epoch)| *epoch == source.history_epoch)
            .map(|(after_ms, _)| after_ms);
        let targets = match history_after {
            Some(after_ms) => {
                let bucket_ms = source.history_bucket_seconds.max(1) * 1000;
                let first_bucket = after_ms.div_euclid(bucket_ms);
                source
                    .targets
                    .iter()
                    .map(|status| {
                        let first = status.history.partition_point(|sample| {
                            sample.epoch_ms.div_euclid(bucket_ms) < first_bucket
                        });
                        status.clone_with_history(status.history[first..].to_vec())
                    })
                    .collect()
            }
            None => source.targets.clone(),
        };
        let new_events = events_seen
            .and_then(|seen| source.event_count.checked_sub(seen))
            .map(|count| count as usize)
            .filter(|count| *count <= source.recent_events.len());
        let recent_events = match new_events {
            Some(count) => source.recent_events[..count].to_vec(),
            None => source.recent_events.clone(),
        };
        let mut snapshot = source.clone_with(targets, recent_events);
        snapshot.history_delta = history_after.is_some();
        snapshot.events_delta = new_events.is_some();
        snapshot
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
        snapshot.event_count += 1;
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
                let cycle_started = Instant::now();
                engine.run_cycle();
                let interval = Duration::from_secs_f64(engine.config().interval_seconds.max(0.5));
                // Measure from the start of the cycle so checks do not stretch the interval.
                let next_cycle = cycle_started + interval;
                while engine.running.load(Ordering::SeqCst) {
                    let remaining = next_cycle.saturating_duration_since(Instant::now());
                    if remaining.is_zero() {
                        break;
                    }
                    thread::sleep(remaining.min(Duration::from_millis(100)));
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
        self.run_cycle();
        self.snapshot()
    }

    fn run_cycle(&self) {
        let config = self.config();
        let client = Arc::clone(&self.http_client);
        let mut statuses = check_targets(
            targets::all_targets(&config.custom_targets),
            config.timeout_ms,
            &client,
        );

        let history_epoch = {
            let mut runtime = self.runtime.lock().expect("runtime lock poisoned");
            let runtime = &mut *runtime;
            let cutoff_ms = (Utc::now()
                - chrono::Duration::minutes(config.graph_range_minutes as i64))
            .timestamp_millis();
            let bucket_ms = bucket_seconds(config.graph_range_minutes) * 1000;

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
                let added = status.history.len();
                target_history.append(&mut status.history);
                drop_before(target_history, cutoff_ms);
                if target_history.len() > 50_000 {
                    let remove = target_history.len() - 50_000;
                    target_history.drain(0..remove);
                }
                let chart = runtime.charts.entry(status.target.id.clone()).or_default();
                update_chart(chart, target_history, added, bucket_ms);
                status.history = chart.clone();
            }
            runtime.history_epoch
        };

        let measurements: Vec<Measurement> = statuses
            .iter()
            .map(|status| Measurement {
                timestamp: status.last_check.clone(),
                target_id: status.target.id.clone(),
                target_name: status.target.name.clone(),
                host: status.target.host.clone(),
                kind: status.target.kind.clone(),
                mode: status.target.mode.clone(),
                success: status.state == "online",
                latency: status.latency,
                message: status.message.clone(),
            })
            .collect();
        let _ = self.store.append_measurements(&measurements);

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
        set_history_metadata(&mut snapshot, history_epoch, config.graph_range_minutes);
    }
}

impl Drop for Engine {
    fn drop(&mut self) {
        self.running.store(false, Ordering::SeqCst);
    }
}

const MAX_GRAPH_POINTS: i64 = 600;

/// Width of one chart bucket. Buckets are aligned to the Unix epoch, so completed
/// buckets never change and the UI can receive only the newest ones.
fn bucket_seconds(range_minutes: u32) -> i64 {
    let range_seconds = range_minutes.max(1) as i64 * 60;
    ((range_seconds + MAX_GRAPH_POINTS - 1) / MAX_GRAPH_POINTS).max(1)
}

fn set_history_metadata(snapshot: &mut Snapshot, history_epoch: u64, range_minutes: u32) {
    snapshot.history_epoch = history_epoch;
    snapshot.history_bucket_seconds = bucket_seconds(range_minutes);
    snapshot.history_cutoff_ms =
        (Utc::now() - chrono::Duration::minutes(range_minutes as i64)).timestamp_millis();
    snapshot.history_delta = false;
}

/// Removes samples older than `cutoff_ms` from a time-ordered history.
fn drop_before(samples: &mut Vec<Sample>, cutoff_ms: i64) {
    let first = samples.partition_point(|sample| sample.epoch_ms < cutoff_ms);
    samples.drain(..first);
}

/// Keeps one sample per chart bucket: the latest failure in the bucket if there is one,
/// otherwise the latest sample.
fn history_for_range(samples: &[Sample], range_minutes: u32) -> Vec<Sample> {
    let cutoff_ms =
        (Utc::now() - chrono::Duration::minutes(range_minutes as i64)).timestamp_millis();
    let first = samples.partition_point(|sample| sample.epoch_ms < cutoff_ms);
    bucketed(&samples[first..], bucket_seconds(range_minutes) * 1000)
}

fn charts_for_range(
    history: &HashMap<String, Vec<Sample>>,
    range_minutes: u32,
) -> HashMap<String, Vec<Sample>> {
    history
        .iter()
        .map(|(id, samples)| (id.clone(), history_for_range(samples, range_minutes)))
        .collect()
}

fn bucketed(samples: &[Sample], bucket_ms: i64) -> Vec<Sample> {
    let mut output = Vec::new();
    for sample in samples {
        push_bucketed(&mut output, sample, bucket_ms);
    }
    output
}

fn push_bucketed(chart: &mut Vec<Sample>, sample: &Sample, bucket_ms: i64) {
    let bucket = sample.epoch_ms.div_euclid(bucket_ms);
    match chart.last_mut() {
        Some(last) if last.epoch_ms.div_euclid(bucket_ms) == bucket => {
            if sample.success <= last.success {
                *last = sample.clone();
            }
        }
        _ => chart.push(sample.clone()),
    }
}

/// Brings `chart` up to date after `added` samples were appended to `samples` and old
/// samples were removed from its start. Gives the same result as `bucketed(samples)`
/// without rebuilding every bucket on each monitoring cycle.
fn update_chart(chart: &mut Vec<Sample>, samples: &[Sample], added: usize, bucket_ms: i64) {
    for sample in &samples[samples.len().saturating_sub(added)..] {
        push_bucketed(chart, sample, bucket_ms);
    }
    let Some(oldest) = samples.first() else {
        chart.clear();
        return;
    };
    // Buckets before the oldest kept sample have left the range. The oldest bucket may
    // have lost some of its samples, so it is rebuilt from the samples that remain.
    let first_bucket = oldest.epoch_ms.div_euclid(bucket_ms);
    let stale = chart.partition_point(|sample| sample.epoch_ms.div_euclid(bucket_ms) <= first_bucket);
    let rebuilt = bucketed(
        &samples[..samples.partition_point(|sample| sample.epoch_ms.div_euclid(bucket_ms) <= first_bucket)],
        bucket_ms,
    );
    chart.splice(..stale, rebuilt);
}

/// Checks every target at the same time, so failing targets do not delay each other.
fn check_targets(
    targets: Vec<Target>,
    timeout_ms: u64,
    http_client: &reqwest::blocking::Client,
) -> Vec<TargetStatus> {
    thread::scope(|scope| {
        let checks: Vec<_> = targets
            .into_iter()
            .map(|target| scope.spawn(move || check_target(target, timeout_ms, http_client)))
            .collect();
        checks
            .into_iter()
            .map(|check| check.join().expect("target check panicked"))
            .collect()
    })
}

fn check_target(
    target: Target,
    timeout_ms: u64,
    http_client: &reqwest::blocking::Client,
) -> TargetStatus {
    let now = Utc::now();
    let checked_at = now.to_rfc3339();
    let checked_ms = now.timestamp_millis();
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
            last_check: checked_at,
            message: "OK".into(),
            history: vec![Sample::new(checked_ms, latency, true)],
        },
        Err(error) => TargetStatus {
            target,
            state: "offline".into(),
            latency: 0.0,
            packet_loss: 100.0,
            jitter: 0.0,
            last_check: checked_at,
            message: error.to_string(),
            history: vec![Sample::new(checked_ms, 0.0, false)],
        },
    }
}

fn check_tcp(host: &str, timeout_ms: u64) -> io::Result<f64> {
    let deadline = Instant::now() + Duration::from_millis(timeout_ms);
    let address = match host.parse::<SocketAddr>() {
        Ok(address) => address,
        Err(_) => resolve(host.to_owned(), Duration::from_millis(timeout_ms))?
            .into_iter()
            .next()
            .ok_or_else(|| io::Error::new(io::ErrorKind::AddrNotAvailable, "address not found"))?,
    };
    let started = Instant::now();
    TcpStream::connect_timeout(&address, remaining(deadline))?;
    Ok(started.elapsed().as_secs_f64() * 1000.0)
}

/// Time left until `deadline`, at least one millisecond so a check can still be attempted.
fn remaining(deadline: Instant) -> Duration {
    deadline
        .saturating_duration_since(Instant::now())
        .max(Duration::from_millis(1))
}

/// Resolves a host name on a helper thread, so a slow DNS server cannot hold the
/// monitoring cycle past the check timeout.
fn resolve<A: ToSocketAddrs + Send + 'static>(
    address: A,
    timeout: Duration,
) -> io::Result<Vec<SocketAddr>> {
    let (sender, receiver) = mpsc::channel();
    thread::spawn(move || {
        let _ = sender.send(address.to_socket_addrs().map(Iterator::collect::<Vec<_>>));
    });
    receiver
        .recv_timeout(timeout)
        .unwrap_or_else(|_| Err(io::Error::new(io::ErrorKind::TimedOut, "name resolution timed out")))
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
    #[cfg(windows)]
    {
        // IPv4 targets use the Windows ICMP API instead of starting a ping process.
        let deadline = Instant::now() + Duration::from_millis(timeout_ms);
        let address = match host.parse::<IpAddr>() {
            Ok(address) => Some(address),
            Err(_) => resolve((host.to_owned(), 0), Duration::from_millis(timeout_ms))?
                .into_iter()
                .map(|address| address.ip())
                .find(IpAddr::is_ipv4),
        };
        if let Some(IpAddr::V4(address)) = address {
            if let Some(result) = icmp::echo(address, remaining(deadline)) {
                return result;
            }
        }
    }
    ping_command(host, timeout_ms)
}

/// Pings with the system ping command. Used for IPv6 targets and outside Windows.
fn ping_command(host: &str, timeout_ms: u64) -> io::Result<f64> {
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
    let reported = parse_ping_time(&String::from_utf8_lossy(&output.stdout));
    Ok(reported.unwrap_or_else(|| started.elapsed().as_secs_f64() * 1000.0))
}

#[cfg(windows)]
mod icmp {
    use std::{
        io,
        mem::{size_of, size_of_val},
        net::Ipv4Addr,
        ptr,
        time::{Duration, Instant},
    };

    use windows_sys::Win32::{
        Foundation::INVALID_HANDLE_VALUE,
        NetworkManagement::IpHelper::{
            IcmpCloseHandle, IcmpCreateFile, IcmpSendEcho, ICMP_ECHO_REPLY,
        },
    };

    const IP_SUCCESS: u32 = 0;
    /// Same 32-byte payload as the Windows ping command.
    const PAYLOAD: &[u8; 32] = b"abcdefghijklmnopqrstuvwabcdefghi";

    /// Sends one ICMP echo request and returns the round-trip time in milliseconds.
    /// Returns `None` when the ICMP API cannot be opened.
    pub fn echo(address: Ipv4Addr, timeout: Duration) -> Option<io::Result<f64>> {
        // SAFETY: IcmpCreateFile has no preconditions; the handle is closed below.
        let handle = unsafe { IcmpCreateFile() };
        if handle == INVALID_HANDLE_VALUE {
            return None;
        }
        // Room for the reply, the echoed payload and an ICMP error message, aligned for
        // ICMP_ECHO_REPLY.
        let mut reply = [0u64; (size_of::<ICMP_ECHO_REPLY>() + PAYLOAD.len() + 8 + 64) / 8];
        let started = Instant::now();
        // SAFETY: the request and reply buffers are valid for the sizes passed.
        let replies = unsafe {
            IcmpSendEcho(
                handle,
                u32::from_ne_bytes(address.octets()),
                PAYLOAD.as_ptr().cast(),
                PAYLOAD.len() as u16,
                ptr::null(),
                reply.as_mut_ptr().cast(),
                size_of_val(&reply) as u32,
                timeout.as_millis().clamp(1, u32::MAX as u128) as u32,
            )
        };
        let elapsed = started.elapsed();
        // SAFETY: the handle came from IcmpCreateFile and is closed once.
        unsafe { IcmpCloseHandle(handle) };
        if replies == 0 {
            return Some(Err(io::Error::new(io::ErrorKind::TimedOut, "ping failed")));
        }
        // SAFETY: a successful call writes an ICMP_ECHO_REPLY at the start of the buffer.
        let reply = unsafe { &*(reply.as_ptr() as *const ICMP_ECHO_REPLY) };
        if reply.Status != IP_SUCCESS {
            return Some(Err(io::Error::other("ping failed")));
        }
        Some(Ok(elapsed.as_secs_f64() * 1000.0))
    }
}

/// Reads the round-trip time from ping output, such as `time=12ms`, `time<1ms`,
/// `süre=12ms` or `time=12.3 ms`. Starting the ping process adds its own delay,
/// so the reported time is more accurate than the elapsed time.
fn parse_ping_time(output: &str) -> Option<f64> {
    output.lines().find_map(|line| {
        let bytes = line.as_bytes();
        bytes.iter().enumerate().find_map(|(index, byte)| {
            if *byte != b'=' && *byte != b'<' {
                return None;
            }
            let rest = &line[index + 1..];
            let number_length = rest
                .find(|c: char| !(c.is_ascii_digit() || c == '.' || c == ','))
                .unwrap_or(rest.len());
            if number_length == 0 || !rest[number_length..].trim_start().starts_with("ms") {
                return None;
            }
            rest[..number_length].replace(',', ".").parse::<f64>().ok()
        })
    })
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

    fn http_target(id: &str, url: String) -> Target {
        Target {
            id: id.into(),
            name: id.into(),
            host: url,
            kind: "internet".into(),
            mode: "http".into(),
            custom: true,
        }
    }

    #[test]
    fn targets_are_checked_at_the_same_time() {
        let servers: Vec<_> = (0..3)
            .map(|_| slow_http_server(Duration::from_millis(400)))
            .collect();
        let targets = servers
            .iter()
            .enumerate()
            .map(|(index, (url, _))| http_target(&format!("slow-{index}"), url.clone()))
            .collect();
        let client = reqwest::blocking::Client::builder().no_proxy().build().unwrap();

        let started = Instant::now();
        let statuses = check_targets(targets, 2_000, &client);
        let elapsed = started.elapsed();
        for (_, server) in servers {
            server.join().unwrap();
        }

        assert_eq!(
            statuses.iter().map(|status| status.target.id.as_str()).collect::<Vec<_>>(),
            ["slow-0", "slow-1", "slow-2"]
        );
        assert!(statuses.iter().all(|status| status.state == "online"));
        assert!(elapsed < Duration::from_millis(1_000), "checks ran one after another: {elapsed:?}");
    }

    #[test]
    fn ping_time_is_read_from_the_output() {
        let windows = "Pinging 1.1.1.1 with 32 bytes of data:\r\nReply from 1.1.1.1: bytes=32 time=12ms TTL=57\r\n\r\n    Minimum = 12ms, Maximum = 12ms, Average = 12ms";
        assert_eq!(parse_ping_time(windows), Some(12.0));
        let turkish = "1.1.1.1 yanıtı: bayt=32 süre=7ms TTL=57";
        assert_eq!(parse_ping_time(turkish), Some(7.0));
        assert_eq!(parse_ping_time("Reply from 192.168.1.1: bytes=32 time<1ms TTL=64"), Some(1.0));
        let linux = "64 bytes from 1.1.1.1: icmp_seq=1 ttl=57 time=12.3 ms\n\nrtt min/avg/max/mdev = 12.3/12.3/12.3/0.000 ms";
        assert_eq!(parse_ping_time(linux), Some(12.3));
        assert_eq!(parse_ping_time("Request timed out."), None);
    }

    fn sample_at(epoch_ms: i64, success: bool) -> Sample {
        Sample::new(epoch_ms, if success { 10.0 } else { 0.0 }, success)
    }

    #[test]
    fn samples_are_sent_with_an_rfc3339_time() {
        let value = serde_json::to_value(sample_at(1_700_000_000_123, true)).unwrap();
        assert_eq!(value["time"], "2023-11-14T22:13:20.123Z");
        let parsed: Sample = serde_json::from_value(value).unwrap();
        assert_eq!(parsed.epoch_ms, 1_700_000_000_123);
    }

    #[test]
    fn incremental_chart_matches_a_full_rebuild() {
        let bucket_ms = 6_000;
        let window_ms = 60_000;
        let mut samples = Vec::new();
        let mut chart = Vec::new();
        for index in 0..400i64 {
            let now = 1_700_000_000_000 + index * 2_000;
            samples.push(sample_at(now, index % 7 != 3));
            drop_before(&mut samples, now - window_ms);
            update_chart(&mut chart, &samples, 1, bucket_ms);
            let expected = bucketed(&samples, bucket_ms);
            let key = |chart: &[Sample]| {
                chart.iter().map(|sample| (sample.epoch_ms, sample.success)).collect::<Vec<_>>()
            };
            assert_eq!(key(&chart), key(&expected), "chart differs after sample {index}");
        }
    }

    #[test]
    fn slow_name_resolution_stops_at_the_timeout() {
        struct SlowAddress;
        impl ToSocketAddrs for SlowAddress {
            type Iter = std::vec::IntoIter<SocketAddr>;
            fn to_socket_addrs(&self) -> io::Result<Self::Iter> {
                thread::sleep(Duration::from_secs(2));
                Ok(Vec::new().into_iter())
            }
        }

        let started = Instant::now();
        let result = resolve(SlowAddress, Duration::from_millis(100));

        assert_eq!(result.unwrap_err().kind(), io::ErrorKind::TimedOut);
        assert!(started.elapsed() < Duration::from_secs(1));
    }

    #[cfg(windows)]
    #[test]
    fn icmp_ping_reaches_loopback_without_a_ping_process() {
        let latency = icmp::echo(std::net::Ipv4Addr::LOCALHOST, Duration::from_secs(1))
            .expect("ICMP API is available")
            .expect("loopback answers ping");
        assert!(latency >= 0.0 && latency < 1_000.0);
    }

    #[cfg(windows)]
    #[test]
    fn icmp_ping_fails_for_an_unreachable_address() {
        let result = icmp::echo(std::net::Ipv4Addr::new(192, 0, 2, 1), Duration::from_millis(300))
            .expect("ICMP API is available");
        assert!(result.is_err());
    }

    #[test]
    fn snapshot_since_sends_only_new_events() {
        let engine = Engine::new(Config::default());
        engine.push_event("info", "monitor", "first".into());
        let seen = engine.snapshot().event_count;
        engine.push_event("warning", "outage", "second".into());
        engine.push_event("success", "recovery", "third".into());

        let delta = engine.snapshot_since(None, Some(seen));
        assert!(delta.events_delta);
        assert_eq!(
            delta.recent_events.iter().map(|event| event.message.as_str()).collect::<Vec<_>>(),
            ["third", "second"]
        );

        let full = engine.snapshot_since(None, None);
        assert!(!full.events_delta);
        assert_eq!(full.recent_events.len(), 3);

        let restarted_client = engine.snapshot_since(None, Some(seen + 10));
        assert!(!restarted_client.events_delta);
        assert_eq!(restarted_client.recent_events.len(), 3);
    }

    #[test]
    fn wide_chart_ranges_keep_one_sample_per_bucket() {
        let now = Utc::now().timestamp_millis();
        let samples: Vec<_> = (0..43_200)
            .map(|index| sample_at(now - (43_200 - index) * 2_000, index != 42_000))
            .collect();

        let history = history_for_range(&samples, 24 * 60);

        assert!(history.len() <= 601, "too many chart points: {}", history.len());
        assert!(history.iter().any(|sample| !sample.success), "failure was hidden by bucketing");
        assert_eq!(history.last().unwrap().epoch_ms, samples.last().unwrap().epoch_ms);
    }

    #[test]
    fn snapshot_since_sends_only_new_chart_buckets() {
        let engine = Engine::new(Config::default());
        let now = Utc::now().timestamp_millis();
        let history: Vec<_> = (0..10).map(|index| sample_at(now - (10 - index) * 2_000, true)).collect();
        let epoch = {
            let mut snapshot = engine.snapshot.write().unwrap();
            let mut status = snapshot.targets[0].clone();
            status.history = history.clone();
            snapshot.targets = vec![status];
            snapshot.history_epoch
        };

        let delta = engine.snapshot_since(Some((history[7].epoch_ms, epoch)), None);
        assert!(delta.history_delta);
        assert_eq!(delta.targets[0].history.len(), 3);
        assert_eq!(delta.targets[0].history[0].epoch_ms, history[7].epoch_ms);

        let full = engine.snapshot_since(Some((history[7].epoch_ms, epoch + 1)), None);
        assert!(!full.history_delta);
        assert_eq!(full.targets[0].history.len(), 10);
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
