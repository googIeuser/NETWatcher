import 'package:flutter/material.dart';

import 'app_state.dart';
import 'models.dart';
import 'motion.dart';
import 'widgets.dart';

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
            padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 4),
            child: Column(
              children: [
                for (var index = 0;
                    index < state.snapshot.targets.length;
                    index++)
                  TargetCard(
                    status: state.snapshot.targets[index],
                    showTopDivider: index > 0,
                  ),
              ],
            ),
          ),
        ],
      );
}

class OutagesPage extends StatelessWidget {
  const OutagesPage({super.key, required this.state});
  final AppState state;

  Future<void> _confirmClearHistory(BuildContext context) async {
    final scheme = Theme.of(context).colorScheme;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: const Text('Delete outage history?'),
        content: const Text(
          'This permanently deletes all saved outage records. '
          'An outage currently in progress will remain visible.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(dialogContext).pop(false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            style: FilledButton.styleFrom(
              backgroundColor: scheme.error,
              foregroundColor: scheme.onError,
            ),
            onPressed: () => Navigator.of(dialogContext).pop(true),
            child: const Text('Delete'),
          ),
        ],
      ),
    );

    if (confirmed != true || !context.mounted) return;
    final deleted = await state.clearOutageHistory();
    if (!context.mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(
          deleted
              ? 'Outage history deleted.'
              : state.error ?? 'Outage history could not be deleted.',
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final incidents = state.outages;
    final activeCount = incidents.where((item) => item.active).length;
    final totalSeconds = incidents.fold<double>(
      0,
      (sum, item) => sum + item.durationSeconds,
    );
    final longestSeconds = incidents.fold<double>(
      0,
      (longest, item) =>
          item.durationSeconds > longest ? item.durationSeconds : longest,
    );

    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        const _PageHeader(
          eyebrow: 'HISTORY',
          title: 'Outage history',
          subtitle:
              'Review each confirmed incident with its type, start and end time, duration and diagnostic details.',
        ),
        const SizedBox(height: 16),
        LayoutBuilder(
          builder: (context, constraints) {
            final narrow = constraints.maxWidth < 520;
            final range = SizedBox(
              width: narrow ? constraints.maxWidth : 270,
              child: DropdownButtonFormField<int>(
                key: ValueKey<int>(state.outageRangeDays),
                initialValue: state.outageRangeDays,
                isExpanded: true,
                decoration: const InputDecoration(labelText: 'History range'),
                items: const [
                  DropdownMenuItem(value: 1, child: Text('Last 24 hours')),
                  DropdownMenuItem(value: 7, child: Text('Last 7 days')),
                  DropdownMenuItem(value: 30, child: Text('Last 30 days')),
                  DropdownMenuItem(value: 365, child: Text('Last year')),
                  DropdownMenuItem(value: 36500, child: Text('All time')),
                ],
                onChanged: state.outagesLoading
                    ? null
                    : (value) => state.refreshOutages(value ?? 30),
              ),
            );
            final buttons = Wrap(
              spacing: 10,
              runSpacing: 10,
              children: [
                IconButton.filledTonal(
                  tooltip: 'Refresh outage history',
                  onPressed: state.outagesLoading
                      ? null
                      : () => state.refreshOutages(),
                  icon: const Icon(Icons.refresh),
                ),
                IconButton.filledTonal(
                  key: const ValueKey<String>('delete-outage-history'),
                  tooltip: 'Delete outage history',
                  style: IconButton.styleFrom(
                    foregroundColor: Theme.of(context).colorScheme.error,
                  ),
                  onPressed: state.outagesLoading
                      ? null
                      : () => _confirmClearHistory(context),
                  icon: const Icon(Icons.delete_outline_rounded),
                ),
              ],
            );

            return Align(
              alignment: Alignment.centerRight,
              child: Wrap(
                key: const ValueKey<String>('outage-history-actions'),
                alignment: WrapAlignment.end,
                crossAxisAlignment: WrapCrossAlignment.center,
                spacing: 10,
                runSpacing: 10,
                children: [range, buttons],
              ),
            );
          },
        ),
        const SizedBox(height: 18),
        if (incidents.isNotEmpty) ...[
          LayoutBuilder(
            builder: (context, constraints) {
              final columns = constraints.maxWidth >= 760
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
                    child: _OutageSummaryCard(
                      label: 'Incidents',
                      value: incidents.length.toString(),
                      icon: Icons.warning_amber_rounded,
                    ),
                  ),
                  SizedBox(
                    width: width,
                    child: _OutageSummaryCard(
                      label: 'Active now',
                      value: activeCount.toString(),
                      icon: Icons.bolt_rounded,
                    ),
                  ),
                  SizedBox(
                    width: width,
                    child: _OutageSummaryCard(
                      label: 'Total downtime',
                      value: _formatDuration(totalSeconds),
                      icon: Icons.timer_outlined,
                    ),
                  ),
                  SizedBox(
                    width: width,
                    child: _OutageSummaryCard(
                      label: 'Longest incident',
                      value: _formatDuration(longestSeconds),
                      icon: Icons.timeline_rounded,
                    ),
                  ),
                ],
              );
            },
          ),
          const SizedBox(height: 16),
        ],
        AnimatedSwitcher(
          duration: NetWatcherMotion.normal,
          child: state.outagesLoading && incidents.isEmpty
              ? const Panel(
                  key: ValueKey<String>('outages-loading'),
                  child: Center(
                    child: Padding(
                      padding: EdgeInsets.symmetric(vertical: 42),
                      child: CircularProgressIndicator(),
                    ),
                  ),
                )
              : incidents.isEmpty
                  ? const Panel(
                      key: ValueKey<String>('outages-empty'),
                      child: Center(
                        child: Padding(
                          padding: EdgeInsets.symmetric(vertical: 34),
                          child: Column(
                            children: [
                              Icon(Icons.verified_outlined, size: 36),
                              SizedBox(height: 12),
                              Text('No confirmed outages in this range.'),
                              SizedBox(height: 5),
                              Text(
                                  'Connection incidents will appear here as they happen.'),
                            ],
                          ),
                        ),
                      ),
                    )
                  : Column(
                      key: ValueKey<String>(
                        'outages-${state.outageRangeDays}-${incidents.length}',
                      ),
                      children: [
                        for (final incident in incidents) ...[
                          OutageIncidentCard(incident: incident),
                          const SizedBox(height: 12),
                        ],
                      ],
                    ),
        ),
      ],
    );
  }
}

