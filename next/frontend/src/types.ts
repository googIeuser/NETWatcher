export type PageId = 'dashboard' | 'statistics' | 'outages' | 'reports' | 'targets' | 'settings'

export interface Settings {
  language: 'en' | 'tr'
  theme: 'light' | 'dark'
  interval_seconds: number
  timeout_ms: number
  high_latency_ms: number
  confirm_cycles: number
  start_with_windows: boolean
  start_minimized_to_notification_area: boolean
  start_monitoring_automatically: boolean
  keep_running_in_tray_on_close: boolean
  automatically_check_for_updates: boolean
  show_outage_notifications: boolean
  first_run_setup_completed: boolean
  custom_targets: string[]
  graph_range_minutes: number
  log_retention_days: number
}
export interface Sample { time: string; latency: number; success: boolean }
export interface Target { id: string; name: string; host: string; kind: string; mode: string; custom: boolean }
export interface TargetStatus { target: Target; state: 'waiting'|'online'|'offline'; latency: number; packetLoss: number; jitter: number; lastCheck: string; message: string; history: Sample[] }
export interface EventItem { time: string; level: string; category: string; message: string }
export interface Snapshot { monitoring: boolean; connectionState: string; connectionLabel: string; qualityScore: number; averageLatency: number; packetLoss: number; jitter: number; samples: number; outages: number; targets: TargetStatus[]; recentEvents: EventItem[]; updatedAt: string; version: string }
export interface Outage { start: string; end: string; category: string; details: string; durationSeconds: number; active: boolean }
export interface TargetStatistics { targetId: string; targetName: string; host: string; mode: string; samples: number; successful: number; packetLoss: number; averageLatency: number; minimumLatency: number; maximumLatency: number; p95Latency: number; jitter: number; uptime: number }
export interface Statistics { rangeHours: number; from: string; to: string; samples: number; successful: number; packetLoss: number; averageLatency: number; p95Latency: number; jitter: number; uptime: number; outageCount: number; outageSeconds: number; targetBreakdown: TargetStatistics[] }
export interface ReportResult { kind: string; path: string; createdAt: string }
export interface UpdateInfo { checked: boolean; available: boolean; currentVersion: string; latestVersion: string; releaseUrl: string; publishedAt: string; message: string }
