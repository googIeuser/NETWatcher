import 'package:flutter/material.dart';

import 'app_state.dart';
import 'models.dart';
import 'motion.dart';
import 'widgets.dart';

class RestoredDashboardPage extends StatelessWidget {
  const RestoredDashboardPage({super.key, required this.state});

  final AppState state;

  @override
  Widget build(BuildContext context) {
    final snapshot = state.snapshot;
    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        _RestoredPageHeader(
          eyebrow: 'OVERVIEW',
          title: 'Connection dashboard',
          subtitle:
              'Live local monitoring powered by the Rust core architecture.',
          trailing: FilledButton.icon(
            onPressed: state.toggleMonitoring,
            icon: Icon(snapshot.monitoring ? Icons.stop : Icons.play_arrow),
            label: Text(
              snapshot.monitoring ? 'Stop monitoring' : 'Start monitoring',
            ),
          ),
        ),
        const SizedBox(height: 18),
        _RestoredHero(snapshot: snapshot),
        const SizedBox(height: 16),
        LayoutBuilder(
          builder: (context, constraints) {
            final columns = constraints.maxWidth >= 1000
                ? 4
                : constraints.maxWidth >= 560
                    ? 2
                    : 1;
            final width =
                (constraints.maxWidth - (columns - 1) * 12) / columns;
            return Wrap(
              spacing: 12,
              runSpacing: 12,
              children: [
                SizedBox(
                  width: width,
                  child: MetricCard(
                    label: 'Average latency',
                    value: snapshot.averageLatency.toStringAsFixed(1),
                    unit: 'ms',
                    icon: Icons.speed,
                  ),
                ),
                SizedBox(
                  width: width,
                  child: MetricCard(
                    label: 'Packet loss',
                    value: snapshot.packetLoss.toStringAsFixed(1),
                    unit: '%',
                    icon: Icons.signal_cellular_alt,
                  ),
                ),
                SizedBox(
                  width: width,
                  child: MetricCard(
                    label: 'Jitter',
                    value: snapshot.jitter.toStringAsFixed(1),
                    unit: 'ms',
                    icon: Icons.show_chart,
                  ),
                ),
                SizedBox(
                  width: width,
                  child: MetricCard(
                    label: 'Samples',
                    value: snapshot.samples.toString(),
                    icon: Icons.data_usage,
                  ),
                ),
              ],
            );
          },
        ),
        const SizedBox(height: 16),
        LayoutBuilder(
          builder: (context, constraints) {
            final sideBySide = constraints.maxWidth >= 1050;
            final chart = Panel(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'LIVE SIGNAL',
                    style: Theme.of(context).textTheme.labelSmall?.copyWith(
                          color: Theme.of(context).colorScheme.primary,
                          fontWeight: FontWeight.w800,
                          letterSpacing: 1.2,
                        ),
                  ),
                  const SizedBox(height: 7),
                  Wrap(
                    spacing: 14,
                    runSpacing: 10,
                    crossAxisAlignment: WrapCrossAlignment.center,
                    children: [
                      Text(
                        'Latency history',
                        style: Theme.of(context).textTheme.titleMedium?.copyWith(
                              fontWeight: FontWeight.w800,
                            ),
                      ),
                      DropdownButton<int>(
                        value: state.config.graphRangeMinutes,
                        underline: const SizedBox.shrink(),
                        items: const [
                          DropdownMenuItem(value: 5, child: Text('5 minutes')),
                          DropdownMenuItem(value: 30, child: Text('30 minutes')),
                          DropdownMenuItem(value: 60, child: Text('1 hour')),
                          DropdownMenuItem(value: 1440, child: Text('24 hours')),
                        ],
                        onChanged: (value) {
                          if (value != null) state.setGraphRange(value);
                        },
                      ),
                    ],
                  ),
                  const SizedBox(height: 18),
                  LatencyChart(
                    targets: snapshot.targets,
                    rangeMinutes: state.config.graphRangeMinutes,
                  ),
                ],
              ),
            );
            final targets = Panel(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Row(
                    children: [
                      Expanded(
                        child: Text(
                          'Monitored targets',
                          style: Theme.of(context)
                              .textTheme
                              .titleMedium
                              ?.copyWith(fontWeight: FontWeight.w800),
                        ),
                      ),
                      Badge(label: Text(snapshot.targets.length.toString())),
                    ],
                  ),
                  const SizedBox(height: 10),
                  for (final target in snapshot.targets)
                    TargetCard(status: target),
                ],
              ),
            );
            if (sideBySide) {
              return Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Expanded(flex: 3, child: chart),
                  const SizedBox(width: 16),
                  Expanded(flex: 2, child: targets),
                ],
              );
            }
            return Column(
              children: [
                chart,
                const SizedBox(height: 16),
                targets,
              ],
            );
          },
        ),
        const SizedBox(height: 16),
        Panel(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Text(
                'Recent events',
                style: Theme.of(context)
                    .textTheme
                    .titleMedium
                    ?.copyWith(fontWeight: FontWeight.w800),
              ),
              const SizedBox(height: 10),
              if (snapshot.recentEvents.isEmpty)
                const Text('No events yet.')
              else
                for (final event in snapshot.recentEvents)
                  ListTile(
                    contentPadding: EdgeInsets.zero,
                    leading: const Icon(Icons.circle, size: 10),
                    title: Text(event.message),
                    subtitle: Text(event.time),
                  ),
            ],
          ),
        ),
      ],
    );
  }
}

