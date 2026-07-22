import 'package:flutter/material.dart';

import 'app_state.dart';
import 'models.dart';
import 'motion.dart';
import 'widgets.dart';

class DashboardPage extends StatelessWidget {
  const DashboardPage({super.key, required this.state});
  final AppState state;

  @override
  Widget build(BuildContext context) {
    final snapshot = state.snapshot;
    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        _PageHeader(
          eyebrow: 'OVERVIEW',
          title: 'Connection dashboard',
          subtitle:
              'Live local monitoring powered by the new Rust core architecture.',
          trailing: FilledButton.icon(
            onPressed: state.toggleMonitoring,
            icon: Icon(snapshot.monitoring ? Icons.stop : Icons.play_arrow),
            label: Text(snapshot.monitoring ? 'Stop monitoring' : 'Start monitoring'),
          ),
        ),
        const SizedBox(height: 18),
        _Hero(snapshot: snapshot),
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
                  Text('LIVE SIGNAL',
                      style: Theme.of(context).textTheme.labelSmall?.copyWith(
                            color: Theme.of(context).colorScheme.primary,
                            fontWeight: FontWeight.w800,
                            letterSpacing: 1.2,
                          )),
                  const SizedBox(height: 7),
                  Text('Latency history',
                      style: Theme.of(context).textTheme.titleMedium?.copyWith(
                            fontWeight: FontWeight.w800,
                          )),
                  const SizedBox(height: 18),
                  LatencyChart(targets: snapshot.targets),
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
              Text('Recent events',
                  style: Theme.of(context)
                      .textTheme
                      .titleMedium
                      ?.copyWith(fontWeight: FontWeight.w800)),
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

class _Hero extends StatelessWidget {
  const _Hero({required this.snapshot});
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
                        style: Theme.of(context).textTheme.headlineSmall?.copyWith(
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

class StatisticsPage extends StatelessWidget {
  const StatisticsPage({super.key, required this.state});
  final AppState state;

  @override
  Widget build(BuildContext context) => ListView(
        padding: const EdgeInsets.all(24),
        children: [
          const _PageHeader(
            eyebrow: 'ANALYTICS',
            title: 'Statistics',
            subtitle: 'Target-by-target performance summary.',
          ),
          const SizedBox(height: 18),
          Panel(
            child: Column(
              children: [
                for (final item in state.snapshot.targets)
                  TargetCard(status: item),
              ],
            ),
          ),
        ],
      );
}

class OutagesPage extends StatelessWidget {
  const OutagesPage({super.key, required this.state});
  final AppState state;

  @override
  Widget build(BuildContext context) => ListView(
        padding: const EdgeInsets.all(24),
        children: [
          const _PageHeader(
            eyebrow: 'HISTORY',
            title: 'Outage history',
            subtitle: 'Confirmed incidents will appear here.',
          ),
          const SizedBox(height: 18),
          Panel(
            child: Center(
              child: Padding(
                padding: const EdgeInsets.symmetric(vertical: 50),
                child: Column(
                  children: [
                    const Icon(Icons.verified_outlined, size: 48),
                    const SizedBox(height: 14),
                    Text(
                      state.snapshot.outages == 0
                          ? 'No confirmed outages.'
                          : '${state.snapshot.outages} outages recorded.',
                    ),
                  ],
                ),
              ),
            ),
          ),
        ],
      );
}

class ReportsPage extends StatefulWidget {
  const ReportsPage({super.key, required this.state});
  final AppState state;

  @override
  State<ReportsPage> createState() => _ReportsPageState();
}

class _ReportsPageState extends State<ReportsPage> {
  int htmlHours = 24;
  int evidenceDays = 7;
  int diagnosticsHours = 168;

  @override
  Widget build(BuildContext context) {
    final state = widget.state;
    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        _PageHeader(
          eyebrow: 'EXPORT',
          title: 'Reports',
          subtitle:
              'Create shareable HTML reports, ISP evidence and a diagnostics archive from local measurements.',
          trailing: OutlinedButton.icon(
            onPressed: state.reportBusy ? null : state.openReportsFolder,
            icon: const Icon(Icons.folder_open_outlined),
            label: const Text('Open reports folder'),
          ),
        ),
        const SizedBox(height: 18),
        LayoutBuilder(
          builder: (context, constraints) {
            final width = constraints.maxWidth >= 1000
                ? (constraints.maxWidth - 32) / 3
                : constraints.maxWidth >= 620
                    ? (constraints.maxWidth - 16) / 2
                    : constraints.maxWidth;
            return Wrap(
              spacing: 16,
              runSpacing: 16,
              children: [
                _ReportCard(
                  width: width,
                  title: 'HTML report',
                  description:
                      'Connection measurements, target summaries and completed outage events in a printable page.',
                  icon: Icons.description_outlined,
                  busy: state.reportBusy,
                  selector: DropdownButtonFormField<int>(
                    key: ValueKey<int>(htmlHours),
                    initialValue: htmlHours,
                    decoration: const InputDecoration(labelText: 'Measurement range'),
                    items: const [
                      DropdownMenuItem(value: 1, child: Text('Last hour')),
                      DropdownMenuItem(value: 24, child: Text('Last 24 hours')),
                      DropdownMenuItem(value: 168, child: Text('Last 7 days')),
                      DropdownMenuItem(value: 720, child: Text('Last 30 days')),
                    ],
                    onChanged: state.reportBusy
                        ? null
                        : (value) => setState(() => htmlHours = value ?? 24),
                  ),
                  buttonText: 'Create and open HTML report',
                  onPressed: () => state.generateHtmlReport(htmlHours),
                ),
                _ReportCard(
                  width: width,
                  title: 'ISP Evidence Report',
                  description:
                      'Availability, packet loss, latency, jitter and outage evidence formatted for an ISP or regulator.',
                  icon: Icons.fact_check_outlined,
                  busy: state.reportBusy,
                  selector: DropdownButtonFormField<int>(
                    key: ValueKey<int>(evidenceDays),
                    initialValue: evidenceDays,
                    decoration: const InputDecoration(labelText: 'Evidence range'),
                    items: const [
                      DropdownMenuItem(value: 1, child: Text('Last 1 day')),
                      DropdownMenuItem(value: 7, child: Text('Last 7 days')),
                      DropdownMenuItem(value: 30, child: Text('Last 30 days')),
                    ],
                    onChanged: state.reportBusy
                        ? null
                        : (value) => setState(() => evidenceDays = value ?? 7),
                  ),
                  buttonText: 'Create and open evidence report',
                  onPressed: () => state.generateEvidenceReport(evidenceDays),
                ),
                _ReportCard(
                  width: width,
                  title: 'Diagnostics ZIP',
                  description:
                      'Exports settings, snapshot, calculated statistics, outages and the original local CSV logs.',
                  icon: Icons.archive_outlined,
                  busy: state.reportBusy,
                  selector: DropdownButtonFormField<int>(
                    key: ValueKey<int>(diagnosticsHours),
                    initialValue: diagnosticsHours,
                    decoration: const InputDecoration(labelText: 'Summary range'),
                    items: const [
                      DropdownMenuItem(value: 24, child: Text('Last 24 hours')),
                      DropdownMenuItem(value: 168, child: Text('Last 7 days')),
                      DropdownMenuItem(value: 720, child: Text('Last 30 days')),
                    ],
                    onChanged: state.reportBusy
                        ? null
                        : (value) =>
                            setState(() => diagnosticsHours = value ?? 168),
                  ),
                  buttonText: 'Create diagnostics ZIP',
                  onPressed: () => state.exportDiagnostics(diagnosticsHours),
                ),
              ],
            );
          },
        ),
        const SizedBox(height: 16),
        AnimatedSwitcher(
          duration: NetWatcherMotion.normal,
          child: state.reportBusy
              ? const Panel(
                  key: ValueKey<String>('report-progress'),
                  child: Row(
                    children: [
                      SizedBox(
                        width: 22,
                        height: 22,
                        child: CircularProgressIndicator(strokeWidth: 2.5),
                      ),
                      SizedBox(width: 14),
                      Expanded(child: Text('Preparing the report from local data...')),
                    ],
                  ),
                )
              : state.lastReport == null
                  ? const SizedBox.shrink(key: ValueKey<String>('no-report'))
                  : Panel(
                      key: ValueKey<String>(state.lastReport!.path),
                      child: LayoutBuilder(
                        builder: (context, constraints) {
                          final copy = Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Row(
                                children: [
                                  const Icon(Icons.check_circle_outline,
                                      color: Color(0xFF42D99A)),
                                  const SizedBox(width: 10),
                                  Expanded(
                                    child: Text(
                                      state.reportNotice ?? 'Report created.',
                                      style: const TextStyle(
                                        fontWeight: FontWeight.w800,
                                      ),
                                    ),
                                  ),
                                ],
                              ),
                              const SizedBox(height: 8),
                              SelectableText(
                                state.lastReport!.path,
                                style: Theme.of(context).textTheme.bodySmall,
                              ),
                            ],
                          );
                          final actions = Wrap(
                            spacing: 10,
                            runSpacing: 10,
                            children: [
                              FilledButton.icon(
                                onPressed: state.openLastReport,
                                icon: const Icon(Icons.open_in_new),
                                label: const Text('Open file'),
                              ),
                              OutlinedButton.icon(
                                onPressed: state.openReportsFolder,
                                icon: const Icon(Icons.folder_open),
                                label: const Text('Open folder'),
                              ),
                              TextButton.icon(
                                onPressed: state.openLogsFolder,
                                icon: const Icon(Icons.storage_outlined),
                                label: const Text('Raw logs'),
                              ),
                            ],
                          );
                          if (constraints.maxWidth < 720) {
                            return Column(
                              crossAxisAlignment: CrossAxisAlignment.stretch,
                              children: [copy, const SizedBox(height: 16), actions],
                            );
                          }
                          return Row(
                            children: [
                              Expanded(child: copy),
                              const SizedBox(width: 18),
                              actions,
                            ],
                          );
                        },
                      ),
                    ),
        ),
        const SizedBox(height: 12),
        Text(
          'Reports are generated locally. ISP Evidence Report includes a Print / Save PDF button and does not upload measurements anywhere.',
          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: Theme.of(context).colorScheme.onSurfaceVariant,
              ),
        ),
      ],
    );
  }
}

