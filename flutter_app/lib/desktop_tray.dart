import 'dart:async';
import 'dart:io';

import 'package:tray_manager/tray_manager.dart';
import 'package:window_manager/window_manager.dart';

import 'app_state.dart';

class DesktopTrayController with TrayListener {
  DesktopTrayController(this.state);

  final AppState state;

  bool _initialised = false;
  bool _exiting = false;
  bool? _lastMonitoring;

  Future<void> initialise() async {
    if (!Platform.isWindows || _initialised) return;

    final iconPath = _findIconPath();
    await windowManager.setIcon(iconPath);
    await trayManager.setIcon(iconPath);
    await trayManager.setToolTip('NetWatcher 4.0.2');

    trayManager.addListener(this);
    _lastMonitoring = state.snapshot.monitoring;
    state.addListener(_stateChanged);
    await _rebuildMenu();
    _initialised = true;
  }

  String _findIconPath() {
    final separator = Platform.pathSeparator;
    final executableDir = File(Platform.resolvedExecutable).parent.path;
    final candidates = [
      '$executableDir${separator}data${separator}flutter_assets'
          '${separator}assets${separator}app_icon.ico',
      '${Directory.current.path}${separator}assets${separator}app_icon.ico',
      '${Directory.current.path}${separator}flutter_app'
          '${separator}assets${separator}app_icon.ico',
    ];
    for (final candidate in candidates) {
      if (File(candidate).existsSync()) return candidate;
    }
    throw StateError('NetWatcher icon asset was not found.');
  }

  void _stateChanged() {
    final monitoring = state.snapshot.monitoring;
    if (_lastMonitoring == monitoring) return;
    _lastMonitoring = monitoring;
    unawaited(_rebuildMenu());
  }

  Future<void> _rebuildMenu() async {
    if (!Platform.isWindows || _exiting) return;
    final menu = Menu(
      items: [
        MenuItem(key: 'show_window', label: 'Open NetWatcher'),
        MenuItem(
          key: 'toggle_monitoring',
          label: state.snapshot.monitoring
              ? 'Stop monitoring'
              : 'Start monitoring',
        ),
        MenuItem.separator(),
        MenuItem(key: 'exit_app', label: 'Exit'),
      ],
    );
    await trayManager.setContextMenu(menu);
    await trayManager.setToolTip(
      state.snapshot.monitoring
          ? 'NetWatcher 4.0.2 · Monitoring'
          : 'NetWatcher 4.0.2 · Stopped',
    );
  }

  @override
  void onTrayIconMouseDown() {
    unawaited(showWindow());
  }

  @override
  void onTrayIconRightMouseDown() {
    unawaited(trayManager.popUpContextMenu());
  }

  @override
  void onTrayMenuItemClick(MenuItem menuItem) {
    switch (menuItem.key) {
      case 'show_window':
        unawaited(showWindow());
        break;
      case 'toggle_monitoring':
        unawaited(_toggleMonitoring());
        break;
      case 'exit_app':
        unawaited(exitApplication());
        break;
    }
  }

  Future<void> _toggleMonitoring() async {
    await state.toggleMonitoring();
    await _rebuildMenu();
  }

  Future<void> showWindow() async {
    if (_exiting) return;
    await windowManager.setSkipTaskbar(false);
    if (await windowManager.isMinimized()) {
      await windowManager.restore();
    }
    await windowManager.show();
    await windowManager.focus();
  }

  Future<void> hideToTray() async {
    if (_exiting) return;
    await windowManager.setSkipTaskbar(true);
    await windowManager.hide();
  }

  Future<void> applyInitialVisibility(List<String> arguments) async {
    final autoStarted = arguments.any(
      (argument) => argument.toLowerCase() == '--autostart',
    );
    if (autoStarted && state.config.startMinimizedToNotificationArea) {
      await hideToTray();
    } else {
      await showWindow();
    }
  }

  Future<void> exitApplication() async {
    if (_exiting) return;
    _exiting = true;
    state.removeListener(_stateChanged);
    trayManager.removeListener(this);
    if (_initialised) {
      await trayManager.destroy();
    }
    await state.shutdown();
    await windowManager.setPreventClose(false);
    await windowManager.destroy();
  }

  Future<void> dispose() async {
    if (_exiting) return;
    state.removeListener(_stateChanged);
    trayManager.removeListener(this);
    if (_initialised) {
      await trayManager.destroy();
    }
    _initialised = false;
  }
}
