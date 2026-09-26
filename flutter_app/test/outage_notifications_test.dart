import 'package:flutter_test/flutter_test.dart';
import 'package:netwatcher/app_state.dart';
import 'support/mock_core_service.dart';
import 'package:netwatcher/models.dart';

class EventCoreService extends MockCoreService {
  List<NetworkEvent> events = [];

  @override
  Future<NetworkSnapshot> snapshot() async => NetworkSnapshot(
        recentEvents: List<NetworkEvent>.of(events),
      );
}

class StartupOutageCoreService extends EventCoreService {
  @override
  Future<NetWatcherConfig> loadSettings() async =>
      const NetWatcherConfig(startMonitoringAutomatically: true);

  @override
  Future<NetworkSnapshot> startMonitoring() async {
    events = [
      const NetworkEvent(
        time: '2026-09-26T12:00:00Z',
        level: 'warning',
        category: 'outage',
        message: 'Connection lost on startup.',
      ),
    ];
    return NetworkSnapshot(recentEvents: events);
  }
}

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  test('notifies for an outage found during automatic startup', () async {
    final notified = <String>[];
    final state = await AppState.create(
      service: StartupOutageCoreService(),
      pollSnapshots: false,
      manageWindowsStartup: false,
      onOutageEvent: (event) async => notified.add(event.category),
    );
    addTearDown(state.dispose);

    expect(notified, ['outage']);
  });

  test('saving settings does not swallow a concurrent outage event', () async {
    final service = EventCoreService();
    final notified = <String>[];
    final state = await AppState.create(
      service: service,
      pollSnapshots: false,
      manageWindowsStartup: false,
      onOutageEvent: (event) async => notified.add(event.category),
    );
    addTearDown(state.dispose);

    service.events = [
      const NetworkEvent(
        time: '2026-09-26T12:03:00Z',
        level: 'warning',
        category: 'outage',
        message: 'Connection lost while saving settings.',
      ),
    ];
    await state.saveConfig(state.config.copyWith(timeoutMs: 1000));

    expect(notified, ['outage']);
  });

  test('notifies for new outages and recoveries only when enabled', () async {
    final service = EventCoreService();
    service.events = [
      const NetworkEvent(
        time: '2026-09-26T11:00:00Z',
        level: 'warning',
        category: 'outage',
        message: 'Previous outage.',
      ),
    ];
    final notified = <String>[];
    final state = await AppState.create(
      service: service,
      pollSnapshots: false,
      manageWindowsStartup: false,
      onOutageEvent: (event) async => notified.add(event.category),
    );
    addTearDown(state.dispose);
    expect(notified, isEmpty);

    service.events = [
      const NetworkEvent(
        time: '2026-09-26T12:00:00Z',
        level: 'warning',
        category: 'outage',
        message: 'Connection lost.',
      ),
      ...service.events,
    ];
    await state.refreshSnapshot();
    await state.refreshSnapshot();
    expect(notified, ['outage']);

    await state.saveConfig(
      state.config.copyWith(showOutageNotifications: false),
    );
    service.events = [
      const NetworkEvent(
        time: '2026-09-26T12:01:00Z',
        level: 'success',
        category: 'recovery',
        message: 'Connection recovered.',
      ),
      ...service.events,
    ];
    await state.refreshSnapshot();
    expect(notified, ['outage']);

    await state.saveConfig(
      state.config.copyWith(showOutageNotifications: true),
    );
    await state.refreshSnapshot();
    expect(notified, ['outage']);

    service.events = [
      const NetworkEvent(
        time: '2026-09-26T12:02:00Z',
        level: 'warning',
        category: 'outage',
        message: 'Connection lost again.',
      ),
      ...service.events,
    ];
    await state.refreshSnapshot();
    expect(notified, ['outage', 'outage']);
  });
}
