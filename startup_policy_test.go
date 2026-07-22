package main

import "testing"

func TestAutoStartArgument(t *testing.T) {
	tests := []struct {
		name           string
		startMinimized bool
		want           string
	}{
		{name: "startup hidden", startMinimized: true, want: "--startup"},
		{name: "startup visible", startMinimized: false, want: "--app"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := autoStartArgument(tt.startMinimized); got != tt.want {
				t.Fatalf("autoStartArgument(%v) = %q, want %q", tt.startMinimized, got, tt.want)
			}
		})
	}
}

func TestMonitorButtonState(t *testing.T) {
	tests := []struct {
		name       string
		monitoring bool
		wantStart  bool
		wantStop   bool
	}{
		{name: "stopped", monitoring: false, wantStart: true, wantStop: false},
		{name: "running", monitoring: true, wantStart: false, wantStop: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, stop := monitorButtonState(tt.monitoring)
			if start != tt.wantStart || stop != tt.wantStop {
				t.Fatalf("monitorButtonState(%v) = (%v, %v), want (%v, %v)", tt.monitoring, start, stop, tt.wantStart, tt.wantStop)
			}
			if start == stop {
				t.Fatalf("exactly one action must be enabled, got start=%v stop=%v", start, stop)
			}
		})
	}
}

func TestShouldMoveToTrayOnSize(t *testing.T) {
	tests := []struct {
		name     string
		sizeType uintptr
		want     bool
	}{
		{name: "restored", sizeType: 0, want: false},
		{name: "minimized", sizeType: 1, want: true},
		{name: "maximized", sizeType: 2, want: false},
		{name: "max show", sizeType: 3, want: false},
		{name: "max hide", sizeType: 4, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldMoveToTrayOnSize(tt.sizeType); got != tt.want {
				t.Fatalf("shouldMoveToTrayOnSize(%d) = %v, want %v", tt.sizeType, got, tt.want)
			}
		})
	}
}
