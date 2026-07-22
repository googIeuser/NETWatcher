import 'dart:async';

import 'package:flutter/foundation.dart';

import 'core_service.dart';
import 'mock_core_service.dart';
import 'models.dart';
import 'process_core_service.dart';

class AppState extends ChangeNotifier {
  AppState._(this._service);

  final CoreService _service;
  Timer? _pollTimer;
  NetWatcherConfig config = const NetWatcherConfig();
  NetworkSnapshot snapshot = const NetworkSnapshot();
  ReportResult? lastReport;
  bool loading = true;
  bool reportBusy = false;
  String? reportNotice;
  String? error;

  static Future<AppState> create() async {
    CoreService service;
    try {
      service = await ProcessCoreService.tryCreate() ?? MockCoreService();
    } catch (_) {
      service = MockCoreService();
    }
    final state = AppState._(service);
    await state._initialise();
    return state;
  }

  Future<void> _initialise() async {
    try {
      await _service.initialise();
      config = await _service.loadSettings();
      snapshot = await _service.snapshot();
      _pollTimer = Timer.periodic(
        const Duration(seconds: 1),
        (_) => refreshSnapshot(),
      );
    } catch (exception) {
      error = exception.toString();
    } finally {
      loading = false;
      notifyListeners();
    }
  }

  Future<void> refreshSnapshot() async {
    try {
      snapshot = await _service.snapshot();
      error = null;
      notifyListeners();
    } catch (exception) {
      error = exception.toString();
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
      error = null;
    } catch (exception) {
      error = exception.toString();
    }
    notifyListeners();
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

  @override
  void dispose() {
    _pollTimer?.cancel();
    _service.dispose();
    super.dispose();
  }
}
