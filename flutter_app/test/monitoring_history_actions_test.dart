import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:netwatcher/app.dart';
import 'package:netwatcher/app_state.dart';
import 'package:netwatcher/mock_core_service.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  test('stopping monitoring records a recent event', () async {
    final state = await AppState.create(
      service: MockCoreService(),
      pollSnapshots: false,
      manageWindowsStartup: false,
    );
    addTearDown(state.dispose);

    await state.toggleMonitoring();
    expect(state.snapshot.monitoring, isTrue);
    expect(state.snapshot.recentEvents.first.message, 'Monitoring started.');

    await state.toggleMonitoring();
    expect(state.snapshot.monitoring, isFalse);
    expect(state.snapshot.recentEvents.first.message, 'Monitoring stopped.');
  });

  testWidgets('outage history can be deleted after confirmation',
      (tester) async {
    await tester.binding.setSurfaceSize(const Size(1366, 900));
    final state = await AppState.create(
      service: MockCoreService(),
      pollSnapshots: false,
      manageWindowsStartup: false,
    );

    addTearDown(() async {
      state.dispose();
      await tester.binding.setSurfaceSize(null);
    });

    await tester.pumpWidget(NetWatcherApp(state: state));
    await tester.pumpAndSettle();

    await tester.tap(find.byKey(const ValueKey<String>('nav-2')));
    await tester.pumpAndSettle();

    expect(state.outages, isNotEmpty);
    expect(
      find.byKey(const ValueKey<String>('delete-outage-history')),
      findsOneWidget,
    );

    await tester.tap(
      find.byKey(const ValueKey<String>('delete-outage-history')),
    );
    await tester.pumpAndSettle();

    expect(find.text('Delete outage history?'), findsOneWidget);
    await tester.tap(find.widgetWithText(FilledButton, 'Delete'));
    await tester.pumpAndSettle();

    expect(state.outages, isEmpty);
    expect(find.text('No confirmed outages in this range.'), findsOneWidget);
    expect(find.text('Outage history deleted.'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });
}
