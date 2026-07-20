package main

import (
	"testing"
	"time"
)

func TestCalculateRollingMetrics(t *testing.T) {
	now := time.Now()
	samples := []Sample{
		{Time: now.Add(-4 * time.Second), Success: true, Latency: 20},
		{Time: now.Add(-3 * time.Second), Success: true, Latency: 30},
		{Time: now.Add(-2 * time.Second), Success: false},
		{Time: now.Add(-1 * time.Second), Success: true, Latency: 40},
	}
	m := calculateRollingMetrics(samples, now.Add(-10*time.Second))
	if m.Samples != 4 || m.Failures != 1 || m.Successes != 3 {
		t.Fatalf("unexpected counts: %+v", m)
	}
	if m.PacketLoss != 25 {
		t.Fatalf("packet loss = %v", m.PacketLoss)
	}
	if m.Average != 30 {
		t.Fatalf("average = %v", m.Average)
	}
	if m.Jitter != 10 {
		t.Fatalf("jitter = %v", m.Jitter)
	}
	if m.QualityLabel == "" {
		t.Fatal("quality label empty")
	}
}

func TestGraphRangeNormalization(t *testing.T) {
	if normalizeGraphRange(999) != 5 || graphRangeDuration(1440) != 24*time.Hour {
		t.Fatal("graph range normalization failed")
	}
}
