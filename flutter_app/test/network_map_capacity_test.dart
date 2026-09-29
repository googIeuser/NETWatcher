import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:netwatcher/models.dart';
import 'package:netwatcher/network_map.dart';

TargetStatus targetAt(int index) => TargetStatus(
      target: TargetInfo(
        id: 'target-$index',
        name: 'Target $index',
        host: '10.0.0.$index',
        kind: 'internet',
        mode: 'ping',
      ),
      state: 'online',
      latency: 20,
      packetLoss: 0,
      jitter: 0,
    );

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  for (final width in [1120.0, 600.0]) {
    testWidgets('network map shows ten targets at width $width',
        (tester) async {
      await tester.binding.setSurfaceSize(Size(width, 850));
      addTearDown(() => tester.binding.setSurfaceSize(null));
      await tester.pumpWidget(MaterialApp(
        home: Scaffold(
          body: SingleChildScrollView(
            child: NetworkMap(
                snapshot: NetworkSnapshot(
              monitoring: true,
              samples: 1,
              targets: List.generate(11, targetAt),
            )),
          ),
        ),
      ));
      await tester.pumpAndSettle();

      for (var i = 0; i < 10; i++) {
        expect(find.text('Target $i'), findsOneWidget);
      }
      expect(find.text('Target 10'), findsNothing);
      expect(find.text('+1 more targets below'), findsOneWidget);
      if (width >= 700) {
        expect(
          tester.getTopLeft(find.text('+1 more targets below')).dy,
          greaterThan(tester.getBottomLeft(find.text('Target 9')).dy),
        );
      }
      expect(tester.takeException(), isNull);
    });
  }

  for (final count in [5, 10]) {
    testWidgets('map rows do not overlap with $count paused targets',
        (tester) async {
      await tester.binding.setSurfaceSize(const Size(1120, 850));
      addTearDown(() => tester.binding.setSurfaceSize(null));
      await tester.pumpWidget(MaterialApp(
        home: Scaffold(
          body: SingleChildScrollView(
            child: NetworkMap(
              snapshot: NetworkSnapshot(
                targets: List.generate(count, targetAt),
              ),
            ),
          ),
        ),
      ));
      await tester.pumpAndSettle();

      for (var i = 2; i < count; i += 2) {
        final gap = tester.getTopLeft(find.text('Target $i')).dy -
            tester.getTopLeft(find.text('Target ${i - 2}')).dy;
        expect(gap, greaterThanOrEqualTo(96));
      }
      expect(tester.takeException(), isNull);
    });
  }
}
