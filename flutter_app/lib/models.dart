class NetWatcherConfig {
  const NetWatcherConfig({
    this.language = 'en',
    this.theme = 'dark',
    this.intervalSeconds = 2,
    this.timeoutMs = 1500,
    this.highLatencyMs = 150,
    this.confirmCycles = 2,
    this.startWithWindows = false,
    this.startMinimizedToNotificationArea = true,
    this.startMonitoringAutomatically = false,
    this.keepRunningInTrayOnClose = true,
    this.automaticallyCheckForUpdates = true,
    this.showOutageNotifications = true,
    this.firstRunSetupCompleted = true,
    this.customTargets = const [],
    this.graphRangeMinutes = 5,
    this.logRetentionDays = 30,
  });

  final String language;
  final String theme;
  final double intervalSeconds;
  final int timeoutMs;
  final double highLatencyMs;
  final int confirmCycles;
  final bool startWithWindows;
  final bool startMinimizedToNotificationArea;
  final bool startMonitoringAutomatically;
  final bool keepRunningInTrayOnClose;
  final bool automaticallyCheckForUpdates;
  final bool showOutageNotifications;
  final bool firstRunSetupCompleted;
  final List<String> customTargets;
  final int graphRangeMinutes;
  final int logRetentionDays;

  factory NetWatcherConfig.fromJson(Map<String, dynamic> json) {
    return NetWatcherConfig(
      language: json['language'] as String? ?? 'en',
      theme: json['theme'] as String? ?? 'dark',
      intervalSeconds:
          (json['interval_seconds'] as num?)?.toDouble() ?? 2,
      timeoutMs: (json['timeout_ms'] as num?)?.toInt() ?? 1500,
      highLatencyMs:
          (json['high_latency_ms'] as num?)?.toDouble() ?? 150,
      confirmCycles: (json['confirm_cycles'] as num?)?.toInt() ?? 2,
      startWithWindows: json['start_with_windows'] as bool? ?? false,
      startMinimizedToNotificationArea:
          json['start_minimized_to_notification_area'] as bool? ?? true,
      startMonitoringAutomatically:
          json['start_monitoring_automatically'] as bool? ?? false,
      keepRunningInTrayOnClose:
          json['keep_running_in_tray_on_close'] as bool? ?? true,
      automaticallyCheckForUpdates:
          json['automatically_check_for_updates'] as bool? ?? true,
      showOutageNotifications:
          json['show_outage_notifications'] as bool? ?? true,
      firstRunSetupCompleted:
          json['first_run_setup_completed'] as bool? ?? true,
      customTargets: (json['custom_targets'] as List<dynamic>? ?? const [])
          .map((value) => value.toString())
          .toList(growable: false),
      graphRangeMinutes:
          (json['graph_range_minutes'] as num?)?.toInt() ?? 5,
      logRetentionDays:
          (json['log_retention_days'] as num?)?.toInt() ?? 30,
    );
  }

  Map<String, dynamic> toJson() => {
        'language': language,
        'theme': theme,
        'interval_seconds': intervalSeconds,
        'timeout_ms': timeoutMs,
        'high_latency_ms': highLatencyMs,
        'confirm_cycles': confirmCycles,
        'start_with_windows': startWithWindows,
        'start_minimized_to_notification_area':
            startMinimizedToNotificationArea,
        'start_monitoring_automatically': startMonitoringAutomatically,
        'keep_running_in_tray_on_close': keepRunningInTrayOnClose,
        'automatically_check_for_updates': automaticallyCheckForUpdates,
        'show_outage_notifications': showOutageNotifications,
        'first_run_setup_completed': firstRunSetupCompleted,
        'custom_targets': customTargets,
        'graph_range_minutes': graphRangeMinutes,
        'log_retention_days': logRetentionDays,
      };

  NetWatcherConfig copyWith({
    String? language,
    String? theme,
    double? intervalSeconds,
    int? timeoutMs,
    double? highLatencyMs,
    int? confirmCycles,
    bool? startWithWindows,
    bool? startMinimizedToNotificationArea,
    bool? startMonitoringAutomatically,
    bool? keepRunningInTrayOnClose,
    bool? automaticallyCheckForUpdates,
    bool? showOutageNotifications,
    bool? firstRunSetupCompleted,
    List<String>? customTargets,
    int? graphRangeMinutes,
    int? logRetentionDays,
  }) {
    return NetWatcherConfig(
      language: language ?? this.language,
      theme: theme ?? this.theme,
      intervalSeconds: intervalSeconds ?? this.intervalSeconds,
      timeoutMs: timeoutMs ?? this.timeoutMs,
      highLatencyMs: highLatencyMs ?? this.highLatencyMs,
      confirmCycles: confirmCycles ?? this.confirmCycles,
      startWithWindows: startWithWindows ?? this.startWithWindows,
      startMinimizedToNotificationArea:
          startMinimizedToNotificationArea ??
              this.startMinimizedToNotificationArea,
      startMonitoringAutomatically:
          startMonitoringAutomatically ?? this.startMonitoringAutomatically,
      keepRunningInTrayOnClose:
          keepRunningInTrayOnClose ?? this.keepRunningInTrayOnClose,
      automaticallyCheckForUpdates:
          automaticallyCheckForUpdates ?? this.automaticallyCheckForUpdates,
      showOutageNotifications:
          showOutageNotifications ?? this.showOutageNotifications,
      firstRunSetupCompleted:
          firstRunSetupCompleted ?? this.firstRunSetupCompleted,
      customTargets: customTargets ?? this.customTargets,
      graphRangeMinutes: graphRangeMinutes ?? this.graphRangeMinutes,
      logRetentionDays: logRetentionDays ?? this.logRetentionDays,
    );
  }
}

