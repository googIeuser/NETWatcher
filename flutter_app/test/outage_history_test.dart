import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:netwatcher/models.dart';
import 'package:netwatcher/pages.dart';

void main() {
  testWidgets('outage incident card shows details without overflow',
      (tester) async {
    await tester.binding.setSurfaceSize(const Size(420, 720));
    await tester.pumpWidget(
      const MaterialApp(
        home: Scaffold(
          body: SingleChildScrollView(
            child: OutageIncidentCard(
              incident: OutageRecord(
                start: '2026-07-22T12:00:00Z',
                end: '2026-07-22T12:05:30Z',
                category: 'offline',
                details:
                    'Gateway responds but all monitored internet targets failed for several consecutive checks.',
                durationSeconds: 330,
                active: false,
              ),
            ),
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();
    expect(tester.takeException(), isNull);
    expect(find.text('Internet outage'), findsOneWidget);
    expect(find.text('RESOLVED'), findsOneWidget);
    expect(find.text('5m 30s'), findsOneWidget);
  });
}
