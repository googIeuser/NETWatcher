import 'dart:math' as math;

import 'package:flutter/material.dart';

import 'models.dart';
import 'motion.dart';

class Panel extends StatelessWidget {
  const Panel({
    super.key,
    required this.child,
    this.padding = const EdgeInsets.all(20),
  });

  final Widget child;
  final EdgeInsets padding;

  @override
  Widget build(BuildContext context) =>
      Card(child: Padding(padding: padding, child: child));
}

class TargetCard extends StatelessWidget {
  const TargetCard({
    super.key,
    required this.status,
    this.showTopDivider = true,
  });

  final TargetStatus status;
  final bool showTopDivider;

  Color _stateColor(BuildContext context) {
    return switch (status.state) {
      'online' => const Color(0xFF72B984),
      'offline' => Theme.of(context).colorScheme.error,
      _ => Theme.of(context).colorScheme.onSurfaceVariant,
    };
  }

  @override
  Widget build(BuildContext context) {
    final stateColor = _stateColor(context);
    return AnimatedContainer(
      duration: NetWatcherMotion.normal,
      curve: NetWatcherMotion.curve,
      padding: const EdgeInsets.symmetric(vertical: 14),
      decoration: BoxDecoration(
        border: showTopDivider
            ? Border(
                top: BorderSide(color: Theme.of(context).dividerColor),
              )
            : null,
      ),
      child: LayoutBuilder(
        builder: (context, constraints) {
          final compact = constraints.maxWidth < 520;
          final identity = Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Padding(
                padding: const EdgeInsets.only(top: 5),
                child: AnimatedContainer(
                  duration: NetWatcherMotion.normal,
                  curve: NetWatcherMotion.curve,
                  width: 9,
                  height: 9,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    color: stateColor,
                    boxShadow: [
                      BoxShadow(
                        color: stateColor.withValues(alpha: .35),
                        blurRadius: status.state == 'online' ? 10 : 2,
                      ),
                    ],
                  ),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      status.target.name,
                      maxLines: compact ? 3 : 2,
                      overflow: TextOverflow.visible,
                      style: const TextStyle(fontWeight: FontWeight.w700),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      '${status.target.host} · ${status.target.mode.toUpperCase()}',
                      overflow: TextOverflow.visible,
                      style: Theme.of(context).textTheme.bodySmall?.copyWith(
                            color:
                                Theme.of(context).colorScheme.onSurfaceVariant,
                          ),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 10),
              ConstrainedBox(
                constraints: const BoxConstraints(minWidth: 68),
                child: AnimatedContainer(
                  duration: NetWatcherMotion.normal,
                  curve: NetWatcherMotion.curve,
                  decoration: BoxDecoration(
                    color: stateColor.withValues(alpha: .12),
                    borderRadius: BorderRadius.circular(9),
                  ),
                  child: Padding(
                    padding:
                        const EdgeInsets.symmetric(horizontal: 10, vertical: 7),
                    child: AnimatedSwitcher(
                      duration: NetWatcherMotion.normal,
                      child: Text(
                        status.state.toUpperCase(),
                        key: ValueKey<String>(status.state),
                        textAlign: TextAlign.center,
                        maxLines: 1,
                        overflow: TextOverflow.visible,
                        style: TextStyle(
                          color: stateColor,
                          fontSize: 11,
                          fontWeight: FontWeight.w800,
                        ),
                      ),
                    ),
                  ),
                ),
              ),
            ],
          );

          final metrics = Wrap(
            spacing: 18,
            runSpacing: 10,
            children: [
              _TargetMetric(
                label: 'Latency',
                value: status.state == 'waiting'
                    ? '—'
                    : '${status.latency.toStringAsFixed(1)} ms',
              ),
              _TargetMetric(
                label: 'Packet loss',
                value: status.state == 'waiting'
                    ? '—'
                    : '${status.packetLoss.toStringAsFixed(1)}%',
              ),
              _TargetMetric(
                label: 'Jitter',
                value: status.state == 'waiting'
                    ? '—'
                    : '${status.jitter.toStringAsFixed(1)} ms',
              ),
            ],
          );

          if (compact) {
            return Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                identity,
                const SizedBox(height: 14),
                Padding(
                  padding: const EdgeInsets.only(left: 21),
                  child: metrics,
                ),
              ],
            );
          }
          return Row(
            children: [
              Expanded(flex: 3, child: identity),
              const SizedBox(width: 20),
              Flexible(flex: 2, child: metrics),
            ],
          );
        },
      ),
    );
  }
}