class TargetInfo {
  const TargetInfo({
    required this.id,
    required this.name,
    required this.host,
    required this.kind,
    required this.mode,
    this.custom = false,
  });

  final String id;
  final String name;
  final String host;
  final String kind;
  final String mode;
  final bool custom;

  factory TargetInfo.fromJson(Map<String, dynamic> json) => TargetInfo(
        id: json['id'] as String? ?? '',
        name: json['name'] as String? ?? '',
        host: json['host'] as String? ?? '',
        kind: json['kind'] as String? ?? 'internet',
        mode: json['mode'] as String? ?? 'ping',
        custom: json['custom'] as bool? ?? false,
      );
}

class TargetStatus {
  const TargetStatus({
    required this.target,
    required this.state,
    required this.latency,
    required this.packetLoss,
    required this.jitter,
    this.message = '',
  });

  final TargetInfo target;
  final String state;
  final double latency;
  final double packetLoss;
  final double jitter;
  final String message;

  factory TargetStatus.fromJson(Map<String, dynamic> json) => TargetStatus(
        target:
            TargetInfo.fromJson(json['target'] as Map<String, dynamic>? ?? {}),
        state: json['state'] as String? ?? 'waiting',
        latency: (json['latency'] as num?)?.toDouble() ?? 0,
        packetLoss: (json['packetLoss'] as num?)?.toDouble() ?? 0,
        jitter: (json['jitter'] as num?)?.toDouble() ?? 0,
        message: json['message'] as String? ?? '',
      );
}

class NetworkEvent {
  const NetworkEvent({
    required this.time,
    required this.level,
    required this.category,
    required this.message,
  });

  final String time;
  final String level;
  final String category;
  final String message;

  factory NetworkEvent.fromJson(Map<String, dynamic> json) => NetworkEvent(
        time: json['time'] as String? ?? '',
        level: json['level'] as String? ?? 'info',
        category: json['category'] as String? ?? 'monitor',
        message: json['message'] as String? ?? '',
      );
}

class NetworkSnapshot {
  const NetworkSnapshot({
    this.monitoring = false,
    this.connectionState = 'waiting',
    this.connectionLabel = 'Waiting',
    this.qualityScore = 0,
    this.averageLatency = 0,
    this.packetLoss = 0,
    this.jitter = 0,
    this.samples = 0,
    this.outages = 0,
    this.targets = const [],
    this.recentEvents = const [],
    this.updatedAt = '',
    this.version = '4.0.1',
  });

  final bool monitoring;
  final String connectionState;
  final String connectionLabel;
  final int qualityScore;
  final double averageLatency;
  final double packetLoss;
  final double jitter;
  final int samples;
  final int outages;
  final List<TargetStatus> targets;
  final List<NetworkEvent> recentEvents;
  final String updatedAt;
  final String version;

  factory NetworkSnapshot.fromJson(Map<String, dynamic> json) {
    return NetworkSnapshot(
      monitoring: json['monitoring'] as bool? ?? false,
      connectionState: json['connectionState'] as String? ?? 'waiting',
      connectionLabel: json['connectionLabel'] as String? ?? 'Waiting',
      qualityScore: (json['qualityScore'] as num?)?.toInt() ?? 0,
      averageLatency:
          (json['averageLatency'] as num?)?.toDouble() ?? 0,
      packetLoss: (json['packetLoss'] as num?)?.toDouble() ?? 0,
      jitter: (json['jitter'] as num?)?.toDouble() ?? 0,
      samples: (json['samples'] as num?)?.toInt() ?? 0,
      outages: (json['outages'] as num?)?.toInt() ?? 0,
      targets: (json['targets'] as List<dynamic>? ?? const [])
          .whereType<Map<String, dynamic>>()
          .map(TargetStatus.fromJson)
          .toList(growable: false),
      recentEvents: (json['recentEvents'] as List<dynamic>? ?? const [])
          .whereType<Map<String, dynamic>>()
          .map(NetworkEvent.fromJson)
          .toList(growable: false),
      updatedAt: json['updatedAt'] as String? ?? '',
      version: json['version'] as String? ?? '4.0.1',
    );
  }
}


class OutageRecord {
  const OutageRecord({
    required this.start,
    required this.end,
    required this.category,
    required this.details,
    required this.durationSeconds,
    required this.active,
  });

  final String start;
  final String end;
  final String category;
  final String details;
  final double durationSeconds;
  final bool active;

  factory OutageRecord.fromJson(Map<String, dynamic> json) => OutageRecord(
        start: json['start'] as String? ?? '',
        end: json['end'] as String? ?? '',
        category: json['category'] as String? ?? 'offline',
        details: json['details'] as String? ?? '',
        durationSeconds:
            (json['durationSeconds'] as num?)?.toDouble() ?? 0,
        active: json['active'] as bool? ?? false,
      );
}


class ReportResult {
  const ReportResult({
    required this.kind,
    required this.path,
    required this.createdAt,
    this.message = '',
  });

  final String kind;
  final String path;
  final String createdAt;
  final String message;

  factory ReportResult.fromJson(Map<String, dynamic> json) => ReportResult(
        kind: json['kind'] as String? ?? 'report',
        path: json['path'] as String? ?? '',
        createdAt: json['createdAt'] as String? ?? '',
        message: json['message'] as String? ?? '',
      );
}
