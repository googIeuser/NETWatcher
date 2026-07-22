package main

import (
	"math"
	"sort"
	"time"
)

type RollingMetrics struct {
	Samples      int
	Successes    int
	Failures     int
	PacketLoss   float64
	Average      float64
	P95          float64
	Jitter       float64
	QualityScore int
	QualityLabel string
}

func calculateRollingMetrics(samples []Sample, since time.Time) RollingMetrics {
	filtered := make([]Sample, 0, len(samples))
	for _, s := range samples {
		if since.IsZero() || !s.Time.Before(since) {
			filtered = append(filtered, s)
		}
	}
	m := RollingMetrics{Samples: len(filtered)}
	if len(filtered) == 0 {
		m.QualityLabel = "No data"
		return m
	}

	latencies := make([]float64, 0, len(filtered))
	var sum float64
	var previous float64
	havePrevious := false
	var jitterSum float64
	jitterPairs := 0

	for _, s := range filtered {
		if !s.Success {
			m.Failures++
			continue
		}
		m.Successes++
		latencies = append(latencies, s.Latency)
		sum += s.Latency
		if havePrevious {
			jitterSum += math.Abs(s.Latency - previous)
			jitterPairs++
		}
		previous = s.Latency
		havePrevious = true
	}

	m.PacketLoss = 100 * float64(m.Failures) / float64(m.Samples)
	if m.Successes > 0 {
		m.Average = sum / float64(m.Successes)
		sorted := append([]float64(nil), latencies...)
		sort.Float64s(sorted)
		index := int(math.Ceil(float64(len(sorted))*0.95)) - 1
		if index < 0 {
			index = 0
		}
		if index >= len(sorted) {
			index = len(sorted) - 1
		}
		m.P95 = sorted[index]
	}
	if jitterPairs > 0 {
		m.Jitter = jitterSum / float64(jitterPairs)
	}
	m.QualityScore, m.QualityLabel = connectionQuality(m)
	return m
}

func connectionQuality(m RollingMetrics) (int, string) {
	if m.Samples == 0 {
		return 0, "No data"
	}
	score := 100.0
	score -= math.Min(70, m.PacketLoss*7)
	score -= math.Min(20, m.Jitter/3)
	if m.Average > 40 {
		score -= math.Min(20, (m.Average-40)/8)
	}
	if score < 0 {
		score = 0
	}
	rounded := int(math.Round(score))
	switch {
	case rounded >= 90:
		return rounded, "Excellent"
	case rounded >= 75:
		return rounded, "Good"
	case rounded >= 55:
		return rounded, "Fair"
	case rounded >= 30:
		return rounded, "Poor"
	default:
		return rounded, "Unstable"
	}
}

func graphRangeDuration(minutes int) time.Duration {
	switch minutes {
	case 30:
		return 30 * time.Minute
	case 60:
		return time.Hour
	case 1440:
		return 24 * time.Hour
	default:
		return 5 * time.Minute
	}
}

func normalizeGraphRange(minutes int) int {
	switch minutes {
	case 5, 30, 60, 1440:
		return minutes
	default:
		return 5
	}
}

func downsampleSamples(samples []Sample, maxPoints int) []Sample {
	if maxPoints < 2 || len(samples) <= maxPoints {
		return append([]Sample(nil), samples...)
	}
	out := make([]Sample, 0, maxPoints)
	step := float64(len(samples)-1) / float64(maxPoints-1)
	for i := 0; i < maxPoints; i++ {
		idx := int(math.Round(float64(i) * step))
		if idx >= len(samples) {
			idx = len(samples) - 1
		}
		out = append(out, samples[idx])
	}
	return out
}