class _TargetMetric extends StatelessWidget {
  const _TargetMetric({required this.label, required this.value});
  final String label;
  final String value;

  @override
  Widget build(BuildContext context) => ConstrainedBox(
        constraints: const BoxConstraints(minWidth: 82),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              label,
              style: Theme.of(context).textTheme.labelSmall?.copyWith(
                    color: Theme.of(context).colorScheme.onSurfaceVariant,
                  ),
            ),
            const SizedBox(height: 4),
            AnimatedSwitcher(
              duration: NetWatcherMotion.normal,
              child: Text(
                value,
                key: ValueKey<String>(value),
                style: const TextStyle(fontWeight: FontWeight.w700),
              ),
            ),
          ],
        ),
      );
}

/// The actual measured interval inside the selected history range.
class LatencyViewport {
  const LatencyViewport({required this.start, required this.end});

  final DateTime start;
  final DateTime end;

  static LatencyViewport? forTargets(
    List<TargetStatus> targets, {
    required int rangeMinutes,
    required DateTime now,
  }) {
    final cutoff = now.subtract(Duration(minutes: rangeMinutes));
    DateTime? first;
    DateTime? last;
    for (final target in targets) {
      for (final sample in target.history) {
        if (sample.time.isBefore(cutoff) || sample.time.isAfter(now)) continue;
        if (first == null || sample.time.isBefore(first)) first = sample.time;
        if (last == null || sample.time.isAfter(last)) last = sample.time;
      }
    }
    if (first == null || last == null) return null;
    final span = last.difference(first).inMilliseconds;
    final padding =
        Duration(milliseconds: math.max(1200, (span * .08).round()));
    final paddedStart = first.subtract(padding);
    return LatencyViewport(
      start: paddedStart.isBefore(cutoff) ? cutoff : paddedStart,
      end: last.add(padding),
    );
  }
}

class LatencyChart extends StatefulWidget {
  const LatencyChart({
    super.key,
    required this.targets,
    this.rangeMinutes = 5,
  });

  final List<TargetStatus> targets;
  final int rangeMinutes;

  @override
  State<LatencyChart> createState() => _LatencyChartState();
}

class _LatencyChartState extends State<LatencyChart> {
  String? selectedTargetId;

  String _legendValue(TargetStatus target) {
    if (target.state == 'offline' || target.latency <= 0) {
      return 'No response';
    }
    return '${target.latency.toStringAsFixed(1)} ms';
  }

