import 'package:flutter_test/flutter_test.dart';
import 'package:netwatcher/app_state.dart';
import 'package:netwatcher/models.dart';

import 'support/mock_core_service.dart';

class _CountingCoreService extends MockCoreService {
  int snapshotCalls = 0;

  @override
  Future<NetworkSnapshot> snapshot() {
    snapshotCalls++;
    return super.snapshot();
  }
}

void main() {
  testWidgets('snapshot polling runs only while monitoring', (tester) async {
    final service = _CountingCoreService();
    final state = await AppState.create(
      service: service,
      manageWindowsStartup: false,
    );
    addTearDown(state.dispose);

    final idleCalls = service.snapshotCalls;
    await tester.pump(const Duration(seconds: 6));
    expect(service.snapshotCalls, idleCalls);

    await state.toggleMonitoring();
    final startCalls = service.snapshotCalls;
    await tester.pump(const Duration(seconds: 2));
    expect(service.snapshotCalls, greaterThan(startCalls));

    await state.toggleMonitoring();
    final stoppedCalls = service.snapshotCalls;
    await tester.pump(const Duration(seconds: 6));
    expect(service.snapshotCalls, stoppedCalls);
  });
}
