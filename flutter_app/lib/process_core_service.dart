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
      NetworkSnapshot.fromJson(_data(await _request('start')));

  @override
  Future<NetworkSnapshot> stopMonitoring() async =>
      NetworkSnapshot.fromJson(_data(await _request('stop')));

  @override
  Future<NetworkSnapshot> snapshot() async =>
      NetworkSnapshot.fromJson(_data(await _request('snapshot')));

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
