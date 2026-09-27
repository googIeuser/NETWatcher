import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:netwatcher/app.dart';
import 'package:netwatcher/app_state.dart';
import 'support/mock_core_service.dart';
import 'package:netwatcher/widgets.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  testWidgets('dashboard does not present zeroes as measurements while idle',
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

    expect(find.byKey(const ValueKey<String>('dashboard-empty-state')),
        findsOneWidget);
    expect(find.text('0.0'), findsNothing);
    expect(find.text('QUALITY'), findsNothing);
    expect(tester.takeException(), isNull);
  });

  testWidgets('dashboard exposes the styled latency history range selector',
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

    expect(find.text('Latency history'), findsOneWidget);
    expect(find.text('History range'), findsOneWidget);
    expect(find.byType(DropdownButtonFormField<int>), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('four dashboard metrics share one row on desktop',
      (tester) async {
    await tester.binding.setSurfaceSize(const Size(1260, 760));
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

    final first = tester.getTopLeft(find.text('01  /  AVG LATENCY'));
    final last = tester.getTopLeft(find.text('04  /  SAMPLES'));
    expect(last.dy, first.dy);
    expect(last.dx, greaterThan(first.dx));
    expect(tester.takeException(), isNull);
  });

  testWidgets('statistics list starts without an empty top divider strip',
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

    await tester.tap(find.byKey(const ValueKey<String>('nav-1')));
    await tester.pumpAndSettle();

    final firstTarget = find.byType(TargetCard).first;
    final outerContainer = find
        .descendant(
          of: firstTarget,
          matching: find.byType(AnimatedContainer),
        )
        .first;
    final widget = tester.widget<AnimatedContainer>(outerContainer);
    final decoration = widget.decoration as BoxDecoration?;

    expect(decoration?.border, isNull);
    expect(tester.takeException(), isNull);
  });
}
