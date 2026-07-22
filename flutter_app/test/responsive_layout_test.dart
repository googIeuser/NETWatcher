import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:netwatcher/app.dart';
import 'package:netwatcher/app_state.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  Future<void> pumpAt(
    WidgetTester tester,
    Size size,
  ) async {
    await tester.binding.setSurfaceSize(size);
    final state = await AppState.create();
    await tester.pumpWidget(NetWatcherApp(state: state));
    await tester.pumpAndSettle();
    expect(tester.takeException(), isNull);
    state.dispose();
  }

  testWidgets('dashboard has no overflow at common Windows sizes',
      (tester) async {
    for (final size in const [
      Size(800, 600),
      Size(1024, 768),
      Size(1280, 720),
      Size(1366, 768),
      Size(1920, 1080),
    ]) {
      await pumpAt(tester, size);
    }
  });
}
