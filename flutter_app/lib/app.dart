import 'dart:async';

import 'package:flutter/material.dart';
import 'package:window_manager/window_manager.dart';

import 'app_state.dart';
import 'desktop_tray.dart';
import 'motion.dart';
import 'pages.dart';
import 'restored_pages.dart';
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
            : AppShell(state: widget.state),
      ),
    );
  }
}

class AppShell extends StatefulWidget {
  const AppShell({super.key, required this.state});
  final AppState state;

  @override
  State<AppShell> createState() => _AppShellState();
}

class _AppShellState extends State<AppShell> {
  int selected = 0;

  static const destinations = [
    (Icons.dashboard_outlined, Icons.dashboard, 'Dashboard'),
    (Icons.bar_chart_outlined, Icons.bar_chart, 'Statistics'),
    (Icons.warning_amber_outlined, Icons.warning_amber, 'Outage History'),
    (Icons.description_outlined, Icons.description, 'Reports'),
    (Icons.radar_outlined, Icons.radar, 'Targets'),
    (Icons.settings_outlined, Icons.settings, 'Settings'),
  ];

  Widget page() => KeyedSubtree(
        key: ValueKey<int>(selected),
        child: switch (selected) {
          0 => RestoredDashboardPage(state: widget.state),
          1 => StatisticsPage(state: widget.state),
          2 => OutagesPage(state: widget.state),
          3 => ReportsPage(state: widget.state),
          4 => TargetsPage(state: widget.state),
          _ => RestoredSettingsPage(state: widget.state),
        },
      );

  void selectPage(int value) {
    if (selected == value) return;
    setState(() => selected = value);
  }

  Widget animatedContent() {
    return Column(
      children: [
        AnimatedSwitcher(
          duration: NetWatcherMotion.normal,
          child: widget.state.error == null
              ? const SizedBox.shrink(key: ValueKey('no-error'))
              : MaterialBanner(
                  key: const ValueKey('error'),
                  content: Text(widget.state.error!),
                  actions: [
                    TextButton(
                      onPressed: widget.state.refreshSnapshot,
                      child: const Text('Retry'),
                    ),
                  ],
                ),
        ),
        Expanded(child: FadeSlideSwitcher(child: page())),
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (context, constraints) {
        final expandedSidebar = constraints.maxWidth >= 1120;
        final desktopSidebar = constraints.maxWidth >= 760;

        if (desktopSidebar) {
          return Scaffold(
            body: Row(
              children: [
                _DesktopSidebar(
                  expanded: expandedSidebar,
                  selected: selected,
                  destinations: destinations,
                  version: widget.state.snapshot.version,
                  onSelected: selectPage,
                ),
                const VerticalDivider(width: 1),
                Expanded(child: animatedContent()),
              ],
            ),
          );
        }

        return Scaffold(
          appBar: AppBar(
            title: const Text(
              'NetWatcher',
              style: TextStyle(fontWeight: FontWeight.w900),
            ),
          ),
          body: animatedContent(),
          bottomNavigationBar: NavigationBar(
            selectedIndex: selected,
            labelBehavior: NavigationDestinationLabelBehavior.onlyShowSelected,
            onDestinationSelected: selectPage,
            destinations: [
              for (var index = 0; index < destinations.length; index++)
                NavigationDestination(
                  icon: _PointerIcon(
                    tooltip: destinations[index].$3,
                    icon: destinations[index].$1,
                  ),
                  selectedIcon: _PointerIcon(
                    tooltip: destinations[index].$3,
                    icon: destinations[index].$2,
                    selected: true,
                  ),
                  label: destinations[index].$3,
                ),
            ],
          ),
        );
      },
    );
  }
}

class _DesktopSidebar extends StatelessWidget {
  const _DesktopSidebar({
    required this.expanded,
    required this.selected,
    required this.destinations,
    required this.version,
    required this.onSelected,
  });

  final bool expanded;
  final int selected;
  final List<(IconData, IconData, String)> destinations;
  final String version;
  final ValueChanged<int> onSelected;

