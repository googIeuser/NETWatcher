import 'dart:async';

import 'package:flutter/material.dart';
import 'package:window_manager/window_manager.dart';

import 'app_state.dart';
import 'desktop_tray.dart';
import 'console_shell.dart';
import 'motion.dart';
import 'theme.dart';

class NetWatcherApp extends StatefulWidget {
  const NetWatcherApp({
    super.key,
    required this.state,
    this.tray,
  });

  final AppState state;
  final DesktopTrayController? tray;

  @override
  State<NetWatcherApp> createState() => _NetWatcherAppState();
}

class _NetWatcherAppState extends State<NetWatcherApp> with WindowListener {
  static final _lightTheme = NetWatcherTheme.light();
  static final _darkTheme = NetWatcherTheme.dark();

  late String _theme;
  late bool _loading;

  @override
  void initState() {
    super.initState();
    _theme = widget.state.config.theme;
    _loading = widget.state.loading;
    widget.state.addListener(_appStateChanged);
    if (widget.tray != null) {
      windowManager.addListener(this);
    }
  }

  @override
  void dispose() {
    widget.state.removeListener(_appStateChanged);
    if (widget.tray != null) {
      windowManager.removeListener(this);
    }
    super.dispose();
  }

  /// Rebuilds the app only when the theme or loading state changes. Pages listen
  /// to the state themselves, so snapshot refreshes do not rebuild MaterialApp.
  void _appStateChanged() {
    final theme = widget.state.config.theme;
    final loading = widget.state.loading;
    if (theme == _theme && loading == _loading) return;
    setState(() {
      _theme = theme;
      _loading = loading;
    });
  }

  @override
  void onWindowClose() {
    unawaited(_handleWindowClose());
  }

  Future<void> _handleWindowClose() async {
    final tray = widget.tray;
    if (tray == null) {
      await widget.state.shutdown();
      await windowManager.setPreventClose(false);
      await windowManager.destroy();
      return;
    }
    if (widget.state.config.keepRunningInTrayOnClose) {
      await tray.hideToTray();
    } else {
      await tray.exitApplication();
    }
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      debugShowCheckedModeBanner: false,
      title: 'NetWatcher',
      theme: _lightTheme,
      darkTheme: _darkTheme,
      themeMode: _theme == 'light' ? ThemeMode.light : ThemeMode.dark,
      themeAnimationDuration: NetWatcherMotion.slow,
      themeAnimationCurve: NetWatcherMotion.emphasizedCurve,
      home: _loading
          ? const Scaffold(body: Center(child: CircularProgressIndicator()))
          : NetworkConsoleShell(state: widget.state),
    );
  }
}
