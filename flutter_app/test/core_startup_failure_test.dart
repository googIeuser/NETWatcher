import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:netwatcher/app.dart';
import 'package:netwatcher/app_state.dart';
import 'package:netwatcher/desktop_tray.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  testWidgets('missing core shows an error and cannot start mock monitoring',
      (tester) async {
    final state = await AppState.create(
      pollSnapshots: false,
      manageWindowsStartup: false,
    );
    addTearDown(state.dispose);

    await tester.pumpWidget(NetWatcherApp(state: state));
    await tester.pumpAndSettle();

    expect(find.textContaining('netwatcher_core.exe'), findsOneWidget);
    expect(state.snapshot.monitoring, isFalse);

    await state.toggleMonitoring();
    expect(state.snapshot.monitoring, isFalse);
  });

  testWidgets('autostart shows the window when core startup failed',
      (tester) async {
    final calls = <String>[];
    const channel = MethodChannel('window_manager');
    tester.binding.defaultBinaryMessenger.setMockMethodCallHandler(
      channel,
      (call) async {
        calls.add(call.method);
        if (call.method == 'isMinimized') return false;
        return null;
      },
    );
    addTearDown(() => tester.binding.defaultBinaryMessenger
        .setMockMethodCallHandler(channel, null));

    final state = await AppState.create(
      pollSnapshots: false,
      manageWindowsStartup: false,
    );
    addTearDown(state.dispose);

    await DesktopTrayController(state).applyInitialVisibility(['--autostart']);

    expect(calls, contains('show'));
    expect(calls, isNot(contains('hide')));
  });
}
