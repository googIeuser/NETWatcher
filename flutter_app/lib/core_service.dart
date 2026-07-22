import 'models.dart';

abstract interface class CoreService {
  Future<void> initialise();
  Future<NetWatcherConfig> loadSettings();
  Future<NetWatcherConfig> saveSettings(NetWatcherConfig config);
  Future<NetworkSnapshot> startMonitoring();
  Future<NetworkSnapshot> stopMonitoring();
  Future<NetworkSnapshot> snapshot();
  Future<List<OutageRecord>> getOutages(int days);
  Future<ReportResult> generateHtmlReport(int hours);
  Future<ReportResult> generateEvidenceReport(int days);
  Future<ReportResult> exportDiagnostics(int hours);
  Future<void> openFile(String path);
  Future<void> openReportsFolder();
  Future<void> openLogsFolder();
  Future<void> dispose();
}
