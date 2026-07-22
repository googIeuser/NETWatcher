package domain

import "time"

type Config struct {
	Language            string   `json:"language"`
	Theme               string   `json:"theme"`
	Interval            float64  `json:"interval_seconds"`
	TimeoutMS           int      `json:"timeout_ms"`
	HighLatencyMS       float64  `json:"high_latency_ms"`
	ConfirmCycles       int      `json:"confirm_cycles"`
	AutoStart           bool     `json:"start_with_windows"`
	StartMinimizedTray  bool     `json:"start_minimized_to_notification_area"`
	AutoMonitor         bool     `json:"start_monitoring_automatically"`
	CloseToTray         bool     `json:"keep_running_in_tray_on_close"`
	AutoCheckUpdates    bool     `json:"automatically_check_for_updates"`
	OutageNotifications bool     `json:"show_outage_notifications"`
	FirstRunComplete    bool     `json:"first_run_setup_completed"`
	CustomTargets       []string `json:"custom_targets"`
	GraphRangeMinutes   int      `json:"graph_range_minutes"`
	LogRetentionDays    int      `json:"log_retention_days"`
}

type Target struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Host   string `json:"host"`
	Kind   string `json:"kind"`
	Mode   string `json:"mode"`
	Custom bool   `json:"custom"`
}

type Result struct {
	Timestamp time.Time
	Target    Target
	Success   bool
	Latency   float64
	Message   string
}

type Sample struct {
	Time    string  `json:"time"`
	Latency float64 `json:"latency"`
	Success bool    `json:"success"`
}

type TargetStatus struct {
	Target     Target   `json:"target"`
	State      string   `json:"state"`
	Latency    float64  `json:"latency"`
	PacketLoss float64  `json:"packetLoss"`
	Jitter     float64  `json:"jitter"`
	LastCheck  string   `json:"lastCheck"`
	Message    string   `json:"message"`
	History    []Sample `json:"history"`
}

type Event struct {
	Time     string `json:"time"`
	Level    string `json:"level"`
	Category string `json:"category"`
	Message  string `json:"message"`
}

type Outage struct {
	Start           string  `json:"start"`
	End             string  `json:"end"`
	Category        string  `json:"category"`
	Details         string  `json:"details"`
	DurationSeconds float64 `json:"durationSeconds"`
	Active          bool    `json:"active"`
}

type Snapshot struct {
	Monitoring      bool           `json:"monitoring"`
	ConnectionState string         `json:"connectionState"`
	ConnectionLabel string         `json:"connectionLabel"`
	QualityScore    int            `json:"qualityScore"`
	AverageLatency  float64        `json:"averageLatency"`
	PacketLoss      float64        `json:"packetLoss"`
	Jitter          float64        `json:"jitter"`
	Samples         int            `json:"samples"`
	Outages         int            `json:"outages"`
	Targets         []TargetStatus `json:"targets"`
	RecentEvents    []Event        `json:"recentEvents"`
	UpdatedAt       string         `json:"updatedAt"`
	Version         string         `json:"version"`
}

type Measurement struct {
	Timestamp  time.Time
	TargetID   string
	TargetName string
	Host       string
	Kind       string
	Mode       string
	Success    bool
	Latency    float64
	Message    string
}

type TargetStatistics struct {
	TargetID       string  `json:"targetId"`
	TargetName     string  `json:"targetName"`
	Host           string  `json:"host"`
	Mode           string  `json:"mode"`
	Samples        int     `json:"samples"`
	Successful     int     `json:"successful"`
	PacketLoss     float64 `json:"packetLoss"`
	AverageLatency float64 `json:"averageLatency"`
	MinimumLatency float64 `json:"minimumLatency"`
	MaximumLatency float64 `json:"maximumLatency"`
	P95Latency     float64 `json:"p95Latency"`
	Jitter         float64 `json:"jitter"`
	Uptime         float64 `json:"uptime"`
}

type Statistics struct {
	RangeHours      int                `json:"rangeHours"`
	From            string             `json:"from"`
	To              string             `json:"to"`
	Samples         int                `json:"samples"`
	Successful      int                `json:"successful"`
	PacketLoss      float64            `json:"packetLoss"`
	AverageLatency  float64            `json:"averageLatency"`
	P95Latency      float64            `json:"p95Latency"`
	Jitter          float64            `json:"jitter"`
	Uptime          float64            `json:"uptime"`
	OutageCount     int                `json:"outageCount"`
	OutageSeconds   float64            `json:"outageSeconds"`
	TargetBreakdown []TargetStatistics `json:"targetBreakdown"`
}

type ReportResult struct {
	Kind      string `json:"kind"`
	Path      string `json:"path"`
	CreatedAt string `json:"createdAt"`
}

type UpdateInfo struct {
	Checked        bool   `json:"checked"`
	Available      bool   `json:"available"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	ReleaseURL     string `json:"releaseUrl"`
	PublishedAt    string `json:"publishedAt"`
	Message        string `json:"message"`
}