class _ReportCard extends StatelessWidget {
  const _ReportCard({
    required this.width,
    required this.title,
    required this.description,
    required this.icon,
    required this.selector,
    required this.buttonText,
    required this.busy,
    required this.onPressed,
  });

  final double width;
  final String title;
  final String description;
  final IconData icon;
  final Widget selector;
  final String buttonText;
  final bool busy;
  final VoidCallback onPressed;

  @override
  Widget build(BuildContext context) => SizedBox(
        width: width,
        child: Panel(
          child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Align(
                  alignment: Alignment.centerLeft,
                  child: DecoratedBox(
                    decoration: BoxDecoration(
                      color: Theme.of(context)
                          .colorScheme
                          .primary
                          .withValues(alpha: .12),
                      borderRadius: BorderRadius.circular(14),
                    ),
                    child: Padding(
                      padding: const EdgeInsets.all(13),
                      child: Icon(
                        icon,
                        size: 30,
                        color: Theme.of(context).colorScheme.primary,
                      ),
                    ),
                  ),
                ),
                const SizedBox(height: 18),
                Text(
                  title,
                  style: Theme.of(context)
                      .textTheme
                      .titleMedium
                      ?.copyWith(fontWeight: FontWeight.w800),
                ),
                const SizedBox(height: 8),
                Text(description),
                const SizedBox(height: 20),
                selector,
                const SizedBox(height: 14),
                FilledButton.icon(
                  onPressed: busy ? null : onPressed,
                  icon: const Icon(Icons.auto_awesome),
                  label: Text(buttonText),
                ),
            ],
          ),
        ),
      );
}