  @override
  Widget build(BuildContext context) {
    final now = DateTime.now();
    final cutoff = now.subtract(Duration(minutes: widget.rangeMinutes));
    final allVisibleTargets = widget.targets
        .where((target) => target.history.any((sample) =>
            !sample.time.isBefore(cutoff) && !sample.time.isAfter(now)))
        .toList(growable: false);
    final selectedTargets = allVisibleTargets
        .where((target) => target.target.id == selectedTargetId)
        .toList(growable: false);
    final showingOne = selectedTargetId != null && selectedTargets.isNotEmpty;
    final visibleTargets = showingOne
        ? selectedTargets
        : allVisibleTargets.take(4).toList(growable: false);
    final moreCount =
        showingOne ? 0 : allVisibleTargets.length - visibleTargets.length;
    final viewport = LatencyViewport.forTargets(
      visibleTargets,
      rangeMinutes: widget.rangeMinutes,
      now: now,
    );

    if (viewport == null) {
      return SizedBox(
        height: 176,
        child: Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(
                Icons.show_chart,
                size: 30,
                color: Theme.of(context).colorScheme.onSurfaceVariant,
              ),
              const SizedBox(height: 10),
              Text(
                'Latency history will appear after measurements arrive.',
                textAlign: TextAlign.center,
                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                      color: Theme.of(context).colorScheme.onSurfaceVariant,
                    ),
              ),
            ],
          ),
        ),
      );
    }

    final scheme = Theme.of(context).colorScheme;
    const colors = <Color>[
      Color(0xFF81B989),
      Color(0xFFE7A574),
      Color(0xFF76B9BE),
      Color(0xFFB7A2D3),
      Color(0xFFE4C06F),
      Color(0xFFD8807B),
    ];

    return TweenAnimationBuilder<double>(
      tween: Tween<double>(begin: 0, end: 1),
      duration: NetWatcherMotion.slow,
      curve: NetWatcherMotion.curve,
      builder: (context, progress, child) => Opacity(
        opacity: progress,
        child: Transform.translate(
          offset: Offset(0, 8 * (1 - progress)),
          child: child,
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          SizedBox(
            height: 224,
            child: CustomPaint(
              painter: _LatencyPainter(
                grid: Theme.of(context).dividerColor,
                textColor: scheme.onSurface.withValues(alpha: .88),
                targets: visibleTargets,
                colors: colors,
                viewport: viewport,
                cutoff: cutoff,
              ),
              child: const SizedBox.expand(),
            ),
          ),
          const SizedBox(height: 12),
          Wrap(
            spacing: 16,
            runSpacing: 8,
            children: [
              for (var index = 0; index < visibleTargets.length; index++)
                Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Container(
                      width: 10,
                      height: 10,
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        color: colors[index % colors.length],
                        boxShadow: [
                          BoxShadow(
                            color: colors[index % colors.length]
                                .withValues(alpha: .38),
                            blurRadius: 8,
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(width: 7),
                    Flexible(
                      child: Text(
                        '${visibleTargets[index].target.name}: '
                        '${_legendValue(visibleTargets[index])}',
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                              color: scheme.onSurface,
                              fontWeight: FontWeight.w600,
                            ),
                      ),
                    ),
                  ],
                ),
              if (moreCount > 0)
                Text('$moreCount more targets',
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: scheme.onSurfaceVariant,
                        )),
            ],
          ),
          if (allVisibleTargets.length > 4) ...[
            const SizedBox(height: 12),
            Align(
              alignment: Alignment.centerRight,
              child: PopupMenuButton<String>(
                key: const ValueKey<String>('chart-target-filter'),
                tooltip: 'Choose chart target',
                onSelected: (value) => setState(() {
                  selectedTargetId = value.isEmpty ? null : value;
                }),
                itemBuilder: (context) => [
                  const PopupMenuItem<String>(
                    value: '',
                    child: Text('First 4 targets'),
                  ),
                  for (final target in allVisibleTargets)
                    PopupMenuItem<String>(
                      value: target.target.id,
                      child: Text(target.target.name,
                          maxLines: 1, overflow: TextOverflow.ellipsis),
                    ),
                ],
                child: Padding(
                  padding: const EdgeInsets.symmetric(vertical: 6),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(Icons.tune_rounded, size: 16, color: scheme.primary),
                      const SizedBox(width: 7),
                      Text(
                        showingOne ? 'Change target' : 'Choose a target',
                        style:
                            Theme.of(context).textTheme.labelMedium?.copyWith(
                                  color: scheme.primary,
                                  fontWeight: FontWeight.w700,
                                ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ],
        ],
      ),
    );
  }
}

class _LatencyPainter extends CustomPainter {
  _LatencyPainter({
    required this.grid,
    required this.textColor,
    required this.targets,
    required this.colors,
    required this.viewport,
    required this.cutoff,
  });

  final Color grid;
  final Color textColor;
  final List<TargetStatus> targets;
  final List<Color> colors;
  final LatencyViewport viewport;
  final DateTime cutoff;

  @override
  void paint(Canvas canvas, Size size) {
    final plot = Rect.fromLTRB(64, 12, size.width - 14, size.height - 34);

    final successful = <LatencySample>[
      for (final target in targets)
        for (final sample in target.history)
          if (sample.success && !sample.time.isBefore(cutoff)) sample,
    ];

    final rawMax = successful.isEmpty
        ? 0.0
        : successful.map((sample) => sample.latency).reduce(math.max);
    final axisStep = math.max(
      10.0,
      ((rawMax / 4) / 10).ceilToDouble() * 10,
    );
    final maxValue = axisStep * 4;

    final gridPaint = Paint()
      ..color = grid.withValues(alpha: .72)
      ..strokeWidth = 1;

    for (var index = 0; index <= 4; index++) {
      final fraction = index / 4;
      final y = plot.bottom - plot.height * fraction;
      canvas.drawLine(Offset(plot.left, y), Offset(plot.right, y), gridPaint);
      _drawText(
        canvas,
        '${(maxValue * fraction).round()} ms',
        Offset(0, y - 8),
        textColor,
        maxWidth: 58,
        align: TextAlign.right,
        fontSize: 12,
        fontWeight: FontWeight.w600,
      );
    }

    for (var index = 0; index <= 4; index++) {
      final x = plot.left + plot.width * index / 4;
      canvas.drawLine(Offset(x, plot.top), Offset(x, plot.bottom), gridPaint);
    }

    _drawText(
      canvas,
      _formatTime(viewport.start, withSeconds: _showSeconds),
      Offset(plot.left, plot.bottom + 8),
      textColor,
      fontSize: 11,
      fontWeight: FontWeight.w600,
    );
    _drawText(
      canvas,
      _formatTime(
          viewport.start.add(Duration(
              milliseconds:
                  viewport.end.difference(viewport.start).inMilliseconds ~/ 2)),
          withSeconds: _showSeconds),
      Offset(plot.center.dx - 30, plot.bottom + 8),
      textColor,
      fontSize: 11,
      fontWeight: FontWeight.w600,
    );
    _drawText(
      canvas,
      _formatTime(viewport.end, withSeconds: _showSeconds),
      Offset(plot.right - 58, plot.bottom + 8),
      textColor,
      fontSize: 11,
      fontWeight: FontWeight.w600,
    );

    final rangeMs =
        viewport.end.difference(viewport.start).inMilliseconds.toDouble();
    for (var targetIndex = 0; targetIndex < targets.length; targetIndex++) {
      final samples = targets[targetIndex]
          .history
          .where((sample) => !sample.time.isBefore(cutoff))
          .toList(growable: false);
      if (samples.isEmpty) continue;

      final path = Path();
      Offset? latestPoint;
      var drawing = false;
      for (final sample in samples) {
        if (!sample.success) {
          drawing = false;
          continue;
        }
        final elapsed =
            sample.time.difference(viewport.start).inMilliseconds.toDouble();
        final x = plot.left +
            (elapsed / rangeMs).clamp(0.0, 1.0).toDouble() * plot.width;
        final y = plot.bottom -
            (sample.latency / maxValue).clamp(0.0, 1.0).toDouble() *
                plot.height;
        latestPoint = Offset(x, y);
        if (!drawing) {
          path.moveTo(x, y);
          drawing = true;
        } else {
          path.lineTo(x, y);
        }
      }

      final color = colors[targetIndex % colors.length];
      canvas.drawPath(
        path,
        Paint()
          ..color = color.withValues(alpha: .22)
          ..style = PaintingStyle.stroke
          ..strokeWidth = 8
          ..strokeCap = StrokeCap.round
          ..strokeJoin = StrokeJoin.round
          ..maskFilter = const MaskFilter.blur(BlurStyle.normal, 4),
      );
      canvas.drawPath(
        path,
        Paint()
          ..color = color
          ..style = PaintingStyle.stroke
          ..strokeWidth = 3.2
          ..strokeCap = StrokeCap.round
          ..strokeJoin = StrokeJoin.round,
      );

      if (latestPoint != null) {
        canvas.drawCircle(
          latestPoint,
          6,
          Paint()..color = color.withValues(alpha: .20),
        );
        canvas.drawCircle(latestPoint, 2.8, Paint()..color = color);
      }
    }
  }

  bool get _showSeconds =>
      viewport.end.difference(viewport.start) < const Duration(minutes: 2);

  static String _formatTime(DateTime value, {required bool withSeconds}) {
    String two(int number) => number.toString().padLeft(2, '0');
    final minutes = '${two(value.hour)}:${two(value.minute)}';
    return withSeconds ? '$minutes:${two(value.second)}' : minutes;
  }

  static void _drawText(
    Canvas canvas,
    String text,
    Offset offset,
    Color color, {
    double maxWidth = 72,
    TextAlign align = TextAlign.left,
    double fontSize = 10,
    FontWeight fontWeight = FontWeight.w500,
  }) {
    final painter = TextPainter(
      text: TextSpan(
        text: text,
        style: TextStyle(
          color: color,
          fontSize: fontSize,
          fontWeight: fontWeight,
        ),
      ),
      textDirection: TextDirection.ltr,
      textAlign: align,
      maxLines: 1,
    )..layout(maxWidth: maxWidth);
    painter.paint(canvas, offset);
  }

  @override
  bool shouldRepaint(covariant _LatencyPainter oldDelegate) =>
      oldDelegate.targets != targets ||
      oldDelegate.grid != grid ||
      oldDelegate.textColor != textColor ||
      oldDelegate.colors != colors ||
      oldDelegate.viewport.start != viewport.start ||
      oldDelegate.viewport.end != viewport.end ||
      oldDelegate.cutoff != cutoff;
}
