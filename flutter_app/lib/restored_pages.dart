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
    final hasLiveData = snapshot.monitoring && snapshot.samples > 0;
    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        _RestoredPageHeader(
          eyebrow: 'NETWORK OVERVIEW',
          title: 'Dashboard',
          subtitle:
              'Connection health, latency and recent activity in one place.',
          trailing: FilledButton.icon(
            onPressed: state.toggleMonitoring,
            icon: Icon(snapshot.monitoring ? Icons.stop : Icons.play_arrow),
            label: Text(
              snapshot.monitoring ? 'Stop monitoring' : 'Start monitoring',
            ),
          ),
        ),
        const SizedBox(height: 18),
        _RestoredHero(snapshot: snapshot, hasLiveData: hasLiveData),
        const SizedBox(height: 14),
        LayoutBuilder(
          builder: (context, constraints) {
            final columns = constraints.maxWidth >= 760
                ? 4
                : constraints.maxWidth >= 440
                    ? 2
                    : 1;
            final width = (constraints.maxWidth - (columns - 1) * 12) / columns;
            return Wrap(
              spacing: 12,
              runSpacing: 12,
              children: [
                SizedBox(
                  width: width,
                  child: MetricCard(
                    label: 'Average latency',
                    value: hasLiveData
                        ? snapshot.averageLatency.toStringAsFixed(1)
                        : '—',
                    unit: hasLiveData ? 'ms' : '',
                    icon: Icons.speed,
                  ),
                ),
                SizedBox(
                  width: width,
                  child: MetricCard(
                    label: 'Packet loss',
                    value: hasLiveData
                        ? snapshot.packetLoss.toStringAsFixed(1)
                        : '—',
                    unit: hasLiveData ? '%' : '',
                    icon: Icons.signal_cellular_alt,
                  ),
                ),
                SizedBox(
                  width: width,
                  child: MetricCard(
                    label: 'Jitter',
                    value:
                        hasLiveData ? snapshot.jitter.toStringAsFixed(1) : '—',
                    unit: hasLiveData ? 'ms' : '',
                    icon: Icons.show_chart,
                  ),
                ),
                SizedBox(
                  width: width,
                  child: MetricCard(
                    label: 'Samples',
                    value: hasLiveData ? snapshot.samples.toString() : '—',
                    icon: Icons.data_usage,
                  ),
                ),
              ],
            );
          },
        ),
        const SizedBox(height: 14),
        LayoutBuilder(
          builder: (context, constraints) {
            final sideBySide = constraints.maxWidth >= 860;
            final chart = Panel(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Wrap(
                    alignment: WrapAlignment.spaceBetween,
                    spacing: 12,
                    runSpacing: 10,
                    crossAxisAlignment: WrapCrossAlignment.center,
                    children: [
                      Text(
                        'Latency history',
                        style:
                            Theme.of(context).textTheme.titleMedium?.copyWith(
                                  fontWeight: FontWeight.w800,
                                ),
                      ),
                      SizedBox(
                        width: 176,
                        child: DropdownButtonFormField<int>(
                          key: ValueKey<int>(state.config.graphRangeMinutes),
                          initialValue: state.config.graphRangeMinutes,
                          isExpanded: true,
                          decoration: const InputDecoration(
                            labelText: 'History range',
                          ),
                          items: const [
                            DropdownMenuItem(
                              value: 5,
                              child: Text('Last 5 minutes'),
                            ),
                            DropdownMenuItem(
                              value: 30,
                              child: Text('Last 30 minutes'),
                            ),
                            DropdownMenuItem(
                              value: 60,
                              child: Text('Last hour'),
                            ),
                            DropdownMenuItem(
                              value: 1440,
                              child: Text('Last 24 hours'),
                            ),
                          ],
                          onChanged: (value) {
                            if (value != null) state.setGraphRange(value);
                          },
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),
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
                      Badge(
                        backgroundColor: Theme.of(context)
                            .colorScheme
                            .primary
                            .withValues(alpha: .14),
                        textColor: Theme.of(context).colorScheme.primary,
                        label: Text(snapshot.targets.length.toString()),
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  if (snapshot.targets.isEmpty)
                    const Padding(
                      padding: EdgeInsets.symmetric(vertical: 26),
                      child:
                          Text('Targets will appear when monitoring starts.'),
                    )
                  else
                    for (var index = 0;
                        index < snapshot.targets.length;
                        index++)
                      _DashboardTargetRow(
                        status: snapshot.targets[index],
                        showDivider: index > 0,
                        showLatency: hasLiveData,
                      ),
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
  const _RestoredHero({required this.snapshot, required this.hasLiveData});

  final NetworkSnapshot snapshot;
  final bool hasLiveData;

  Color _color() => switch (snapshot.connectionState) {
        'online' => const Color(0xFF42D99A),
        'offline' => const Color(0xFFFF6D80),
        _ => const Color(0xFFFFBD59),
      };

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    final color = snapshot.monitoring ? _color() : scheme.onSurfaceVariant;
    final title = !snapshot.monitoring
        ? 'Monitoring is off'
        : hasLiveData
            ? snapshot.connectionLabel
            : 'Checking connection';
    final description = !snapshot.monitoring
        ? 'Start monitoring to see live connection health.'
        : hasLiveData
            ? 'Your connection is being checked continuously.'
            : 'Waiting for the first measurements.';
    return Panel(
      padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 18),
      child: LayoutBuilder(
        builder: (context, constraints) {
          final compact = constraints.maxWidth < 560;
          final status = Row(
            children: [
              AnimatedContainer(
                duration: NetWatcherMotion.normal,
                curve: NetWatcherMotion.curve,
                width: 48,
                height: 48,
                decoration: BoxDecoration(
                  color: color.withValues(alpha: .12),
                  borderRadius: BorderRadius.circular(14),
                ),
                child: AnimatedSwitcher(
                  duration: NetWatcherMotion.normal,
                  child: Icon(
                    snapshot.monitoring ? Icons.network_check : Icons.pause,
                    key: ValueKey<String>(snapshot.connectionState),
                    color: color,
                    size: 24,
                  ),
                ),
              ),
              const SizedBox(width: 14),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    AnimatedSwitcher(
                      duration: NetWatcherMotion.normal,
                      child: Text(
                        title,
                        key: ValueKey<String>(title),
                        style: Theme.of(context).textTheme.titleLarge?.copyWith(
                              fontWeight: FontWeight.w800,
                            ),
                      ),
                    ),
                    const SizedBox(height: 3),
                    Text(
                      description,
                      overflow: TextOverflow.visible,
                      style: Theme.of(context).textTheme.bodySmall?.copyWith(
                            color: scheme.onSurfaceVariant,
                          ),
                    ),
                  ],
                ),
              ),
            ],
          );
          final score = Container(
            key: hasLiveData
                ? null
                : const ValueKey<String>('dashboard-empty-state'),
            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 9),
            decoration: BoxDecoration(
              color: color.withValues(alpha: .09),
              borderRadius: BorderRadius.circular(12),
            ),
            child: hasLiveData
                ? Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text('CONNECTION QUALITY',
                          style: Theme.of(context).textTheme.labelSmall),
                      Text.rich(
                        TextSpan(
                          text: snapshot.qualityScore.toString(),
                          style:
                              Theme.of(context).textTheme.titleLarge?.copyWith(
                                    color: color,
                                    fontWeight: FontWeight.w800,
                                  ),
                          children: [
                            TextSpan(
                              text: ' / 100',
                              style: Theme.of(context).textTheme.bodySmall,
                            ),
                          ],
                        ),
                      ),
                    ],
                  )
                : Text(
                    snapshot.monitoring ? 'GATHERING DATA' : 'NO LIVE DATA',
                    style: Theme.of(context).textTheme.labelSmall?.copyWith(
                          color: color,
                          fontWeight: FontWeight.w800,
                          letterSpacing: .8,
                        ),
                  ),
          );
          if (compact) {
            return Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                status,
                const SizedBox(height: 14),
                Align(alignment: Alignment.centerLeft, child: score),
              ],
            );
          }
          return Row(children: [
            Expanded(child: status),
            const SizedBox(width: 16),
            score
          ]);
        },
      ),
    );
  }
}

