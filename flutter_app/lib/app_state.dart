import 'dart:async';

import 'package:flutter/foundation.dart';

import 'core_service.dart';
import 'mock_core_service.dart';
import 'models.dart';
import 'process_core_service.dart';
import 'windows_startup.dart';

class AppState extends ChangeNotifier {
  AppState._(
    this._service, {
    required bool pollSnapshots,
    required bool manageWindowsStartup,
  })  : _pollSnapshots = pollSnapshots,
        _manageWindowsStartup = manageWindowsStartup;

  final CoreService _service;
  final bool _pollSnapshots;
  final bool _manageWindowsStartup;
  Timer? _pollTimer;
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
  int _outagePollTicks = 0;
  bool _shuttingDown = false;

  static Future<AppState> create({
    CoreService? service,
    bool pollSnapshots = true,
    bool manageWindowsStartup = true,
  }) async {
    CoreService resolvedService;
    if (service != null) {
      resolvedService = service;
    } else {
      try {
        resolvedService =
            await ProcessCoreService.tryCreate() ?? MockCoreService();
      } catch (_) {
        resolvedService = MockCoreService();
      }
    }

    final state = AppState._(
      resolvedService,
      pollSnapshots: pollSnapshots,
      manageWindowsStartup: manageWindowsStartup,
    );
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
        snapshot = await _service.startMonitoring();
      }
      outages = await _service.getOutages(outageRangeDays);
      if (_pollSnapshots) {
        _pollTimer = Timer.periodic(
          const Duration(seconds: 1),
          (_) => refreshSnapshot(),
        );
      }
      error = startupError;
    } catch (exception) {
      error = exception.toString();
    } finally {
      loading = false;
      notifyListeners();
    }
  }

  Future<void> refreshSnapshot() async {
    if (_shuttingDown) return;
    try {
      snapshot = await _service.snapshot();
      _outagePollTicks++;
      if (_outagePollTicks >= 5) {
        _outagePollTicks = 0;
        await _refreshOutagesSilently();
      }
      error = null;
      notifyListeners();
    } catch (exception) {
      error = exception.toString();
      notifyListeners();
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
      snapshot = await _service.snapshot();
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
      snapshot = snapshot.monitoring
          ? await _service.stopMonitoring()
          : await _service.startMonitoring();
      error = null;
    } catch (exception) {
      error = exception.toString();
    }
    notifyListeners();
  }

  Future<void> saveConfig(NetWatcherConfig value) async {
    try {
      config = await _service.saveSettings(value);
      snapshot = await _service.snapshot();
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
    await _service.dispose();
  }

  @override
  void dispose() {
    unawaited(shutdown());
    super.dispose();
  }
}
