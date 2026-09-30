import 'dart:async';

import 'package:flutter/foundation.dart';

import 'core_service.dart';
import 'models.dart';
import 'process_core_service.dart';
import 'windows_startup.dart';

class AppState extends ChangeNotifier {
  AppState._(
    this._resolvedService, {
    required bool pollSnapshots,
    required bool manageWindowsStartup,
    Future<void> Function(NetworkEvent)? onOutageEvent,
    String? startupFailure,
  })  : _pollSnapshots = pollSnapshots,
        _manageWindowsStartup = manageWindowsStartup,
        _onOutageEvent = onOutageEvent,
        _startupFailure = startupFailure;

  final CoreService? _resolvedService;
  final String? _startupFailure;
  CoreService get _service =>
      _resolvedService ??
      (throw StateError(_startupFailure ?? 'NetWatcher core is unavailable.'));
  final bool _pollSnapshots;
  final bool _manageWindowsStartup;
  final Future<void> Function(NetworkEvent)? _onOutageEvent;
  Timer? _pollTimer;
  bool _refreshInProgress = false;
  DateTime? _lastOutageRefresh;
  NetWatcherConfig config = const NetWatcherConfig();
  NetworkSnapshot snapshot = const NetworkSnapshot();
  List<OutageRecord> outages = const [];
  int outageRangeDays = 30;
  ReportResult? lastReport;
  bool loading = true;
  bool outagesLoading = false;
  bool reportBusy = false;
  String? reportNotice;
  String? error;
  bool _shuttingDown = false;

  static Future<AppState> create({
    CoreService? service,
    bool pollSnapshots = true,
    bool manageWindowsStartup = true,
    Future<void> Function(NetworkEvent)? onOutageEvent,
  }) async {
    CoreService? resolvedService = service;
    String? startupFailure;
    if (resolvedService == null) {
      try {
        resolvedService = await ProcessCoreService.tryCreate();
        if (resolvedService == null) {
          startupFailure = 'netwatcher_core.exe was not found. '
              'Monitoring is unavailable; reinstall NetWatcher.';
        }
      } catch (exception) {
        startupFailure = 'netwatcher_core.exe could not start: $exception. '
            'Monitoring is unavailable.';
      }
    }

    final state = AppState._(
      resolvedService,
      pollSnapshots: pollSnapshots,
      manageWindowsStartup: manageWindowsStartup,
      onOutageEvent: onOutageEvent,
      startupFailure: startupFailure,
    );
    if (startupFailure != null) {
      state.error = startupFailure;
      state.loading = false;
      return state;
    }
    await state._initialise();
    return state;
  }

  Future<void> _initialise() async {
    try {
      await _service.initialise();
      config = await _service.loadSettings();

      String? startupError;
      if (_manageWindowsStartup) {
        try {
          await WindowsStartup.sync(config.startWithWindows);
        } catch (exception) {
          startupError = exception.toString();
        }
      }

      snapshot = await _service.snapshot();
      if (config.startMonitoringAutomatically && !snapshot.monitoring) {
        final previousEvents = snapshot.recentEvents;
        snapshot = await _service.startMonitoring();
        await _notifyNewEvents(previousEvents);
      }
      outages = await _service.getOutages(outageRangeDays);
      _lastOutageRefresh = DateTime.now();
      _syncPolling();
      error = startupError;
    } catch (exception) {
      error = exception.toString();
    } finally {
      loading = false;
      notifyListeners();
    }
  }

  Future<void> refreshSnapshot() async {
    if (_shuttingDown || _refreshInProgress) return;
    _refreshInProgress = true;
    try {
      final previousEvents = snapshot.recentEvents;
      snapshot = await _service.snapshot();
      await _notifyNewEvents(previousEvents);
      if (DateTime.now().difference(_lastOutageRefresh ?? DateTime(0)) >=
          const Duration(seconds: 30)) {
        _lastOutageRefresh = DateTime.now();
        await _refreshOutagesSilently();
      }
      error = null;
      notifyListeners();
    } catch (exception) {
      error = exception.toString();
      notifyListeners();
    } finally {
      _refreshInProgress = false;
      if (!snapshot.monitoring && _pollTimer != null) _syncPolling();
    }
  }

  void _syncPolling() {
    _pollTimer?.cancel();
    _pollTimer = null;
    if (!_pollSnapshots || _shuttingDown || !snapshot.monitoring) return;
    final intervalMs =
        (config.intervalSeconds * 1000).round().clamp(1000, 60000);
    _pollTimer = Timer.periodic(Duration(milliseconds: intervalMs), (_) {
      if (!_refreshInProgress) unawaited(refreshSnapshot());
    });
  }

