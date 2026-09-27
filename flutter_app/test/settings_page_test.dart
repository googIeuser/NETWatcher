import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:netwatcher/app.dart';
import 'package:netwatcher/app_state.dart';

import 'support/mock_core_service.dart';

void main() {
  testWidgets('settings save keeps the updated notification preference',
      (tester) async {
    await tester.binding.setSurfaceSize(const Size(1246, 752));
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
    await tester.tap(find.byKey(const ValueKey<String>('nav-5')));
    await tester.pumpAndSettle();
    expect(state.config.showOutageNotifications, isTrue);

    await tester.tap(find.widgetWithText(
        SwitchListTile, 'Outage and recovery notifications'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Save settings'));
    await tester.pumpAndSettle();

    expect(state.config.showOutageNotifications, isFalse);
    expect(tester.takeException(), isNull);
  });
}
