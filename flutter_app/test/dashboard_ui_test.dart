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

  testWidgets('network map reflects live endpoints and monitoring state',
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
    expect(find.text('NETWORK MAP'), findsOneWidget);
    expect(find.text('Monitoring paused'), findsOneWidget);
    expect(find.text('Default Gateway'), findsWidgets);

    await state.toggleMonitoring();
    await tester.pumpAndSettle();
    expect(find.text('Online'), findsWidgets);
    expect(find.text('Default Gateway'), findsWidgets);
    expect(find.text('Cloudflare'), findsWidgets);
    expect(find.text('Google'), findsWidgets);
    expect(tester.takeException(), isNull);
  });

  testWidgets('sidebar has no hover or selected background but keeps labels',
      (tester) async {
    await tester.binding.setSurfaceSize(const Size(1024, 768));
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

    final selected = find.byKey(const ValueKey<String>('nav-0'));
    final tile = tester.widget<Container>(
      find.descendant(of: selected, matching: find.byType(Container)).first,
    );
    expect((tile.decoration as BoxDecoration?)?.color, Colors.transparent);
    final ink = tester.widget<InkWell>(
      find.descendant(of: selected, matching: find.byType(InkWell)).first,
    );
    expect(
        ink.overlayColor?.resolve({WidgetState.hovered}), Colors.transparent);
    expect(find.descendant(of: selected, matching: find.byType(Tooltip)),
        findsOneWidget);
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
