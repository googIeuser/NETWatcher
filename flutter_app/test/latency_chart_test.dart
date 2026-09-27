import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:netwatcher/models.dart';
import 'package:netwatcher/widgets.dart';

void main() {
  testWidgets('offline target legend does not present zero latency',
      (tester) async {
    final failed = TargetStatus(
      target: const TargetInfo(
        id: 'google',
        name: 'Google',
        host: '8.8.8.8',
        kind: 'internet',
        mode: 'ping',
      ),
      state: 'offline',
      latency: 0,
      packetLoss: 100,
      jitter: 0,
      history: [
        LatencySample(time: DateTime.now(), latency: 0, success: false),
      ],
    );
    await tester.pumpWidget(MaterialApp(
      home: Scaffold(body: LatencyChart(targets: [failed])),
    ));
    await tester.pumpAndSettle();

    expect(find.text('Google: No response'), findsOneWidget);
    expect(find.text('Google: 0.0 ms'), findsNothing);
    expect(tester.takeException(), isNull);
  });
}
