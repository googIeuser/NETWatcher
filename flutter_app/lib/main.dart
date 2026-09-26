import 'package:flutter/material.dart';
import 'package:local_notifier/local_notifier.dart';
import 'package:window_manager/window_manager.dart';

import 'app.dart';
import 'app_state.dart';
import 'desktop_tray.dart';
import 'models.dart';

Future<void> main(List<String> arguments) async {
  WidgetsFlutterBinding.ensureInitialized();
  await windowManager.ensureInitialized();

  Future<void> Function(NetworkEvent)? onOutageEvent;
  try {
    await localNotifier.setup(
      appName: 'NetWatcher',
      shortcutPolicy: ShortcutPolicy.requireCreate,
    );
    onOutageEvent = (event) => LocalNotification(
          title: event.category == 'recovery'
              ? 'Connection recovered'
              : 'Connection problem',
          body: event.message,
        ).show();
  } catch (exception) {
    debugPrint('Windows notifications are unavailable: $exception');
  }

  final state = await AppState.create(onOutageEvent: onOutageEvent);

  const windowOptions = WindowOptions(
    size: Size(1260, 760),
    minimumSize: Size(760, 560),
    center: true,
    backgroundColor: Colors.transparent,
    skipTaskbar: false,
    title: 'NetWatcher',
    titleBarStyle: TitleBarStyle.normal,
  );

  await windowManager.waitUntilReadyToShow(windowOptions, () async {
    await windowManager.setPreventClose(true);
  });

  final trayController = DesktopTrayController(state);
  DesktopTrayController? tray;
  try {
    await trayController.initialise();
    tray = trayController;
  } catch (_) {
    tray = null;
  }

  runApp(NetWatcherApp(state: state, tray: tray));
  if (tray != null) {
    await tray.applyInitialVisibility(arguments);
  } else {
    await windowManager.show();
    await windowManager.focus();
  }
}