class TargetsPage extends StatefulWidget {
  const TargetsPage({super.key, required this.state});
  final AppState state;

  @override
  State<TargetsPage> createState() => _TargetsPageState();
}

class _TargetsPageState extends State<TargetsPage> {
  final controller = TextEditingController();

  @override
  void dispose() {
    controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => ListView(
        padding: const EdgeInsets.all(24),
        children: [
          const _PageHeader(
            eyebrow: 'ENDPOINTS',
            title: 'Targets',
            subtitle: 'Add Ping, tcp://, http:// or https:// targets.',
          ),
          const SizedBox(height: 18),
          Panel(
            child: Column(
              children: [
                TextField(
                  controller: controller,
                  decoration: const InputDecoration(
                    labelText: 'Target',
                    hintText: '1.1.1.1 or tcp://example.com:443',
                  ),
                  onSubmitted: (value) async {
                    await widget.state.addTarget(value);
                    controller.clear();
                  },
                ),
                const SizedBox(height: 12),
                Align(
                  alignment: Alignment.centerRight,
                  child: FilledButton.icon(
                    onPressed: () async {
                      await widget.state.addTarget(controller.text);
                      controller.clear();
                    },
                    icon: const Icon(Icons.add),
                    label: const Text('Add target'),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 16),
          Panel(
            child: Column(
              children: [
                if (widget.state.config.customTargets.isEmpty)
                  const Padding(
                    padding: EdgeInsets.all(24),
                    child: Text('No custom targets.'),
                  ),
                for (final target in widget.state.config.customTargets)
                  ListTile(
                    contentPadding: EdgeInsets.zero,
                    title: Text(target, overflow: TextOverflow.visible),
                    trailing: IconButton(
                      tooltip: 'Remove target',
                      onPressed: () => widget.state.removeTarget(target),
                      icon: const Icon(Icons.delete_outline),
                    ),
                  ),
              ],
            ),
          ),
        ],
      );
}

class SettingsPage extends StatefulWidget {
  const SettingsPage({super.key, required this.state});
  final AppState state;

  @override
  State<SettingsPage> createState() => _SettingsPageState();
}

class _SettingsPageState extends State<SettingsPage> {
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
          const _PageHeader(
            eyebrow: 'PREFERENCES',
            title: 'Settings',
            subtitle: 'All fields use flexible layouts and can wrap safely.',
          ),
          const SizedBox(height: 18),
          Panel(
            child: Column(
              children: [
                _SettingRow(
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
                _SettingRow(
                  label: 'Monitoring interval',
                  child: TextFormField(
                    initialValue: draft.intervalSeconds.toString(),
                    keyboardType: TextInputType.number,
                    onChanged: (value) => draft = draft.copyWith(
                      intervalSeconds:
                          double.tryParse(value) ?? draft.intervalSeconds,
                    ),
                  ),
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
                    () => draft =
                        draft.copyWith(showOutageNotifications: value),
                  ),
                ),
                const SizedBox(height: 14),
                Align(
                  alignment: Alignment.centerRight,
                  child: FilledButton(
                    onPressed: () => widget.state.saveConfig(draft),
                    child: const Text('Save settings'),
                  ),
                ),
              ],
            ),
          ),
        ],
      );
}

class _SettingRow extends StatelessWidget {
  const _SettingRow({required this.label, required this.child});
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
                  Text(label, style: const TextStyle(fontWeight: FontWeight.w700)),
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

class _PageHeader extends StatelessWidget {
  const _PageHeader({
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
