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
  @override
  void initState() {
    super.initState();
    if (widget.tray != null) {
      windowManager.addListener(this);
    }
  }

  @override
  void dispose() {
    if (widget.tray != null) {
      windowManager.removeListener(this);
    }
    super.dispose();
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
    return AnimatedBuilder(
      animation: widget.state,
      builder: (context, _) => MaterialApp(
        debugShowCheckedModeBanner: false,
        title: 'NetWatcher',
        theme: NetWatcherTheme.light(),
        darkTheme: NetWatcherTheme.dark(),
        themeMode: widget.state.config.theme == 'light'
            ? ThemeMode.light
            : ThemeMode.dark,
        themeAnimationDuration: NetWatcherMotion.slow,
        themeAnimationCurve: NetWatcherMotion.emphasizedCurve,
        home: widget.state.loading
            ? const Scaffold(body: Center(child: CircularProgressIndicator()))
            : NetworkConsoleShell(state: widget.state),
      ),
    );
  }
}
