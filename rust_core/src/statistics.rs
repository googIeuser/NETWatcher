use std::collections::BTreeMap;

use chrono::{DateTime, Utc};

use crate::models::{Measurement, Outage, Statistics, TargetStatistics};

#[derive(Default)]
struct TargetAccumulator {
    name: String,
    host: String,
    mode: String,
    samples: usize,
    successful: usize,
    latencies: Vec<f64>,
}

fn percentile95(values: &[f64]) -> f64 {
    if values.is_empty() {
        return 0.0;
    }
    let mut sorted = values.to_vec();
    sorted.sort_by(f64::total_cmp);
    let index = (((sorted.len() - 1) as f64) * 0.95).round() as usize;
    sorted[index.min(sorted.len() - 1)]
}

fn average(values: &[f64]) -> f64 {
    if values.is_empty() {
        0.0
    } else {
        values.iter().sum::<f64>() / values.len() as f64
    }
}

fn jitter(values: &[f64]) -> f64 {
    if values.len() < 2 {
        return 0.0;
    }
    values
        .windows(2)
        .map(|pair| (pair[1] - pair[0]).abs())
        .sum::<f64>()
        / (values.len() - 1) as f64
}

pub fn build(
    range_hours: i64,
    since: DateTime<Utc>,
    now: DateTime<Utc>,
    measurements: &[Measurement],
    outages: &[Outage],
) -> Statistics {
    let mut grouped: BTreeMap<String, TargetAccumulator> = BTreeMap::new();
    for measurement in measurements {
        let item = grouped.entry(measurement.target_id.clone()).or_default();
        item.name = measurement.target_name.clone();
        item.host = measurement.host.clone();
        item.mode = measurement.mode.clone();
        item.samples += 1;
        if measurement.success {
            item.successful += 1;
            item.latencies.push(measurement.latency);
        }
    }

    let mut breakdown = Vec::with_capacity(grouped.len());
    for (target_id, item) in grouped {
        let failures = item.samples.saturating_sub(item.successful);
        let packet_loss = if item.samples == 0 {
            0.0
        } else {
            failures as f64 / item.samples as f64 * 100.0
        };
        let minimum_latency = item
            .latencies
            .iter()
            .copied()
            .min_by(f64::total_cmp)
            .unwrap_or(0.0);
        let maximum_latency = item
            .latencies
            .iter()
            .copied()
            .max_by(f64::total_cmp)
            .unwrap_or(0.0);
        breakdown.push(TargetStatistics {
            target_id,
            target_name: item.name,
            host: item.host,
            mode: item.mode,
            samples: item.samples,
            successful: item.successful,
            packet_loss,
            average_latency: average(&item.latencies),
            minimum_latency,
            maximum_latency,
            p95_latency: percentile95(&item.latencies),
            jitter: jitter(&item.latencies),
            uptime: if item.samples == 0 {
                0.0
            } else {
                item.successful as f64 / item.samples as f64 * 100.0
            },
        });
    }

    let samples = measurements.len();
    let successful = measurements.iter().filter(|item| item.success).count();
    let latencies: Vec<f64> = measurements
        .iter()
        .filter(|item| item.success)
        .map(|item| item.latency)
        .collect();
    Statistics {
        range_hours,
        from: since.to_rfc3339(),
        to: now.to_rfc3339(),
        samples,
        successful,
        packet_loss: if samples == 0 {
            0.0
        } else {
            (samples - successful) as f64 / samples as f64 * 100.0
        },
        average_latency: average(&latencies),
        p95_latency: percentile95(&latencies),
        jitter: jitter(&latencies),
        uptime: if samples == 0 {
            0.0
        } else {
            successful as f64 / samples as f64 * 100.0
        },
        outage_count: outages.len(),
        outage_seconds: outages.iter().map(|item| item.duration_seconds).sum(),
        target_breakdown: breakdown,
    }
}
