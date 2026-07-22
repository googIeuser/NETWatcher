import 'package:flutter/material.dart';
import 'package:window_manager/window_manager.dart';

import 'app.dart';
import 'app_state.dart';
import 'desktop_tray.dart';

Future<void> main(List<String> arguments) async {
  WidgetsFlutterBinding.ensureInitialized();
  await windowManager.ensureInitialized();

  final state = await AppState.create();

  final windowOptions = WindowOptions(
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
