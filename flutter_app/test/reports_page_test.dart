import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:netwatcher/app.dart';
import 'package:netwatcher/app_state.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  testWidgets('reports page exposes working report actions', (tester) async {
    await tester.binding.setSurfaceSize(const Size(1366, 900));
    final state = await AppState.create();
    await tester.pumpWidget(NetWatcherApp(state: state));
    await tester.pumpAndSettle();

    await tester.tap(find.byKey(const ValueKey<String>('nav-3')));
    await tester.pumpAndSettle();

    expect(find.text('ISP Evidence Report'), findsOneWidget);
    expect(find.text('Create and open evidence report'), findsOneWidget);
    expect(find.text('Coming in core integration'), findsNothing);

    await tester.tap(find.text('Create and open evidence report'));
    await tester.pumpAndSettle();
    expect(find.text('Open file'), findsOneWidget);
    expect(tester.takeException(), isNull);

    state.dispose();
  });
}