class _RestoredHero extends StatelessWidget {
  const _RestoredHero({required this.snapshot});

  final NetworkSnapshot snapshot;

  Color _color() => switch (snapshot.connectionState) {
        'online' => const Color(0xFF42D99A),
        'offline' => const Color(0xFFFF6D80),
        _ => const Color(0xFFFFBD59),
      };

  @override
  Widget build(BuildContext context) {
    final color = _color();
    return Panel(
      padding: const EdgeInsets.all(24),
      child: LayoutBuilder(
        builder: (context, constraints) {
          final compact = constraints.maxWidth < 620;
          final status = Row(
            children: [
              AnimatedContainer(
                duration: NetWatcherMotion.normal,
                curve: NetWatcherMotion.curve,
                width: 56,
                height: 56,
                decoration: BoxDecoration(
                  color: color.withValues(alpha: .12),
                  borderRadius: BorderRadius.circular(18),
                ),
                child: AnimatedSwitcher(
                  duration: NetWatcherMotion.normal,
                  child: Icon(
                    Icons.circle,
                    key: ValueKey<String>(snapshot.connectionState),
                    color: color,
                    size: 18,
                  ),
                ),
              ),
              const SizedBox(width: 18),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    AnimatedSwitcher(
                      duration: NetWatcherMotion.normal,
                      child: Text(
                        snapshot.connectionLabel,
                        key: ValueKey<String>(snapshot.connectionLabel),
                        style:
                            Theme.of(context).textTheme.headlineSmall?.copyWith(
                                  fontWeight: FontWeight.w800,
                                ),
                      ),
                    ),
                    const SizedBox(height: 5),
                    Text(
                      snapshot.monitoring
                          ? 'Connection is being checked continuously.'
                          : 'Start monitoring to collect live measurements.',
                      overflow: TextOverflow.visible,
                    ),
                  ],
                ),
              ),
            ],
          );
          final score = SizedBox(
            width: compact ? double.infinity : 110,
            child: Column(
              children: [
                AnimatedSwitcher(
                  duration: NetWatcherMotion.normal,
                  transitionBuilder: (child, animation) => FadeTransition(
                    opacity: animation,
                    child: ScaleTransition(
                      scale: Tween<double>(begin: .9, end: 1).animate(animation),
                      child: child,
                    ),
                  ),
                  child: Text(
                    snapshot.qualityScore.toString(),
                    key: ValueKey<int>(snapshot.qualityScore),
                    style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                          fontWeight: FontWeight.w900,
                          color: Theme.of(context).colorScheme.primary,
                        ),
                  ),
                ),
                const Text('QUALITY'),
              ],
            ),
          );
          if (compact) {
            return Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [status, const SizedBox(height: 22), score],
            );
          }
          return Row(children: [Expanded(child: status), score]);
        },
      ),
    );
  }
}

class RestoredSettingsPage extends StatefulWidget {
  const RestoredSettingsPage({super.key, required this.state});

  final AppState state;

  @override
  State<RestoredSettingsPage> createState() => _RestoredSettingsPageState();
}

class _RestoredSettingsPageState extends State<RestoredSettingsPage> {
  late NetWatcherConfig draft;

  @override
  void initState() {
    super.initState();
    draft = widget.state.config;
  }

