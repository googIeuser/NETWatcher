import 'package:flutter/material.dart';

import 'app_state.dart';
import 'console_dashboard.dart';
import 'glass.dart';
import 'motion.dart';
import 'pages.dart';
import 'restored_pages.dart';
import 'theme.dart';

class NetworkConsoleShell extends StatefulWidget {
  const NetworkConsoleShell({super.key, required this.state});
  final AppState state;

  @override
  State<NetworkConsoleShell> createState() => _NetworkConsoleShellState();
}

class _NetworkConsoleShellState extends State<NetworkConsoleShell> {
  int selected = 0;
  static const sections = [
    (Icons.hub_rounded, 'Overview'),
    (Icons.query_stats_rounded, 'Statistics'),
    (Icons.bolt_rounded, 'Incidents'),
    (Icons.file_present_outlined, 'Reports'),
    (Icons.dns_outlined, 'Targets'),
    (Icons.tune_rounded, 'Settings'),
  ];

  Widget _page() => KeyedSubtree(
        key: ValueKey<int>(selected),
        child: switch (selected) {
          0 => ConsoleDashboardPage(
              state: widget.state,
              onViewTargets: () => setState(() => selected = 4),
            ),
          1 => StatisticsPage(state: widget.state),
          2 => OutagesPage(state: widget.state),
          3 => ReportsPage(state: widget.state),
          4 => TargetsPage(state: widget.state),
          _ => RestoredSettingsPage(state: widget.state),
        },
      );

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    final snapshot = widget.state.snapshot;
    final live = snapshot.monitoring && snapshot.samples > 0;
    final stateLabel = !snapshot.monitoring
        ? 'STANDBY'
        : live
            ? snapshot.connectionState.toUpperCase()
            : 'CHECKING';
    final stateColor = !snapshot.monitoring
        ? scheme.onSurfaceVariant
        : snapshot.connectionState == 'offline'
            ? scheme.error
            : const Color(0xFF72B984);

    return Scaffold(
      body: Stack(
        children: [
          const Positioned.fill(child: GlassAmbient()),
          SafeArea(
            child: Row(
              children: [
                Container(
                  width: 78,
                  decoration: const BoxDecoration(
                    color: NetWatcherTheme.ink,
                    border: Border(right: BorderSide(color: Color(0xFF385348))),
                  ),
                  child: Column(
                    children: [
                      Padding(
                        padding: const EdgeInsets.fromLTRB(12, 14, 12, 18),
                        child: Image.asset(
                          'assets/app_icon.png',
                          width: 48,
                          height: 48,
                          filterQuality: FilterQuality.high,
                          semanticLabel: 'NetWatcher logo',
                        ),
                      ),
                      Container(height: 1, color: const Color(0xFF385348)),
                      const SizedBox(height: 16),
                      for (var i = 0; i < 5; i++)
                        _RailItem(
                          key: ValueKey<String>('nav-$i'),
                          icon: sections[i].$1,
                          label: sections[i].$2,
                          selected: selected == i,
                          onTap: () => setState(() => selected = i),
                        ),
                      const Spacer(),
                      _RailItem(
                        key: const ValueKey<String>('nav-5'),
                        icon: sections[5].$1,
                        label: sections[5].$2,
                        selected: selected == 5,
                        onTap: () => setState(() => selected = 5),
                      ),
                      const SizedBox(height: 12),
                    ],
                  ),
                ),
                Expanded(
                  child: Column(
                    children: [
                      GlassSurface(
                        radius: 0,
                        blur: 16,
                        child: SizedBox(
                          height: 70,
                          child: Padding(
                            padding: const EdgeInsets.symmetric(horizontal: 24),
                            child: Row(
                              children: [
                                Column(
                                  mainAxisAlignment: MainAxisAlignment.center,
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text('NETWATCHER',
                                        style: Theme.of(context)
                                            .textTheme
                                            .titleMedium
                                            ?.copyWith(
                                              fontWeight: FontWeight.w900,
                                              letterSpacing: 1.4,
                                            )),
                                    Text('NETWORK CONTROL DESK',
                                        style: Theme.of(context)
                                            .textTheme
                                            .labelSmall
                                            ?.copyWith(
                                              color: scheme.onSurfaceVariant,
                                              letterSpacing: 1.3,
                                            )),
                                  ],
                                ),
                                const Spacer(),
                                GlassSurface(
                                  radius: 20,
                                  blur: 10,
                                  padding: const EdgeInsets.symmetric(
                                      horizontal: 11, vertical: 7),
                                  child: Row(
                                    mainAxisSize: MainAxisSize.min,
                                    children: [
                                      Container(
                                        width: 7,
                                        height: 7,
                                        decoration: BoxDecoration(
                                            color: stateColor,
                                            shape: BoxShape.circle),
                                      ),
                                      const SizedBox(width: 8),
                                      Text(stateLabel,
                                          style: Theme.of(context)
                                              .textTheme
                                              .labelSmall
                                              ?.copyWith(
                                                fontWeight: FontWeight.w800,
                                                letterSpacing: 1.1,
                                              )),
                                    ],
                                  ),
                                ),
                                const SizedBox(width: 16),
                                Text('v${snapshot.version}',
                                    style: Theme.of(context)
                                        .textTheme
                                        .labelMedium
                                        ?.copyWith(
                                          color: scheme.onSurfaceVariant,
                                          fontFamily: 'Consolas',
                                        )),
                              ],
                            ),
                          ),
                        ),
                      ),
                      if (widget.state.error != null)
                        MaterialBanner(
                          content: Text(widget.state.error!),
                          actions: [
                            TextButton(
                              onPressed: widget.state.refreshSnapshot,
                              child: const Text('Retry'),
                            ),
                          ],
                        ),
                      Expanded(child: FadeSlideSwitcher(child: _page())),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _RailItem extends StatefulWidget {
  const _RailItem({
    super.key,
    required this.icon,
    required this.label,
    required this.selected,
    required this.onTap,
  });

  final IconData icon;
  final String label;
  final bool selected;
  final VoidCallback onTap;

  @override
  State<_RailItem> createState() => _RailItemState();
}

class _RailItemState extends State<_RailItem> {
  bool focused = false;

  @override
  Widget build(BuildContext context) => Tooltip(
        message: widget.label,
        child: Semantics(
          label: widget.label,
          button: true,
          selected: widget.selected,
          child: Material(
            color: Colors.transparent,
            child: InkWell(
              onTap: widget.onTap,
              onFocusChange: (value) => setState(() => focused = value),
              mouseCursor: SystemMouseCursors.click,
              hoverColor: Colors.transparent,
              splashColor: Colors.transparent,
              highlightColor: Colors.transparent,
              focusColor: Colors.transparent,
              overlayColor: const WidgetStatePropertyAll(Colors.transparent),
              child: Container(
                width: 52,
                height: 50,
                margin: const EdgeInsets.symmetric(vertical: 4),
                decoration: BoxDecoration(
                  color: Colors.transparent,
                  border: Border(
                    left: BorderSide(
                      color: widget.selected
                          ? NetWatcherTheme.signal
                          : focused
                              ? Colors.white
                              : Colors.transparent,
                      width: 3,
                    ),
                  ),
                ),
                child: Icon(widget.icon,
                    size: 23,
                    color: widget.selected
                        ? NetWatcherTheme.signal
                        : const Color(0xFF9CB3A3)),
              ),
            ),
          ),
        ),
      );
}
