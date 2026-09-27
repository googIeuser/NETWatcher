import 'package:flutter/material.dart';

import 'app_state.dart';
import 'models.dart';
import 'theme.dart';
import 'widgets.dart';

class ConsoleDashboardPage extends StatelessWidget {
  const ConsoleDashboardPage({super.key, required this.state});

  final AppState state;

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
              const _Eyebrow('CONTROL DESK  /  01'),
              const SizedBox(height: 6),
              Text('Your network, at a glance.',
                  style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                        fontWeight: FontWeight.w900,
                        letterSpacing: -.7,
                      )),
              const SizedBox(height: 4),
              Text('Live signal, history and endpoints in one view.',
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
        _SignalHero(snapshot: snapshot, live: live),
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
                const _Eyebrow('TELEMETRY  /  02'),
                const SizedBox(height: 11),
                Wrap(
                  spacing: 12,
                  runSpacing: 12,
                  crossAxisAlignment: WrapCrossAlignment.center,
                  children: [
                    Text('Latency history',
                        style: Theme.of(context).textTheme.titleLarge?.copyWith(
                              fontWeight: FontWeight.w900,
                            )),
                    SizedBox(
                      width: 174,
                      child: DropdownButtonFormField<int>(
                        key: ValueKey<int>(state.config.graphRangeMinutes),
                        initialValue: state.config.graphRangeMinutes,
                        isExpanded: true,
                        decoration:
                            const InputDecoration(labelText: 'History range'),
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
                  for (var i = 0; i < snapshot.targets.length; i++)
                    _EndpointRow(
                      number: i + 1,
                      status: snapshot.targets[i],
                      live: live,
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

class _SignalHero extends StatelessWidget {
  const _SignalHero({required this.snapshot, required this.live});
  final NetworkSnapshot snapshot;
  final bool live;

  @override
  Widget build(BuildContext context) {
    final status = !snapshot.monitoring
        ? 'Monitoring paused'
        : !live
            ? 'Checking connection'
            : snapshot.connectionLabel;
    final detail = !snapshot.monitoring
        ? 'Start monitoring to begin collecting connection data.'
        : !live
            ? 'Waiting for the first measurements.'
            : 'Live measurements are arriving from your targets.';
    return Container(
      padding: const EdgeInsets.fromLTRB(24, 20, 24, 22),
      decoration: BoxDecoration(
        color: NetWatcherTheme.ink,
        borderRadius: BorderRadius.circular(5),
        border: Border.all(color: const Color(0xFF41664B)),
      ),
      child: LayoutBuilder(builder: (context, constraints) {
        final copy = Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('LIVE SIGNAL   •   NETWORK HEALTH',
                style: TextStyle(
                  color: NetWatcherTheme.signal,
                  fontFamily: 'Consolas',
                  fontSize: 11,
                  letterSpacing: 1.6,
                  fontWeight: FontWeight.w700,
                )),
            const SizedBox(height: 20),
            Text(status,
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 30,
                  fontWeight: FontWeight.w900,
                  letterSpacing: -.5,
                )),
            const SizedBox(height: 6),
            Text(detail, style: const TextStyle(color: Color(0xFFC6D3C4))),
          ],
        );
        final score = Container(
          key: live ? null : const ValueKey<String>('dashboard-empty-state'),
          constraints: const BoxConstraints(minWidth: 130),
          padding: const EdgeInsets.only(left: 18),
          decoration: const BoxDecoration(
            border: Border(left: BorderSide(color: Color(0xFF41664B))),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text('QUALITY INDEX',
                  style: TextStyle(
                    color: Color(0xFFB7C9B6),
                    fontFamily: 'Consolas',
                    fontSize: 10,
                    letterSpacing: 1.4,
                  )),
              const SizedBox(height: 3),
              Text(live ? '${snapshot.qualityScore}' : '—',
                  style: const TextStyle(
                    color: NetWatcherTheme.signal,
                    fontFamily: 'Consolas',
                    fontSize: 43,
                    height: 1.1,
                    fontWeight: FontWeight.bold,
                  )),
              Text(live ? '/ 100' : 'NO LIVE DATA',
                  style: const TextStyle(
                    color: Color(0xFFB7C9B6),
                    fontFamily: 'Consolas',
                    fontSize: 11,
                  )),
            ],
          ),
        );
        if (constraints.maxWidth < 570) {
          return Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [copy, const SizedBox(height: 22), score],
          );
        }
        return Row(
          children: [Expanded(child: copy), const SizedBox(width: 18), score],
        );
      }),
    );
  }
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
