import 'package:flutter/material.dart';

import 'app_state.dart';
import 'models.dart';
import 'network_map.dart';
import 'widgets.dart';

class ConsoleDashboardPage extends StatelessWidget {
  const ConsoleDashboardPage({
    super.key,
    required this.state,
    required this.onViewTargets,
  });

  final AppState state;
  final VoidCallback onViewTargets;

  @override
  Widget build(BuildContext context) {
    final snapshot = state.snapshot;
    final live = snapshot.monitoring && snapshot.samples > 0;
    return ListView(
      padding: const EdgeInsets.fromLTRB(28, 27, 28, 36),
      children: [
        LayoutBuilder(builder: (context, constraints) {
          final heading = Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const _Eyebrow('OVERVIEW  /  NETWORK'),
              const SizedBox(height: 6),
              Text('Your network, mapped.',
                  style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                        fontWeight: FontWeight.w900,
                        letterSpacing: -.7,
                      )),
              const SizedBox(height: 4),
              Text('Every connection in view, as it happens.',
                  style: TextStyle(
                    color: Theme.of(context).colorScheme.onSurfaceVariant,
                  )),
            ],
          );
          final action = FilledButton.icon(
            onPressed: state.toggleMonitoring,
            icon: Icon(snapshot.monitoring
                ? Icons.stop_rounded
                : Icons.play_arrow_rounded),
            label: Text(
                snapshot.monitoring ? 'Stop monitoring' : 'Start monitoring'),
          );
          if (constraints.maxWidth < 640) {
            return Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [heading, const SizedBox(height: 16), action],
            );
          }
          return Row(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              Expanded(child: heading),
              const SizedBox(width: 16),
              action,
            ],
          );
        }),
        const SizedBox(height: 20),
        NetworkMap(snapshot: snapshot),
        const SizedBox(height: 13),
        LayoutBuilder(builder: (context, constraints) {
          final columns = constraints.maxWidth >= 850 ? 4 : 2;
          // Leave room for the enclosing border so the final cell never wraps.
          final width = (constraints.maxWidth - 4) / columns;
          return Container(
            decoration: BoxDecoration(
              color: Theme.of(context).colorScheme.surface,
              border: Border.all(color: Theme.of(context).colorScheme.outline),
              borderRadius: BorderRadius.circular(5),
            ),
            child: Wrap(
              children: [
                _MetricCell(
                  width: width,
                  index: '01',
                  label: 'AVG LATENCY',
                  value:
                      live ? snapshot.averageLatency.toStringAsFixed(1) : '—',
                  unit: live ? 'ms' : '',
                ),
                _MetricCell(
                  width: width,
                  index: '02',
                  label: 'PACKET LOSS',
                  value: live ? snapshot.packetLoss.toStringAsFixed(1) : '—',
                  unit: live ? '%' : '',
                ),
                _MetricCell(
                  width: width,
                  index: '03',
                  label: 'JITTER',
                  value: live ? snapshot.jitter.toStringAsFixed(1) : '—',
                  unit: live ? 'ms' : '',
                ),
                _MetricCell(
                  width: width,
                  index: '04',
                  label: 'SAMPLES',
                  value: live ? snapshot.samples.toString() : '—',
                ),
              ],
            ),
          );
        }),
        const SizedBox(height: 13),
        LayoutBuilder(builder: (context, constraints) {
          final chart = Panel(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                LayoutBuilder(builder: (context, headerWidth) {
                  final heading = Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const _Eyebrow('TELEMETRY  /  LIVE TRACE'),
                      const SizedBox(height: 8),
                      Text('Latency history',
                          style:
                              Theme.of(context).textTheme.titleLarge?.copyWith(
                                    fontWeight: FontWeight.w900,
                                  )),
                      const SizedBox(height: 3),
                      Text('Measured time, shown at a readable scale',
                          style:
                              Theme.of(context).textTheme.bodySmall?.copyWith(
                                    color: Theme.of(context)
                                        .colorScheme
                                        .onSurfaceVariant,
                                  )),
                    ],
                  );
                  final range = SizedBox(
                    width: 166,
                    child: DropdownButtonFormField<int>(
                      key: ValueKey<int>(state.config.graphRangeMinutes),
                      initialValue: state.config.graphRangeMinutes,
                      isExpanded: true,
                      decoration: const InputDecoration(
                        labelText: 'History range',
                        isDense: true,
                      ),
                      items: const [
                        DropdownMenuItem(
                            value: 5, child: Text('Last 5 minutes')),
                        DropdownMenuItem(
                            value: 30, child: Text('Last 30 minutes')),
                        DropdownMenuItem(value: 60, child: Text('Last hour')),
                        DropdownMenuItem(
                            value: 1440, child: Text('Last 24 hours')),
                      ],
                      onChanged: (value) {
                        if (value != null) state.setGraphRange(value);
                      },
                    ),
                  );
                  if (headerWidth.maxWidth < 530) {
                    return Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [heading, const SizedBox(height: 14), range],
                    );
                  }
                  return Row(
                    children: [
                      Expanded(child: heading),
                      const SizedBox(width: 16),
                      range,
                    ],
                  );
                }),
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
                const _Eyebrow('ENDPOINTS  /  03'),
                const SizedBox(height: 11),
                Row(
                  children: [
                    Expanded(
                      child: Text('Monitored targets',
                          style:
                              Theme.of(context).textTheme.titleLarge?.copyWith(
                                    fontWeight: FontWeight.w900,
                                  )),
                    ),
                    Text('${snapshot.targets.length} TOTAL',
                        style: Theme.of(context).textTheme.labelSmall?.copyWith(
                              fontFamily: 'Consolas',
                              color: Theme.of(context)
                                  .colorScheme
                                  .onSurfaceVariant,
                            )),
                  ],
                ),
                const SizedBox(height: 12),
                if (snapshot.targets.isEmpty)
                  const Padding(
                    padding: EdgeInsets.symmetric(vertical: 28),
                    child: Text('Targets will appear when monitoring starts.'),
                  )
                else
                  for (var i = 0; i < snapshot.targets.length && i < 6; i++)
                    _EndpointRow(
                      number: i + 1,
                      status: snapshot.targets[i],
                      live: live,
                    ),
                if (snapshot.targets.length > 6)
                  Align(
                    alignment: Alignment.centerLeft,
                    child: TextButton.icon(
                      onPressed: onViewTargets,
                      icon: const Icon(Icons.arrow_forward_rounded, size: 17),
                      label:
                          Text('View all ${snapshot.targets.length} targets'),
                    ),
                  ),
              ],
            ),
          );
          if (constraints.maxWidth < 890) {
            return Column(
                children: [chart, const SizedBox(height: 13), targets]);
          }
          return Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Expanded(flex: 3, child: chart),
              const SizedBox(width: 13),
              Expanded(flex: 2, child: targets),
            ],
          );
        }),
        const SizedBox(height: 13),
        Panel(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const _Eyebrow('ACTIVITY  /  04'),
              const SizedBox(height: 9),
              Text('Recent events',
                  style: Theme.of(context).textTheme.titleLarge?.copyWith(
                        fontWeight: FontWeight.w900,
                      )),
              const SizedBox(height: 10),
              if (snapshot.recentEvents.isEmpty)
                Text('No events yet.',
                    style: TextStyle(
                        color: Theme.of(context).colorScheme.onSurfaceVariant))
              else
                for (final event in snapshot.recentEvents)
                  ListTile(
                    contentPadding: EdgeInsets.zero,
                    leading: const Icon(Icons.arrow_outward_rounded, size: 17),
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

class _Eyebrow extends StatelessWidget {
  const _Eyebrow(this.text);
  final String text;

  @override
  Widget build(BuildContext context) => Text(
        text,
        style: Theme.of(context).textTheme.labelSmall?.copyWith(
              fontFamily: 'Consolas',
              letterSpacing: 1.8,
              fontWeight: FontWeight.w700,
              color: Theme.of(context).colorScheme.onSurfaceVariant,
            ),
      );
}

class _MetricCell extends StatelessWidget {
  const _MetricCell({
    required this.width,
    required this.index,
    required this.label,
    required this.value,
    this.unit = '',
  });

  final double width;
  final String index;
  final String label;
  final String value;
  final String unit;

  @override
  Widget build(BuildContext context) => SizedBox(
        width: width,
        child: Container(
          padding: const EdgeInsets.fromLTRB(18, 15, 18, 18),
          decoration: BoxDecoration(
            border: Border(
              right: BorderSide(color: Theme.of(context).colorScheme.outline),
            ),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('$index  /  $label',
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: Theme.of(context).textTheme.labelSmall?.copyWith(
                        fontFamily: 'Consolas',
                        color: Theme.of(context).colorScheme.onSurfaceVariant,
                        letterSpacing: 1,
                      )),
              const SizedBox(height: 8),
              Text.rich(TextSpan(
                text: value,
                style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                      fontFamily: 'Consolas',
                      fontWeight: FontWeight.w700,
                    ),
                children: [
                  if (unit.isNotEmpty)
                    TextSpan(
                      text: ' $unit',
                      style: Theme.of(context).textTheme.labelMedium,
                    ),
                ],
              )),
            ],
          ),
        ),
      );
}

