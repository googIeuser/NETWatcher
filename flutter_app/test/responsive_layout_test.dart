import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:netwatcher/app.dart';
import 'package:netwatcher/app_state.dart';
import 'support/mock_core_service.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  Future<void> pumpAt(
    WidgetTester tester,
    Size size,
  ) async {
    await tester.binding.setSurfaceSize(size);
    final state = await AppState.create(
      service: MockCoreService(),
      pollSnapshots: false,
      manageWindowsStartup: false,
    );
    await tester.pumpWidget(NetWatcherApp(state: state));
    await tester.pumpAndSettle();
    expect(tester.takeException(), isNull);
    state.dispose();
  }

  testWidgets('dashboard has no overflow at common Windows sizes',
      (tester) async {
    for (final size in const [
      Size(800, 600),
      Size(1024, 768),
      Size(1280, 720),
      Size(1366, 768),
      Size(1920, 1080),
    ]) {
      await pumpAt(tester, size);
    }
  });

  testWidgets('targets form has no overflow at compact and desktop sizes',
      (tester) async {
    for (final size in const [Size(800, 600), Size(1246, 752)]) {
      await tester.binding.setSurfaceSize(size);
      final state = await AppState.create(
        service: MockCoreService(),
        pollSnapshots: false,
        manageWindowsStartup: false,
      );
      await tester.pumpWidget(NetWatcherApp(state: state));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const ValueKey<String>('nav-4')));
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
      state.dispose();
    }
    await tester.binding.setSurfaceSize(null);
  });
}
