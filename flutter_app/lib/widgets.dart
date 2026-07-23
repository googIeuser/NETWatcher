import 'dart:math' as math;

import 'package:flutter/material.dart';

import 'models.dart';
import 'motion.dart';

class Panel extends StatefulWidget {
  const Panel({
    super.key,
    required this.child,
    this.padding = const EdgeInsets.all(20),
    this.hoverEffect = true,
  });

  final Widget child;
  final EdgeInsets padding;
  final bool hoverEffect;

  @override
  State<Panel> createState() => _PanelState();
}

class _PanelState extends State<Panel> {
  bool hovered = false;

  @override
  Widget build(BuildContext context) {
    final end = widget.hoverEffect && hovered ? 1.0 : 0.0;
    return MouseRegion(
      onEnter: widget.hoverEffect ? (_) => setState(() => hovered = true) : null,
      onExit: widget.hoverEffect ? (_) => setState(() => hovered = false) : null,
      child: TweenAnimationBuilder<double>(
        tween: Tween<double>(end: end),
        duration: NetWatcherMotion.fast,
        curve: NetWatcherMotion.curve,
        builder: (context, progress, child) => Transform.translate(
          offset: Offset(0, -2 * progress),
          child: Transform.scale(
            scale: 1 + (.0025 * progress),
            alignment: Alignment.center,
            child: Card(
              elevation: 5 * progress,
              child: child,
            ),
          ),
        ),
        child: Padding(padding: widget.padding, child: widget.child),
      ),
    );
  }
}

class MetricCard extends StatelessWidget {
  const MetricCard({
    super.key,
    required this.label,
    required this.value,
    this.unit = '',
    this.icon,
  });

  final String label;
  final String value;
  final String unit;
  final IconData? icon;