  @override
  Widget build(BuildContext context) {
    return AnimatedContainer(
      duration: NetWatcherMotion.normal,
      curve: NetWatcherMotion.curve,
      width: expanded ? 228 : 82,
      color: Theme.of(context).navigationRailTheme.backgroundColor,
      child: SafeArea(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Padding(
              padding: EdgeInsets.fromLTRB(expanded ? 18 : 13, 17, 13, 22),
              child: Row(
                mainAxisAlignment:
                    expanded ? MainAxisAlignment.start : MainAxisAlignment.center,
                children: [
                  ClipRRect(
                    borderRadius: BorderRadius.circular(11),
                    child: Image.asset(
                      'assets/app_icon.png',
                      width: 40,
                      height: 40,
                      filterQuality: FilterQuality.high,
                    ),
                  ),
                  if (expanded) ...[
                    const SizedBox(width: 12),
                    const Expanded(
                      child: Text(
                        'NetWatcher',
                        overflow: TextOverflow.fade,
                        softWrap: false,
                        style: TextStyle(fontWeight: FontWeight.w900),
                      ),
                    ),
                  ],
                ],
              ),
            ),
            for (var index = 0; index < destinations.length; index++)
              _SidebarDestination(
                key: ValueKey<String>('nav-$index'),
                expanded: expanded,
                selected: selected == index,
                icon: selected == index
                    ? destinations[index].$2
                    : destinations[index].$1,
                label: destinations[index].$3,
                onTap: () => onSelected(index),
              ),
            const Spacer(),
            Padding(
              padding: const EdgeInsets.all(18),
              child: Text(
                version,
                textAlign: TextAlign.center,
                style: Theme.of(context).textTheme.labelSmall,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _SidebarDestination extends StatefulWidget {
  const _SidebarDestination({
    super.key,
    required this.expanded,
    required this.selected,
    required this.icon,
    required this.label,
    required this.onTap,
  });

  final bool expanded;
  final bool selected;
  final IconData icon;
  final String label;
  final VoidCallback onTap;

  @override
  State<_SidebarDestination> createState() => _SidebarDestinationState();
}

class _SidebarDestinationState extends State<_SidebarDestination> {
  bool hovered = false;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    final active = widget.selected || hovered;
    final foreground = widget.selected
        ? scheme.primary
        : active
            ? scheme.onSurface
            : scheme.onSurfaceVariant;

    return MouseRegion(
      cursor: SystemMouseCursors.click,
      onEnter: (_) => setState(() => hovered = true),
      onExit: (_) => setState(() => hovered = false),
      child: Tooltip(
        message: widget.expanded ? '' : widget.label,
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 11, vertical: 3),
          child: Material(
            color: Colors.transparent,
            child: InkWell(
              mouseCursor: SystemMouseCursors.click,
              borderRadius: BorderRadius.circular(12),
              onTap: widget.onTap,
              child: AnimatedContainer(
                duration: NetWatcherMotion.fast,
                curve: NetWatcherMotion.curve,
                constraints: const BoxConstraints(minHeight: 46),
                padding: EdgeInsets.symmetric(
                  horizontal: widget.expanded ? 13 : 0,
                  vertical: 10,
                ),
                decoration: BoxDecoration(
                  color: widget.selected
                      ? scheme.primary.withValues(alpha: .14)
                      : hovered
                          ? scheme.onSurface.withValues(alpha: .06)
                          : Colors.transparent,
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(
                    color: widget.selected
                        ? scheme.primary.withValues(alpha: .25)
                        : Colors.transparent,
                  ),
                ),
                child: widget.expanded
                    ? Row(
                        children: [
                          AnimatedScale(
                            duration: NetWatcherMotion.fast,
                            scale: hovered ? 1.08 : 1,
                            child: Icon(
                              widget.icon,
                              color: foreground,
                              size: 21,
                            ),
                          ),
                          const SizedBox(width: 13),
                          Expanded(
                            child: Text(
                              widget.label,
                              overflow: TextOverflow.ellipsis,
                              style: TextStyle(
                                color: foreground,
                                fontWeight: widget.selected
                                    ? FontWeight.w800
                                    : FontWeight.w600,
                              ),
                            ),
                          ),
                        ],
                      )
                    : Center(
                        child: SizedBox.square(
                          dimension: 24,
                          child: AnimatedScale(
                            duration: NetWatcherMotion.fast,
                            scale: hovered ? 1.08 : 1,
                            child: Icon(
                              widget.icon,
                              color: foreground,
                              size: 21,
                            ),
                          ),
                        ),
                      ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _PointerIcon extends StatelessWidget {
  const _PointerIcon({
    required this.tooltip,
    required this.icon,
    this.selected = false,
  });

  final String tooltip;
  final IconData icon;
  final bool selected;

  @override
  Widget build(BuildContext context) {
    return MouseRegion(
      cursor: SystemMouseCursors.click,
      child: Tooltip(
        message: tooltip,
        child: AnimatedScale(
          duration: NetWatcherMotion.fast,
          scale: selected ? 1.06 : 1,
          child: Icon(icon),
        ),
      ),
    );
  }
}
