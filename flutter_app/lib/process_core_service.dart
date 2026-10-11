import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'core_service.dart';
import 'models.dart';

class ProcessCoreService implements CoreService {
  Process? _process;
  StreamSubscription<String>? _stdoutSubscription;
  StreamSubscription<String>? _stderrSubscription;
  final List<Completer<Map<String, dynamic>>> _waiting = [];
  final SnapshotHistoryCache _history = SnapshotHistoryCache();

  static Future<ProcessCoreService?> tryCreate() async {
    final path = _findCoreExecutable();
    if (path == null) return null;
    final service = ProcessCoreService();
    await service._start(path);
    return service;
  }

  static String? _findCoreExecutable() {
    final candidates = <String?>[
      Platform.environment['NETWATCHER_CORE_PATH'],
      '${File(Platform.resolvedExecutable).parent.path}${Platform.pathSeparator}netwatcher_core.exe',
      '${Directory.current.path}${Platform.pathSeparator}netwatcher_core.exe',
    ];
    for (final candidate in candidates) {
      if (candidate == null || candidate.trim().isEmpty) continue;
      if (File(candidate).existsSync()) return candidate;
    }
    return null;
  }

  Future<void> _start(String path) async {
    _process = await Process.start(path, const []);
    _stdoutSubscription = _process!.stdout
        .transform(utf8.decoder)
        .transform(const LineSplitter())
        .listen(_handleLine);
    _stderrSubscription = _process!.stderr
        .transform(utf8.decoder)
        .transform(const LineSplitter())
        .listen((_) {});
    final hello = await _request('hello');
    if (hello['ok'] != true) throw StateError('Rust core handshake failed');
  }

  void _handleLine(String line) {
    if (_waiting.isEmpty) return;
    final completer = _waiting.removeAt(0);
    try {
      completer.complete(jsonDecode(line) as Map<String, dynamic>);
    } catch (error, stackTrace) {
      completer.completeError(error, stackTrace);
    }
  }

  Future<Map<String, dynamic>> _request(
    String command, [
    Map<String, dynamic> payload = const {},
  ]) async {
    final process = _process;
    if (process == null) throw StateError('Rust core is not running');
    final completer = Completer<Map<String, dynamic>>();
    _waiting.add(completer);
    process.stdin.writeln(jsonEncode({'command': command, ...payload}));
    final response = await completer.future.timeout(const Duration(seconds: 90));
    if (response['ok'] != true) {
      throw StateError(response['error']?.toString() ?? 'Rust core error');
    }
    return response;
  }

  Map<String, dynamic> _data(Map<String, dynamic> result) =>
      result['data'] as Map<String, dynamic>? ?? {};

  @override
  Future<void> initialise() async {}

  @override
  Future<NetWatcherConfig> loadSettings() async =>
      NetWatcherConfig.fromJson(_data(await _request('load_settings')));

  @override
  Future<NetWatcherConfig> saveSettings(NetWatcherConfig config) async =>
      NetWatcherConfig.fromJson(
        _data(await _request('save_settings', {'config': config.toJson()})),
      );

  @override
  Future<NetworkSnapshot> startMonitoring() async =>
      _history.apply(_data(await _request('start')));

  @override
  Future<NetworkSnapshot> stopMonitoring() async =>
      _history.apply(_data(await _request('stop')));

  @override
  Future<NetworkSnapshot> snapshot() async => _history.apply(
        _data(await _request('snapshot', _history.requestFields())),
      );

  @override
  Future<List<OutageRecord>> getOutages(int days) async {
    final result = await _request('get_outages', {'days': days});
    final values = result['data'] as List<dynamic>? ?? const [];
    return values
        .whereType<Map<String, dynamic>>()
        .map(OutageRecord.fromJson)
        .toList(growable: false);
  }

  @override
  Future<List<OutageRecord>> clearOutageHistory(int days) async {
    final result = await _request('clear_outages', {'days': days});
    final values = result['data'] as List<dynamic>? ?? const [];
    return values
        .whereType<Map<String, dynamic>>()
        .map(OutageRecord.fromJson)
        .toList(growable: false);
  }

  @override
  Future<ReportResult> generateHtmlReport(int hours) async =>
      ReportResult.fromJson(
        _data(await _request('generate_html_report', {'hours': hours})),
      );

  @override
  Future<ReportResult> generateEvidenceReport(int days) async =>
      ReportResult.fromJson(
        _data(await _request('generate_evidence_report', {'days': days})),
      );

  @override
  Future<ReportResult> exportDiagnostics(int hours) async =>
      ReportResult.fromJson(
        _data(await _request('export_diagnostics', {'hours': hours})),
      );

  @override
  Future<void> openFile(String path) async {
    await _request('open_file', {'path': path});
  }

  @override
  Future<void> openReportsFolder() async {
    await _request('open_reports_folder');
  }

  @override
  Future<void> openLogsFolder() async {
    await _request('open_logs_folder');
  }

  @override
  Future<void> dispose() async {
    try {
      await _request('shutdown').timeout(const Duration(seconds: 3));
    } catch (_) {}
    await _process?.stdin.close();
    _process?.kill();
    await _stdoutSubscription?.cancel();
    await _stderrSubscription?.cancel();
    _process = null;
  }
}

/// Keeps chart history between snapshot requests, so the core only has to send
/// the chart buckets that are new since the previous request.
class SnapshotHistoryCache {
  int? _epoch;
  final Map<String, List<LatencySample>> _history = {};

  Map<String, dynamic> requestFields() {
    final epoch = _epoch;
    int? newest;
    for (final samples in _history.values) {
      if (samples.isEmpty) continue;
      final time = samples.last.time.millisecondsSinceEpoch;
      if (newest == null || time > newest) newest = time;
    }
    if (epoch == null || newest == null) return const {};
    return {'history_after': newest, 'history_epoch': epoch};
  }

  NetworkSnapshot apply(Map<String, dynamic> data) {
    final snapshot = NetworkSnapshot.fromJson(data);
    final epoch = (data['historyEpoch'] as num?)?.toInt();
    final isDelta =
        data['historyDelta'] == true && epoch != null && epoch == _epoch;
    final bucketMs =
        ((data['historyBucketSeconds'] as num?)?.toInt() ?? 1).clamp(1, 86400) *
            1000;
    final cutoffMs = (data['historyCutoffMs'] as num?)?.toInt() ?? 0;
    int bucket(LatencySample sample) =>
        sample.time.millisecondsSinceEpoch ~/ bucketMs;

    final merged = <String, List<LatencySample>>{};
    final targets = snapshot.targets.map((status) {
      var history = status.history;
      if (isDelta) {
        final cached = _history[status.target.id] ?? const <LatencySample>[];
        final firstBucket =
            history.isEmpty ? null : bucket(history.first);
        history = [
          for (final sample in cached)
            if (sample.time.millisecondsSinceEpoch >= cutoffMs &&
                (firstBucket == null || bucket(sample) < firstBucket))
              sample,
          ...history,
        ];
      }
      merged[status.target.id] = history;
      return isDelta ? status.copyWith(history: history) : status;
    }).toList(growable: false);

    _epoch = epoch;
    _history
      ..clear()
      ..addAll(merged);
    return isDelta ? snapshot.copyWith(targets: targets) : snapshot;
  }
}