  @override
  Widget build(BuildContext context) => ListView(
        padding: const EdgeInsets.all(24),
        children: [
          const _RestoredPageHeader(
            eyebrow: 'PREFERENCES',
            title: 'Settings',
            subtitle: 'Monitoring, startup and notification preferences.',
          ),
          const SizedBox(height: 18),
          Panel(
            child: Column(
              children: [
                _RestoredSettingRow(
                  label: 'Theme',
                  child: DropdownButtonFormField<String>(
                    initialValue: draft.theme,
                    items: const [
                      DropdownMenuItem(value: 'dark', child: Text('Dark')),
                      DropdownMenuItem(value: 'light', child: Text('Light')),
                    ],
                    onChanged: (value) =>
                        setState(() => draft = draft.copyWith(theme: value)),
                  ),
                ),
                _RestoredSettingRow(
                  label: 'Monitoring interval',
                  child: TextFormField(
                    initialValue: draft.intervalSeconds.toString(),
                    keyboardType: TextInputType.number,
                    decoration: const InputDecoration(suffixText: 'seconds'),
                    onChanged: (value) => draft = draft.copyWith(
                      intervalSeconds:
                          double.tryParse(value) ?? draft.intervalSeconds,
                    ),
                  ),
                ),
                SwitchListTile(
                  contentPadding: EdgeInsets.zero,
                  value: draft.startWithWindows,
                  title: const Text('Start NetWatcher with Windows'),
                  subtitle: const Text(
                    'Adds NetWatcher to the current user startup list.',
                  ),
                  onChanged: (value) => setState(
                    () => draft = draft.copyWith(startWithWindows: value),
                  ),
                ),
                SwitchListTile(
                  contentPadding: EdgeInsets.zero,
                  value: draft.startMinimizedToNotificationArea,
                  title: const Text('Start minimized in the notification area'),
                  onChanged: draft.startWithWindows
                      ? (value) => setState(
                            () => draft = draft.copyWith(
                              startMinimizedToNotificationArea: value,
                            ),
                          )
                      : null,
                ),
                SwitchListTile(
                  contentPadding: EdgeInsets.zero,
                  value: draft.startMonitoringAutomatically,
                  title: const Text('Start monitoring automatically'),
                  onChanged: (value) => setState(
                    () => draft =
                        draft.copyWith(startMonitoringAutomatically: value),
                  ),
                ),
                SwitchListTile(
                  contentPadding: EdgeInsets.zero,
                  value: draft.keepRunningInTrayOnClose,
                  title: const Text(
                    'Keep NetWatcher running in the notification area when the window closes',
                  ),
                  onChanged: (value) => setState(
                    () => draft = draft.copyWith(
                      keepRunningInTrayOnClose: value,
                    ),
                  ),
                ),
                SwitchListTile(
                  contentPadding: EdgeInsets.zero,
                  value: draft.showOutageNotifications,
                  title: const Text('Show outage and recovery notifications'),
                  onChanged: (value) => setState(
                    () =>
                        draft = draft.copyWith(showOutageNotifications: value),
                  ),
                ),
                const SizedBox(height: 14),
                Align(
                  alignment: Alignment.centerRight,
                  child: FilledButton.icon(
                    onPressed: () => widget.state.saveConfig(draft),
                    icon: const Icon(Icons.save_outlined),
                    label: const Text('Save settings'),
                  ),
                ),
              ],
            ),
          ),
        ],
      );
}

class _RestoredSettingRow extends StatelessWidget {
  const _RestoredSettingRow({required this.label, required this.child});

  final String label;
  final Widget child;

  @override
  Widget build(BuildContext context) => Padding(
        padding: const EdgeInsets.only(bottom: 16),
        child: LayoutBuilder(
          builder: (context, constraints) {
            if (constraints.maxWidth < 600) {
              return Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Text(
                    label,
                    style: const TextStyle(fontWeight: FontWeight.w700),
                  ),
                  const SizedBox(height: 8),
                  child,
                ],
              );
            }
            return Row(
              children: [
                Expanded(
                  child: Text(
                    label,
                    style: const TextStyle(fontWeight: FontWeight.w700),
                  ),
                ),
                const SizedBox(width: 20),
                SizedBox(width: 260, child: child),
              ],
            );
          },
        ),
      );
}

class _RestoredPageHeader extends StatelessWidget {
  const _RestoredPageHeader({
    required this.eyebrow,
    required this.title,
    required this.subtitle,
    this.trailing,
  });

  final String eyebrow;
  final String title;
  final String subtitle;
  final Widget? trailing;

  @override
  Widget build(BuildContext context) => LayoutBuilder(
        builder: (context, constraints) {
          final copy = Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                eyebrow,
                style: Theme.of(context).textTheme.labelSmall?.copyWith(
                      color: Theme.of(context).colorScheme.primary,
                      fontWeight: FontWeight.w800,
                      letterSpacing: 1.3,
                    ),
              ),
              const SizedBox(height: 6),
              Text(
                title,
                overflow: TextOverflow.visible,
                style: Theme.of(context)
                    .textTheme
                    .headlineMedium
                    ?.copyWith(fontWeight: FontWeight.w900),
              ),
              const SizedBox(height: 5),
              Text(subtitle, overflow: TextOverflow.visible),
            ],
          );
          if (trailing == null) return copy;
          if (constraints.maxWidth < 620) {
            return Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [copy, const SizedBox(height: 16), trailing!],
            );
          }
          return Row(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              Expanded(child: copy),
              const SizedBox(width: 20),
              trailing!,
            ],
          );
        },
      );
}
