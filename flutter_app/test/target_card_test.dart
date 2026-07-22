import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:netwatcher/models.dart';
import 'package:netwatcher/widgets.dart';

void main() {
  testWidgets('long endpoint and status stay inside narrow target card',
      (tester) async {
    await tester.binding.setSurfaceSize(const Size(330, 700));
    await tester.pumpWidget(
      const MaterialApp(
        home: Scaffold(
          body: SizedBox(
            width: 300,
            child: TargetCard(
              status: TargetStatus(
                target: TargetInfo(
                  id: 'long',
                  name:
                      'Extremely long endpoint name that must wrap without leaving the card',
                  host:
                      'https://very-long-subdomain.example.com/a/very/long/path',
                  kind: 'internet',
                  mode: 'https',
                ),
                state: 'online',
                latency: 123.4,
                packetLoss: 0,
                jitter: 4.2,
              ),
            ),
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();
    expect(tester.takeException(), isNull);
    expect(find.text('ONLINE'), findsOneWidget);
  });
}
