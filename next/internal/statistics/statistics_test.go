package statistics

import (
	"netwatcher/next/internal/domain"
	"testing"
	"time"
)

func TestBuildAggregatesMeasurements(t *testing.T) {
	now := time.Now()
	items := []domain.Measurement{{Timestamp: now, TargetID: "a", TargetName: "A", Success: true, Latency: 20}, {Timestamp: now, TargetID: "a", TargetName: "A", Success: false}, {Timestamp: now, TargetID: "a", TargetName: "A", Success: true, Latency: 40}}
	got := Build(24, items, nil)
	if got.Samples != 3 || got.Successful != 2 {
		t.Fatalf("counts: %+v", got)
	}
	if got.PacketLoss < 33 || got.PacketLoss > 34 {
		t.Fatalf("loss: %v", got.PacketLoss)
	}
	if got.AverageLatency != 30 {
		t.Fatalf("average: %v", got.AverageLatency)
	}
}

func TestOverallJitterDoesNotCompareDifferentTargets(t *testing.T) {
	now := time.Now()
	items := []domain.Measurement{
		{Timestamp: now, TargetID: "a", TargetName: "A", Success: true, Latency: 10},
		{Timestamp: now.Add(time.Millisecond), TargetID: "b", TargetName: "B", Success: true, Latency: 100},
		{Timestamp: now.Add(2 * time.Millisecond), TargetID: "a", TargetName: "A", Success: true, Latency: 12},
		{Timestamp: now.Add(3 * time.Millisecond), TargetID: "b", TargetName: "B", Success: true, Latency: 102},
	}
	got := Build(1, items, nil)
	if got.Jitter != 2 {
		t.Fatalf("expected per-target jitter 2ms, got %v", got.Jitter)
	}
}
