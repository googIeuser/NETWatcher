import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:netwatcher/app.dart';
import 'package:netwatcher/app_state.dart';
import 'package:netwatcher/mock_core_service.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  testWidgets('desktop navigation is clickable and changes page smoothly',
      (tester) async {
    await tester.binding.setSurfaceSize(const Size(1366, 768));
    final state = await AppState.create(
      service: MockCoreService(),
      pollSnapshots: false,
      manageWindowsStartup: false,
    );
    await tester.pumpWidget(NetWatcherApp(state: state));
    await tester.pumpAndSettle();

    final destination = find.byKey(const ValueKey<String>('nav-1'));
    expect(destination, findsOneWidget);
    final inkWell = tester.widget<InkWell>(
      find.descendant(of: destination, matching: find.byType(InkWell)),
    );
    expect(inkWell.mouseCursor, SystemMouseCursors.click);

    await tester.tap(destination);
    await tester.pump(const Duration(milliseconds: 120));
    await tester.pumpAndSettle();
    expect(find.text('Target-by-target performance summary.'), findsOneWidget);
    expect(tester.takeException(), isNull);

    state.dispose();
  });

  testWidgets('theme changes use a non-zero smooth animation', (tester) async {
    await tester.binding.setSurfaceSize(const Size(1024, 768));
    final state = await AppState.create(
      service: MockCoreService(),
      pollSnapshots: false,
      manageWindowsStartup: false,
    );
    await tester.pumpWidget(NetWatcherApp(state: state));
    await tester.pumpAndSettle();

    final app = tester.widget<MaterialApp>(find.byType(MaterialApp));
    expect(app.themeAnimationDuration, const Duration(milliseconds: 360));
    expect(app.themeAnimationDuration, isNot(Duration.zero));

    state.dispose();
  });
}
