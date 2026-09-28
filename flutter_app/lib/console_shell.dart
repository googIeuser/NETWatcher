import 'package:flutter/material.dart';

import 'app_state.dart';
import 'console_dashboard.dart';
import 'motion.dart';
import 'pages.dart';
import 'restored_pages.dart';

class NetworkConsoleShell extends StatefulWidget {
  const NetworkConsoleShell({super.key, required this.state});

  final AppState state;

  @override
  State<NetworkConsoleShell> createState() => _NetworkConsoleShellState();
}

class _NetworkConsoleShellState extends State<NetworkConsoleShell> {
  int selected = 0;

  static const sections = [
    (Icons.grid_view_rounded, 'Overview'),
    (Icons.query_stats_rounded, 'Statistics'),
    (Icons.bolt_rounded, 'Incidents'),
    (Icons.file_present_outlined, 'Reports'),
    (Icons.hub_outlined, 'Targets'),
    (Icons.tune_rounded, 'Settings'),
  ];
  static const compactLabels = [
    'Overview',
    'Stats',
    'History',
    'Reports',
    'Targets',
    'Settings',
  ];

  Widget _page() => KeyedSubtree(
        key: ValueKey<int>(selected),
        child: switch (selected) {
          0 => ConsoleDashboardPage(state: widget.state),
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
    final monitoring = widget.state.snapshot.monitoring;
    final live = monitoring && widget.state.snapshot.samples > 0;
    final stateLabel = !monitoring
        ? 'STANDBY'
        : live
            ? widget.state.snapshot.connectionState.toUpperCase()
            : 'CHECKING';
    final stateColor = !monitoring
        ? scheme.onSurfaceVariant
        : widget.state.snapshot.connectionState == 'offline'
            ? scheme.error
            : const Color(0xFF72B984);

    return Scaffold(
      body: SafeArea(
        child: Column(
          children: [
            Container(
              height: 79,
              padding: const EdgeInsets.symmetric(horizontal: 26),
              decoration: BoxDecoration(
                color: scheme.surface,
                border: Border(bottom: BorderSide(color: scheme.outline)),
              ),
              child: Row(
                children: [
                  Image.asset(
                    'assets/app_icon.png',
                    width: 39,
                    height: 39,
                    filterQuality: FilterQuality.high,
                    semanticLabel: 'NetWatcher logo',
                  ),
                  const SizedBox(width: 13),
                  Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text('NETWATCHER',
                          style:
                              Theme.of(context).textTheme.titleMedium?.copyWith(
                                    fontWeight: FontWeight.w900,
                                    letterSpacing: 1.4,
                                  )),
                      Text('NETWORK CONTROL DESK',
                          style:
                              Theme.of(context).textTheme.labelSmall?.copyWith(
                                    color: scheme.onSurfaceVariant,
                                    letterSpacing: 1.6,
                                  )),
                    ],
                  ),
                  const Spacer(),
                  Container(
                    padding:
                        const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                    decoration: BoxDecoration(
                      border: Border.all(color: scheme.outline),
                      borderRadius: BorderRadius.circular(4),
                    ),
                    child: Row(
                      children: [
                        Container(
                          width: 7,
                          height: 7,
                          decoration: BoxDecoration(
                            color: stateColor,
                            shape: BoxShape.circle,
                          ),
                        ),
                        const SizedBox(width: 8),
                        Text(stateLabel,
                            style: Theme.of(context)
                                .textTheme
                                .labelSmall
                                ?.copyWith(
                                  fontWeight: FontWeight.w800,
                                  letterSpacing: 1.2,
                                )),
                      ],
                    ),
                  ),
                  const SizedBox(width: 16),
                  Text('v${widget.state.snapshot.version}',
                      style: Theme.of(context).textTheme.labelMedium?.copyWith(
                            color: scheme.onSurfaceVariant,
                            fontFamily: 'Consolas',
                          )),
                ],
              ),
            ),
            Container(
              height: 55,
              decoration: BoxDecoration(
                color: scheme.surface,
                border: Border(bottom: BorderSide(color: scheme.outline)),
              ),
              child: LayoutBuilder(builder: (context, constraints) {
                final tabs = [
                  for (var i = 0; i < sections.length; i++)
                    _SectionTab(
                      key: ValueKey<String>('nav-$i'),
                      icon: sections[i].$1,
                      label: constraints.maxWidth < 1000
                          ? compactLabels[i]
                          : sections[i].$2,
                      selected: selected == i,
                      onTap: () => setState(() => selected = i),
                    ),
                ];
                if (constraints.maxWidth < 760) {
                  return SingleChildScrollView(
                    scrollDirection: Axis.horizontal,
                    child: Row(children: tabs),
                  );
                }
                return Row(
                  children: [for (final tab in tabs) Expanded(child: tab)],
                );
              }),
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
    );
  }
}

class _SectionTab extends StatelessWidget {
  const _SectionTab({
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
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    return Material(
      color: Colors.transparent,
      child: InkWell(
        mouseCursor: SystemMouseCursors.click,
        onTap: onTap,
        child: Container(
          height: 55,
          padding: const EdgeInsets.symmetric(horizontal: 8),
          decoration: BoxDecoration(
            border: Border(
              bottom: BorderSide(
                color: selected ? scheme.primary : Colors.transparent,
                width: 3,
              ),
            ),
          ),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(icon,
                  size: 18,
                  color: selected ? scheme.primary : scheme.onSurfaceVariant),
              const SizedBox(width: 8),
              Flexible(
                child: Text(label,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: Theme.of(context).textTheme.labelLarge?.copyWith(
                          color: selected
                              ? scheme.onSurface
                              : scheme.onSurfaceVariant,
                          fontWeight:
                              selected ? FontWeight.w800 : FontWeight.w600,
                        )),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
