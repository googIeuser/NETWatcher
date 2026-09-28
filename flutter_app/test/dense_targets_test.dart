import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:netwatcher/app.dart';
import 'package:netwatcher/app_state.dart';
import 'package:netwatcher/models.dart';
import 'package:netwatcher/widgets.dart';

import 'support/mock_core_service.dart';

TargetStatus makeTarget(int index, {List<LatencySample> history = const []}) =>
    TargetStatus(
      target: TargetInfo(
        id: 'target-$index',
        name: 'Target $index',
        host: '10.0.0.$index',
        kind: 'internet',
        mode: 'ping',
      ),
      state: 'online',
      latency: 20 + index.toDouble(),
      packetLoss: 0,
      jitter: 1,
      history: history,
    );

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  test('a short burst of samples uses a readable chart viewport', () {
    final now = DateTime(2026, 9, 28, 19, 0, 0);
    final target = makeTarget(1, history: [
      LatencySample(
          time: now.subtract(const Duration(seconds: 4)),
          latency: 21,
          success: true),
      LatencySample(
          time: now.subtract(const Duration(seconds: 2)),
          latency: 25,
          success: true),
      LatencySample(time: now, latency: 23, success: true),
    ]);

    final viewport = LatencyViewport.forTargets(
      [target],
      rangeMinutes: 5,
      now: now,
    );

    expect(viewport, isNotNull);
    expect(viewport!.end.difference(viewport.start),
        lessThan(const Duration(seconds: 20)));
    expect(viewport.start.isBefore(target.history.first.time), isTrue);
    expect(viewport.end.isAfter(target.history.last.time), isTrue);
  });

  testWidgets('many targets remain manageable in the dashboard and target page',
      (tester) async {
    await tester.binding.setSurfaceSize(const Size(800, 600));
    final service = MockCoreService();
    await service.saveSettings(NetWatcherConfig(
      customTargets: List.generate(80, (index) => '10.20.0.${index + 1}'),
    ));
    final state = await AppState.create(
      service: service,
      pollSnapshots: false,
      manageWindowsStartup: false,
    );
    addTearDown(() async {
      state.dispose();
      await tester.binding.setSurfaceSize(null);
    });

    await tester.pumpWidget(NetWatcherApp(state: state));
    await tester.pumpAndSettle();
    expect(find.text('83 TARGETS'), findsOneWidget);
    for (var i = 0;
        i < 12 && find.text('View all 83 targets').evaluate().isEmpty;
        i++) {
      await tester.drag(find.byType(ListView).first, const Offset(0, -450));
      await tester.pumpAndSettle();
    }
    expect(find.text('View all 83 targets'), findsOneWidget);
    await tester.ensureVisible(find.text('View all 83 targets'));
    await tester.pumpAndSettle();
    expect(find.text('View all 83 targets'), findsOneWidget);
    expect(tester.takeException(), isNull);

    await tester.tap(find.text('View all 83 targets'));
    await tester.pumpAndSettle();
    for (var i = 0; i < 40 && find.text('10.20.0.80').evaluate().isEmpty; i++) {
      await tester.drag(find.byType(ListView).first, const Offset(0, -450));
      await tester.pumpAndSettle();
    }
    expect(find.text('10.20.0.80'), findsWidgets);
    await tester.ensureVisible(find.text('10.20.0.80'));
    await tester.pumpAndSettle();
    expect(find.text('10.20.0.80'), findsWidgets);
    expect(tester.takeException(), isNull);
  });

  testWidgets('chart legend stays bounded with many active series',
      (tester) async {
    await tester.binding.setSurfaceSize(const Size(640, 560));
    addTearDown(() => tester.binding.setSurfaceSize(null));
    final now = DateTime.now();
    final targets = List.generate(
      60,
      (index) => makeTarget(index, history: [
        LatencySample(
            time: now.subtract(const Duration(seconds: 2)),
            latency: 20 + index.toDouble(),
            success: true),
        LatencySample(time: now, latency: 21 + index.toDouble(), success: true),
      ]),
    );

    await tester.pumpWidget(MaterialApp(
      home: Scaffold(
          body: Center(
              child:
                  SizedBox(width: 600, child: LatencyChart(targets: targets)))),
    ));
    await tester.pumpAndSettle();

    expect(find.text('56 more targets'), findsOneWidget);
    expect(find.textContaining('Target 59:'), findsNothing);
    await tester.tap(find.byKey(const ValueKey<String>('chart-target-filter')));
    await tester.pumpAndSettle();
    await tester.scrollUntilVisible(
      find.text('Target 59'),
      350,
      scrollable: find.byType(Scrollable).last,
      maxScrolls: 20,
    );
    await tester.tap(find.text('Target 59'));
    await tester.pumpAndSettle();
    expect(find.textContaining('Target 59:'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('long target names do not overflow a narrow chart',
      (tester) async {
    await tester.binding.setSurfaceSize(const Size(340, 420));
    addTearDown(() => tester.binding.setSurfaceSize(null));
    final now = DateTime.now();
    final target = TargetStatus(
      target: const TargetInfo(
        id: 'long',
        name:
            'A very long custom endpoint name that should fit safely in the chart legend without spilling out of the card',
        host: 'example.com',
        kind: 'internet',
        mode: 'ping',
      ),
      state: 'online',
      latency: 32,
      packetLoss: 0,
      jitter: 1,
      history: [
        LatencySample(
            time: now.subtract(const Duration(seconds: 2)),
            latency: 30,
            success: true),
        LatencySample(time: now, latency: 32, success: true),
      ],
    );
    await tester.pumpWidget(MaterialApp(
      home: Scaffold(
          body: SizedBox(width: 320, child: LatencyChart(targets: [target]))),
    ));
    await tester.pumpAndSettle();
    expect(tester.takeException(), isNull);
  });
}