class _DashboardTargetRow extends StatelessWidget {
  const _DashboardTargetRow({
    required this.status,
    required this.showDivider,
    required this.showLatency,
  });

  final TargetStatus status;
  final bool showDivider;
  final bool showLatency;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    final color = switch (status.state) {
      'online' => const Color(0xFF42D99A),
      'offline' => scheme.error,
      _ => scheme.onSurfaceVariant,
    };
    return Container(
      padding: const EdgeInsets.symmetric(vertical: 11),
      decoration: BoxDecoration(
        border: showDivider
            ? Border(top: BorderSide(color: Theme.of(context).dividerColor))
            : null,
      ),
      child: Row(
        children: [
          Container(
            width: 8,
            height: 8,
            decoration: BoxDecoration(color: color, shape: BoxShape.circle),
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  status.target.name,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(fontWeight: FontWeight.w700),
                ),
                Text(
                  status.target.host,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: scheme.onSurfaceVariant,
                      ),
                ),
              ],
            ),
          ),
          const SizedBox(width: 10),
          Text(
            showLatency && status.state == 'online'
                ? '${status.latency.toStringAsFixed(1)} ms'
                : status.state == 'offline'
                    ? 'Offline'
                    : '—',
            style: TextStyle(color: color, fontWeight: FontWeight.w700),
          ),
        ],
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