  Future<void> _notifyNewEvents(List<NetworkEvent> previousEvents) async {
    if (!config.showOutageNotifications || _onOutageEvent == null) return;
    final known = previousEvents
        .map((event) => '${event.time}|${event.category}|${event.message}')
        .toSet();
    for (final event in snapshot.recentEvents.reversed) {
      if (event.category != 'outage' && event.category != 'recovery') continue;
      final key = '${event.time}|${event.category}|${event.message}';
      if (known.contains(key)) continue;
      try {
        await _onOutageEvent(event);
      } catch (_) {
        // A notification failure must not interrupt monitoring.
      }
    }
  }

  Future<void> refreshOutages([int? days]) async {
    if (_shuttingDown) return;
    outageRangeDays = days ?? outageRangeDays;
    outagesLoading = true;
    error = null;
    notifyListeners();
    try {
      outages = await _service.getOutages(outageRangeDays);
    } catch (exception) {
      error = exception.toString();
    } finally {
      outagesLoading = false;
      notifyListeners();
    }
  }

  Future<void> _refreshOutagesSilently() async {
    if (_shuttingDown) return;
    try {
      outages = await _service.getOutages(outageRangeDays);
    } catch (_) {
      return;
    }
  }

  Future<bool> clearOutageHistory() async {
    if (_shuttingDown || outagesLoading) return false;
    outagesLoading = true;
    error = null;
    notifyListeners();
    try {
      outages = await _service.clearOutageHistory(outageRangeDays);
      final previousEvents = snapshot.recentEvents;
      snapshot = await _service.snapshot();
      await _notifyNewEvents(previousEvents);
      return true;
    } catch (exception) {
      error = exception.toString();
      return false;
    } finally {
      outagesLoading = false;
      notifyListeners();
    }
  }

  Future<void> toggleMonitoring() async {
    try {
      final previousEvents = snapshot.recentEvents;
      snapshot = snapshot.monitoring
          ? await _service.stopMonitoring()
          : await _service.startMonitoring();
      await _notifyNewEvents(previousEvents);
      _syncPolling();
      error = null;
    } catch (exception) {
      error = exception.toString();
    }
    notifyListeners();
  }

  Future<void> saveConfig(NetWatcherConfig value) async {
    try {
      final previousEvents = snapshot.recentEvents;
      config = await _service.saveSettings(value);
      snapshot = await _service.snapshot();
      await _notifyNewEvents(previousEvents);
      _syncPolling();
      if (_manageWindowsStartup) {
        await WindowsStartup.sync(config.startWithWindows);
      }
      error = null;
    } catch (exception) {
      error = exception.toString();
    }
    notifyListeners();
  }

  Future<void> setGraphRange(int minutes) async {
    if (minutes == config.graphRangeMinutes) return;
    await saveConfig(config.copyWith(graphRangeMinutes: minutes));
  }

  Future<void> addTarget(String raw) async {
    final trimmed = raw.trim();
    if (trimmed.isEmpty || config.customTargets.contains(trimmed)) return;
    await saveConfig(
      config.copyWith(customTargets: [...config.customTargets, trimmed]),
    );
  }

  Future<void> removeTarget(String raw) async {
    await saveConfig(
      config.copyWith(
        customTargets:
            config.customTargets.where((value) => value != raw).toList(),
      ),
    );
  }

  Future<void> _runReport(Future<ReportResult> Function() action) async {
    if (reportBusy) return;
    reportBusy = true;
    reportNotice = null;
    error = null;
    notifyListeners();
    try {
      lastReport = await action();
      reportNotice = lastReport!.message.isEmpty
          ? 'Report created successfully.'
          : lastReport!.message;
    } catch (exception) {
      error = exception.toString();
      reportNotice = null;
    } finally {
      reportBusy = false;
      notifyListeners();
    }
  }

  Future<void> generateHtmlReport(int hours) =>
      _runReport(() => _service.generateHtmlReport(hours));

  Future<void> generateEvidenceReport(int days) =>
      _runReport(() => _service.generateEvidenceReport(days));

  Future<void> exportDiagnostics(int hours) =>
      _runReport(() => _service.exportDiagnostics(hours));

  Future<void> openLastReport() async {
    final report = lastReport;
    if (report == null || report.path.isEmpty) return;
    try {
      await _service.openFile(report.path);
      error = null;
    } catch (exception) {
      error = exception.toString();
    }
    notifyListeners();
  }

  Future<void> openReportsFolder() async {
    try {
      await _service.openReportsFolder();
      error = null;
    } catch (exception) {
      error = exception.toString();
    }
    notifyListeners();
  }

  Future<void> openLogsFolder() async {
    try {
      await _service.openLogsFolder();
      error = null;
    } catch (exception) {
      error = exception.toString();
    }
    notifyListeners();
  }

  Future<void> shutdown() async {
    if (_shuttingDown) return;
    _shuttingDown = true;
    _pollTimer?.cancel();
    await _resolvedService?.dispose();
  }

  @override
  void dispose() {
    unawaited(shutdown());
    super.dispose();
  }
}