class _OutageSummaryCard extends StatelessWidget {
  const _OutageSummaryCard({
    required this.label,
    required this.value,
    required this.icon,
  });

  final String label;
  final String value;
  final IconData icon;

  @override
  Widget build(BuildContext context) => Panel(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Icon(icon, color: Theme.of(context).colorScheme.primary),
            const SizedBox(height: 8),
            Text(
              label,
              style: Theme.of(context).textTheme.labelMedium?.copyWith(
                    color: Theme.of(context).colorScheme.onSurfaceVariant,
                  ),
            ),
            const SizedBox(height: 4),
            Text(
              value,
              overflow: TextOverflow.visible,
              style: Theme.of(context).textTheme.titleLarge?.copyWith(
                    fontWeight: FontWeight.w900,
                  ),
            ),
          ],
        ),
      );
}

class OutageIncidentCard extends StatelessWidget {
  const OutageIncidentCard({
    super.key,
    required this.incident,
  });

  final OutageRecord incident;

  @override
  Widget build(BuildContext context) {
    final appearance = _outageAppearance(incident.category);
    final statusColor =
        incident.active ? const Color(0xFFFF6D80) : const Color(0xFF42D99A);

    return Panel(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Wrap(
            spacing: 13,
            runSpacing: 12,
            crossAxisAlignment: WrapCrossAlignment.center,
            children: [
              Container(
                width: 44,
                height: 44,
                decoration: BoxDecoration(
                  color: appearance.color.withValues(alpha: .13),
                  borderRadius: BorderRadius.circular(13),
                ),
                child: Icon(appearance.icon, color: appearance.color),
              ),
              Text(
                appearance.label,
                style: Theme.of(context).textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.w900,
                    ),
              ),
              DecoratedBox(
                decoration: BoxDecoration(
                  color: statusColor.withValues(alpha: .13),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Padding(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 9,
                    vertical: 5,
                  ),
                  child: Text(
                    incident.active ? 'ACTIVE' : 'RESOLVED',
                    style: TextStyle(
                      color: statusColor,
                      fontSize: 10,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 9),
          Text(
            incident.details.isEmpty
                ? 'No diagnostic description was recorded.'
                : incident.details,
            overflow: TextOverflow.visible,
            style: TextStyle(
              color: Theme.of(context).colorScheme.onSurfaceVariant,
            ),
          ),
          const SizedBox(height: 18),
          Wrap(
            spacing: 22,
            runSpacing: 12,
            children: [
              _IncidentValue(
                label: 'Started',
                value: _formatDateTime(incident.start),
              ),
              _IncidentValue(
                label: incident.active ? 'Status' : 'Ended',
                value: incident.active
                    ? 'Still in progress'
                    : _formatDateTime(incident.end),
              ),
              _IncidentValue(
                label: 'Duration',
                value: _formatDuration(incident.durationSeconds),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _IncidentValue extends StatelessWidget {
  const _IncidentValue({
    required this.label,
    required this.value,
  });

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) => ConstrainedBox(
        constraints: const BoxConstraints(minWidth: 120),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              label,
              style: Theme.of(context).textTheme.labelSmall?.copyWith(
                    color: Theme.of(context).colorScheme.onSurfaceVariant,
                  ),
            ),
            const SizedBox(height: 5),
            Text(
              value,
              overflow: TextOverflow.visible,
              style: const TextStyle(fontWeight: FontWeight.w800),
            ),
          ],
        ),
      );
}

({String label, IconData icon, Color color}) _outageAppearance(
  String category,
) {
  return switch (category.toLowerCase()) {
    'local' => (
        label: 'Local network failure',
        icon: Icons.router_outlined,
        color: const Color(0xFFFF6D80),
      ),
    'partial' => (
        label: 'Partial access',
        icon: Icons.call_split_rounded,
        color: const Color(0xFFFFBD59),
      ),
    'degraded' => (
        label: 'High latency',
        icon: Icons.speed_rounded,
        color: const Color(0xFFFFBD59),
      ),
    _ => (
        label: 'Internet outage',
        icon: Icons.cloud_off_outlined,
        color: const Color(0xFFFF6D80),
      ),
  };
}

String _formatDateTime(String value) {
  final parsed = DateTime.tryParse(value)?.toLocal();
  if (parsed == null) return value.isEmpty ? '—' : value;
  String two(int number) => number.toString().padLeft(2, '0');
  return '${two(parsed.day)}.${two(parsed.month)}.${parsed.year} '
      '${two(parsed.hour)}:${two(parsed.minute)}:${two(parsed.second)}';
}

String _formatDuration(double seconds) {
  final total = seconds.round().clamp(0, 315360000).toInt();
  final days = total ~/ 86400;
  final hours = (total % 86400) ~/ 3600;
  final minutes = (total % 3600) ~/ 60;
  final remainingSeconds = total % 60;
  if (days > 0) return '${days}d ${hours}h ${minutes}m';
  if (hours > 0) return '${hours}h ${minutes}m';
  if (minutes > 0) return '${minutes}m ${remainingSeconds}s';
  return '${remainingSeconds}s';
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
        Column(
          children: [
            _ReportCard(
              title: 'HTML report',
              description:
                  'Connection measurements, target summaries and completed outage events in a printable page.',
              icon: Icons.description_outlined,
              busy: state.reportBusy,
              selector: DropdownButtonFormField<int>(
                key: ValueKey<int>(htmlHours),
                initialValue: htmlHours,
                isExpanded: true,
                decoration:
                    const InputDecoration(labelText: 'Measurement range'),
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
              buttonText: 'Create HTML report',
              onPressed: () => state.generateHtmlReport(htmlHours),
            ),
            const SizedBox(height: 12),
            _ReportCard(
              title: 'ISP Evidence Report',
              description:
                  'Availability, packet loss, latency, jitter and outage evidence formatted for an ISP or regulator.',
              icon: Icons.fact_check_outlined,
              busy: state.reportBusy,
              selector: DropdownButtonFormField<int>(
                key: ValueKey<int>(evidenceDays),
                initialValue: evidenceDays,
                isExpanded: true,
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
              buttonText: 'Create evidence report',
              onPressed: () => state.generateEvidenceReport(evidenceDays),
            ),
            const SizedBox(height: 12),
            _ReportCard(
              title: 'Diagnostics ZIP',
              description:
                  'Exports settings, snapshot, calculated statistics, outages and the original local CSV logs.',
              icon: Icons.archive_outlined,
              busy: state.reportBusy,
              selector: DropdownButtonFormField<int>(
                key: ValueKey<int>(diagnosticsHours),
                initialValue: diagnosticsHours,
                isExpanded: true,
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
                      Expanded(
                          child:
                              Text('Preparing the report from local data...')),
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
                              children: [
                                copy,
                                const SizedBox(height: 16),
                                actions
                              ],
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
    required this.title,
    required this.description,
    required this.icon,
    required this.selector,
    required this.buttonText,
    required this.busy,
    required this.onPressed,
  });

  final String title;
  final String description;
  final IconData icon;
  final Widget selector;
  final String buttonText;
  final bool busy;
  final VoidCallback onPressed;

  @override
  Widget build(BuildContext context) => Panel(
        padding: const EdgeInsets.all(16),
        child: LayoutBuilder(
          builder: (context, constraints) {
            final compact = constraints.maxWidth < 700;
            final identity = Row(
              children: [
                Container(
                  width: 42,
                  height: 42,
                  decoration: BoxDecoration(
                    color: Theme.of(context)
                        .colorScheme
                        .primary
                        .withValues(alpha: .10),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Icon(
                    icon,
                    size: 22,
                    color: Theme.of(context).colorScheme.primary,
                  ),
                ),
                const SizedBox(width: 14),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        title,
                        style:
                            Theme.of(context).textTheme.titleMedium?.copyWith(
                                  fontWeight: FontWeight.w800,
                                ),
                      ),
                      const SizedBox(height: 3),
                      Text(
                        description,
                        maxLines: compact ? 3 : 2,
                        overflow: TextOverflow.ellipsis,
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                              color: Theme.of(context)
                                  .colorScheme
                                  .onSurfaceVariant,
                            ),
                      ),
                    ],
                  ),
                ),
              ],
            );
            final action = FilledButton.icon(
              onPressed: busy ? null : onPressed,
              icon: const Icon(Icons.arrow_forward, size: 18),
              label: Text(buttonText),
            );
            if (compact) {
              return Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  identity,
                  const SizedBox(height: 14),
                  selector,
                  const SizedBox(height: 10),
                  action,
                ],
              );
            }
            return Row(
              children: [
                Expanded(child: identity),
                const SizedBox(width: 20),
                SizedBox(width: 174, child: selector),
                const SizedBox(width: 12),
                SizedBox(width: 258, child: action),
              ],
            );
          },
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
            child: LayoutBuilder(
              builder: (context, constraints) {
                final input = TextField(
                  controller: controller,
                  decoration: const InputDecoration(
                    labelText: 'Target',
                    hintText: '1.1.1.1 or tcp://example.com:443',
                  ),
                  onSubmitted: (value) async {
                    await widget.state.addTarget(value);
                    controller.clear();
                  },
                );
                final add = FilledButton.icon(
                  onPressed: () async {
                    await widget.state.addTarget(controller.text);
                    controller.clear();
                  },
                  icon: const Icon(Icons.add),
                  label: const Text('Add target'),
                );
                if (constraints.maxWidth < 640) {
                  return Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      input,
                      const SizedBox(height: 12),
                      add,
                    ],
                  );
                }
                return Row(
                  children: [
                    Expanded(child: input),
                    const SizedBox(width: 12),
                    add,
                  ],
                );
              },
            ),
          ),
          const SizedBox(height: 16),
          Panel(
            child: Column(
              children: [
                if (widget.state.config.customTargets.isEmpty)
                  const Padding(
                    padding: EdgeInsets.symmetric(vertical: 20),
                    child: Column(
                      children: [
                        Icon(Icons.radar_outlined, size: 30),
                        SizedBox(height: 10),
                        Text('No custom targets yet.'),
                        SizedBox(height: 4),
                        Text(
                            'Default gateway, Cloudflare and Google are already included.'),
                      ],
                    ),
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
  Widget build(BuildContext context) => Container(
        padding: const EdgeInsets.only(bottom: 18),
        decoration: BoxDecoration(
          border:
              Border(bottom: BorderSide(color: Theme.of(context).dividerColor)),
        ),
        child: LayoutBuilder(builder: (context, constraints) {
          final copy = Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                eyebrow,
                style: Theme.of(context).textTheme.labelSmall?.copyWith(
                      color: Theme.of(context).colorScheme.onSurfaceVariant,
                      fontFamily: 'Consolas',
                      fontWeight: FontWeight.w800,
                      letterSpacing: 1.8,
                    ),
              ),
              const SizedBox(height: 7),
              Text(
                title,
                overflow: TextOverflow.visible,
                style: Theme.of(context)
                    .textTheme
                    .headlineMedium
                    ?.copyWith(fontWeight: FontWeight.w900, letterSpacing: -.6),
              ),
              const SizedBox(height: 4),
              Text(subtitle,
                  overflow: TextOverflow.visible,
                  style: TextStyle(
                      color: Theme.of(context).colorScheme.onSurfaceVariant)),
            ],
          );
          if (trailing == null) return copy;
          if (constraints.maxWidth < 760) {
            return Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [copy, const SizedBox(height: 16), trailing!],
            );
          }
          return Row(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              Expanded(flex: 3, child: copy),
              const SizedBox(width: 20),
              Flexible(
                flex: 2,
                child: Align(
                  alignment: Alignment.bottomRight,
                  child: trailing!,
                ),
              ),
            ],
          );
        }),
      );
}