class _EndpointRow extends StatelessWidget {
  const _EndpointRow(
      {required this.number, required this.status, required this.live});
  final int number;
  final TargetStatus status;
  final bool live;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    final color = !live
        ? scheme.onSurfaceVariant
        : status.state == 'online'
            ? const Color(0xFF5DAF70)
            : status.state == 'offline'
                ? scheme.error
                : scheme.onSurfaceVariant;
    return Container(
      padding: const EdgeInsets.symmetric(vertical: 13),
      decoration: BoxDecoration(
        border: Border(top: BorderSide(color: scheme.outline)),
      ),
      child: Row(
        children: [
          Text(number.toString().padLeft(2, '0'),
              style: TextStyle(
                fontFamily: 'Consolas',
                fontSize: 12,
                color: scheme.onSurfaceVariant,
              )),
          const SizedBox(width: 14),
          Container(
              width: 7,
              height: 7,
              decoration: BoxDecoration(color: color, shape: BoxShape.circle)),
          const SizedBox(width: 11),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(status.target.name,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(fontWeight: FontWeight.w800)),
                Text(status.target.host,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: TextStyle(
                        fontSize: 12, color: scheme.onSurfaceVariant)),
              ],
            ),
          ),
          const SizedBox(width: 10),
          Text(
              live && status.state == 'online'
                  ? '${status.latency.toStringAsFixed(1)} ms'
                  : live && status.state == 'offline'
                      ? 'OFFLINE'
                      : '—',
              style: TextStyle(
                fontFamily: 'Consolas',
                fontWeight: FontWeight.w700,
                color: color,
              )),
        ],
      ),
    );
  }
}
