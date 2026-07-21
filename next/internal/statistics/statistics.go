package statistics

import (
	"math"
	"sort"
	"time"

	"netwatcher/next/internal/domain"
)

func Build(hours int, measurements []domain.Measurement, outages []domain.Outage) domain.Statistics {
	if hours <= 0 {
		hours = 24
	}
	from := time.Now().Add(-time.Duration(hours) * time.Hour)
	to := time.Now()
	type bucket struct {
		name, host, mode string
		values           []float64
		success, total   int
	}
	buckets := map[string]*bucket{}
	all := make([]float64, 0)
	successful := 0
	for _, m := range measurements {
		b := buckets[m.TargetID]
		if b == nil {
			b = &bucket{name: m.TargetName, host: m.Host, mode: m.Mode}
			buckets[m.TargetID] = b
		}
		b.total++
		if m.Success {
			b.success++
			successful++
			b.values = append(b.values, m.Latency)
			all = append(all, m.Latency)
		}
	}
	result := domain.Statistics{RangeHours: hours, From: from.Format(time.RFC3339), To: to.Format(time.RFC3339), Samples: len(measurements), Successful: successful}
	if result.Samples > 0 {
		result.PacketLoss = 100 * float64(result.Samples-successful) / float64(result.Samples)
		result.Uptime = 100 * float64(successful) / float64(result.Samples)
	}
	result.AverageLatency = avg(all)
	result.P95Latency = percentile(all, .95)
	targetJitters := make([]float64, 0, len(buckets))
	for _, bucket := range buckets {
		if len(bucket.values) >= 2 {
			targetJitters = append(targetJitters, jitter(bucket.values))
		}
	}
	result.Jitter = avg(targetJitters)
	for _, o := range outages {
		result.OutageCount++
		result.OutageSeconds += o.DurationSeconds
	}
	keys := make([]string, 0, len(buckets))
	for k := range buckets {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, id := range keys {
		b := buckets[id]
		s := domain.TargetStatistics{TargetID: id, TargetName: b.name, Host: b.host, Mode: b.mode, Samples: b.total, Successful: b.success}
		if b.total > 0 {
			s.PacketLoss = 100 * float64(b.total-b.success) / float64(b.total)
			s.Uptime = 100 * float64(b.success) / float64(b.total)
		}
		s.AverageLatency = avg(b.values)
		s.MinimumLatency = min(b.values)
		s.MaximumLatency = max(b.values)
		s.P95Latency = percentile(b.values, .95)
		s.Jitter = jitter(b.values)
		result.TargetBreakdown = append(result.TargetBreakdown, s)
	}
	return result
}
func avg(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	t := 0.
	for _, x := range v {
		t += x
	}
	return t / float64(len(v))
}
func min(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	m := v[0]
	for _, x := range v[1:] {
		if x < m {
			m = x
		}
	}
	return m
}
func max(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	m := v[0]
	for _, x := range v[1:] {
		if x > m {
			m = x
		}
	}
	return m
}
func percentile(v []float64, p float64) float64 {
	if len(v) == 0 {
		return 0
	}
	c := append([]float64(nil), v...)
	sort.Float64s(c)
	i := int(math.Ceil(p*float64(len(c)))) - 1
	if i < 0 {
		i = 0
	}
	if i >= len(c) {
		i = len(c) - 1
	}
	return c[i]
}
func jitter(v []float64) float64 {
	if len(v) < 2 {
		return 0
	}
	d := make([]float64, 0, len(v)-1)
	for i := 1; i < len(v); i++ {
		d = append(d, math.Abs(v[i]-v[i-1]))
	}
	return avg(d)
}
