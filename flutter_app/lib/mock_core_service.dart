import 'dart:io';
import 'dart:math';

import 'core_service.dart';
import 'models.dart';

class MockCoreService implements CoreService {
  final Random _random = Random(7);
  NetWatcherConfig _config = const NetWatcherConfig();
  bool _monitoring = false;
  int _samples = 0;

  @override
  Future<void> initialise() async {}

  @override
  Future<NetWatcherConfig> loadSettings() async => _config;

  @override
  Future<NetWatcherConfig> saveSettings(NetWatcherConfig config) async {
    _config = config;
    return config;
  }

  @override
  Future<NetworkSnapshot> startMonitoring() async {
    _monitoring = true;
    return snapshot();
  }

  @override
  Future<NetworkSnapshot> stopMonitoring() async {
    _monitoring = false;
    return snapshot();
  }

  @override
  Future<NetworkSnapshot> snapshot() async {
    if (_monitoring) _samples += 3;
    final statuses = <TargetStatus>[
      _target('gateway', 'Default Gateway', '192.168.1.1', 'ping', 7.8),
      _target('cloudflare', 'Cloudflare', '1.1.1.1', 'ping', 15.2),
      _target('google', 'Google', '8.8.8.8', 'ping', 39.5),
      ..._config.customTargets.map(
        (raw) => _target(
          raw,
          raw.startsWith('http') ? 'HTTPS: custom target' : 'Custom target',
          raw,
          raw.startsWith('tcp://')
              ? 'tcp'
              : raw.startsWith('http')
                  ? 'https'
                  : 'ping',
          68,
        ),
      ),
    ];
    final average = statuses.isEmpty
        ? 0.0
        : statuses.map((e) => e.latency).reduce((a, b) => a + b) /
            statuses.length;
    return NetworkSnapshot(
      monitoring: _monitoring,
      connectionState: _monitoring ? 'online' : 'waiting',
      connectionLabel: _monitoring ? 'Online' : 'Monitoring stopped',
      qualityScore: _monitoring ? 96 : 0,
      averageLatency: average,
      packetLoss: 0,
      jitter: _monitoring ? 2.4 : 0,
      samples: _samples,
      targets: statuses,
      recentEvents: [
        NetworkEvent(
          time: DateTime.now().toIso8601String(),
          level: 'success',
          category: 'monitor',
          message: _monitoring
              ? 'Monitoring is active.'
              : 'Monitoring is stopped.',
        ),
      ],
      updatedAt: DateTime.now().toIso8601String(),
    );
  }

  TargetStatus _target(
    String id,
    String name,
    String host,
    String mode,
    double baseline,
  ) {
    final latency = _monitoring ? baseline + _random.nextDouble() * 4 : 0.0;
    return TargetStatus(
      target: TargetInfo(
        id: id,
        name: name,
        host: host,
        kind: host.startsWith('192.') ? 'local' : 'internet',
        mode: mode,
        custom: id != 'gateway' && id != 'cloudflare' && id != 'google',
      ),
      state: _monitoring ? 'online' : 'waiting',
      latency: latency,
      packetLoss: 0,
      jitter: _monitoring ? _random.nextDouble() * 3 : 0,
    );
  }

  Future<ReportResult> _mockReport(String kind, String extension) async {
    final dir = await Directory.systemTemp.createTemp('netwatcher_reports_');
    final path = '${dir.path}${Platform.pathSeparator}netwatcher_$kind.$extension';
    final file = File(path);
    await file.writeAsString(
      extension == 'html'
          ? '<!doctype html><title>NetWatcher $kind</title><h1>NetWatcher $kind</h1>'
          : 'NetWatcher diagnostics preview',
    );
    return ReportResult(
      kind: kind,
      path: path,
      createdAt: DateTime.now().toIso8601String(),
      message: 'Preview report created.',
    );
  }

  @override
  Future<ReportResult> generateHtmlReport(int hours) =>
      _mockReport('html_report_${hours}h', 'html');

  @override
  Future<ReportResult> generateEvidenceReport(int days) =>
      _mockReport('evidence_${days}d', 'html');

  @override
  Future<ReportResult> exportDiagnostics(int hours) =>
      _mockReport('diagnostics_${hours}h', 'zip');

  @override
  Future<void> openFile(String path) async {}

  @override
  Future<void> openReportsFolder() async {}

  @override
  Future<void> openLogsFolder() async {}

  @override
  Future<void> dispose() async {}
}