  @override
  Widget build(BuildContext context) {
    return Panel(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              if (icon != null) ...[
                Icon(icon, size: 16, color: Theme.of(context).colorScheme.primary),
                const SizedBox(width: 8),
              ],
              Expanded(
                child: Text(
                  label,
                  style: Theme.of(context).textTheme.labelMedium?.copyWith(
                        color: Theme.of(context).colorScheme.onSurfaceVariant,
                      ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 14),
          AnimatedSwitcher(
            duration: NetWatcherMotion.normal,
            switchInCurve: NetWatcherMotion.curve,
            switchOutCurve: Curves.easeInCubic,
            transitionBuilder: (child, animation) => FadeTransition(
              opacity: animation,
              child: ScaleTransition(
                scale: Tween<double>(begin: .96, end: 1).animate(animation),
                child: child,
              ),
            ),
            child: FittedBox(
              key: ValueKey<String>('$value|$unit'),
              fit: BoxFit.scaleDown,
              alignment: Alignment.centerLeft,
              child: Text.rich(
                TextSpan(
                  text: value,
                  style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                        fontWeight: FontWeight.w800,
                      ),
                  children: [
                    if (unit.isNotEmpty)
                      TextSpan(
                        text: ' $unit',
                        style: Theme.of(context).textTheme.labelMedium,
                      ),
                  ],
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class TargetCard extends StatelessWidget {
  const TargetCard({super.key, required this.status});
  final TargetStatus status;

  Color _stateColor(BuildContext context) {
    return switch (status.state) {
      'online' => const Color(0xFF42D99A),
      'offline' => const Color(0xFFFF6D80),
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
        border: Border(
          top: BorderSide(color: Theme.of(context).dividerColor),
        ),
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
                value: '${status.latency.toStringAsFixed(1)} ms',
              ),
              _TargetMetric(
                label: 'Packet loss',
                value: '${status.packetLoss.toStringAsFixed(1)}%',
              ),
              _TargetMetric(
                label: 'Jitter',
                value: '${status.jitter.toStringAsFixed(1)} ms',
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

class LatencyChart extends StatelessWidget {
  const LatencyChart({
    super.key,
    required this.targets,
    this.rangeMinutes = 5,
  });

  final List<TargetStatus> targets;
  final int rangeMinutes;

  @override
  Widget build(BuildContext context) {
    final visibleTargets = targets
        .where((target) => target.history.isNotEmpty)
        .toList(growable: false);

    if (visibleTargets.isEmpty) {
      return const SizedBox(
        height: 280,
        child: Center(
          child: Text('Latency history will appear after measurements arrive.'),
        ),
      );
    }

    final scheme = Theme.of(context).colorScheme;
    final colors = <Color>[
      scheme.primary,
      scheme.tertiary,
      scheme.secondary,
      scheme.error,
      scheme.primary.withValues(alpha: .65),
      scheme.tertiary.withValues(alpha: .65),
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
            height: 280,
            child: CustomPaint(
              painter: _LatencyPainter(
                grid: Theme.of(context).dividerColor,
                textColor: scheme.onSurfaceVariant,
                targets: visibleTargets,
                colors: colors,
                rangeMinutes: rangeMinutes,
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
                      width: 9,
                      height: 9,
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        color: colors[index % colors.length],
                      ),
                    ),
                    const SizedBox(width: 7),
                    Text(
                      '${visibleTargets[index].target.name}: '
                      '${visibleTargets[index].latency.toStringAsFixed(1)} ms',
                      style: Theme.of(context).textTheme.bodySmall,
                    ),
                  ],
                ),
            ],
          ),
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
    this.rangeMinutes = 5,
  });

  final Color grid;
  final Color textColor;
  final List<TargetStatus> targets;
  final List<Color> colors;
  final int rangeMinutes;

  @override
  void paint(Canvas canvas, Size size) {
    final now = DateTime.now();
    final start = now.subtract(Duration(minutes: rangeMinutes));
    final plot = Rect.fromLTRB(50, 12, size.width - 12, size.height - 32);

    final successful = <LatencySample>[
      for (final target in targets)
        for (final sample in target.history)
          if (sample.success && !sample.time.isBefore(start)) sample,
    ];

    final rawMax = successful.isEmpty
        ? 0.0
        : successful.map((sample) => sample.latency).reduce(math.max);
    final maxValue =
        math.max(50.0, (rawMax / 10.0).ceilToDouble() * 10.0).toDouble();

    final gridPaint = Paint()
      ..color = grid
      ..strokeWidth = 1;

    for (var index = 0; index <= 4; index++) {
      final fraction = index / 4;
      final y = plot.bottom - plot.height * fraction;
      canvas.drawLine(Offset(plot.left, y), Offset(plot.right, y), gridPaint);
      _drawText(
        canvas,
        '${(maxValue * fraction).round()} ms',
        Offset(0, y - 7),
        textColor,
        maxWidth: 46,
        align: TextAlign.right,
      );
    }

    for (var index = 0; index <= 4; index++) {
      final x = plot.left + plot.width * index / 4;
      canvas.drawLine(Offset(x, plot.top), Offset(x, plot.bottom), gridPaint);
    }

    _drawText(
      canvas,
      _formatTime(start),
      Offset(plot.left, plot.bottom + 8),
      textColor,
    );
    _drawText(
      canvas,
      _formatTime(start.add(Duration(minutes: rangeMinutes ~/ 2))),
      Offset(plot.center.dx - 22, plot.bottom + 8),
      textColor,
    );
    _drawText(
      canvas,
      _formatTime(now),
      Offset(plot.right - 42, plot.bottom + 8),
      textColor,
    );

    final rangeMs = Duration(minutes: rangeMinutes).inMilliseconds.toDouble();
    for (var targetIndex = 0; targetIndex < targets.length; targetIndex++) {
      final samples = targets[targetIndex]
          .history
          .where((sample) => !sample.time.isBefore(start))
          .toList(growable: false);
      if (samples.isEmpty) continue;

      final path = Path();
      var drawing = false;
      for (final sample in samples) {
        if (!sample.success) {
          drawing = false;
          continue;
        }
        final elapsed = sample.time.difference(start).inMilliseconds.toDouble();
        final x = plot.left +
            (elapsed / rangeMs).clamp(0.0, 1.0).toDouble() * plot.width;
        final y = plot.bottom -
            (sample.latency / maxValue).clamp(0.0, 1.0).toDouble() *
                plot.height;
        if (!drawing) {
          path.moveTo(x, y);
          drawing = true;
        } else {
          path.lineTo(x, y);
        }
      }

      canvas.drawPath(
        path,
        Paint()
          ..color = colors[targetIndex % colors.length]
          ..style = PaintingStyle.stroke
          ..strokeWidth = 2
          ..strokeCap = StrokeCap.round
          ..strokeJoin = StrokeJoin.round,
      );
    }
  }

  static String _formatTime(DateTime value) {
    String two(int number) => number.toString().padLeft(2, '0');
    return '${two(value.hour)}:${two(value.minute)}';
  }

  static void _drawText(
    Canvas canvas,
    String text,
    Offset offset,
    Color color, {
    double maxWidth = 64,
    TextAlign align = TextAlign.left,
  }) {
    final painter = TextPainter(
      text: TextSpan(
        text: text,
        style: TextStyle(color: color, fontSize: 10),
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
      oldDelegate.rangeMinutes != rangeMinutes;
}
