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
  const LatencyChart({super.key, required this.targets});
  final List<TargetStatus> targets;

  @override
  Widget build(BuildContext context) {
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
      child: SizedBox(
        height: 260,
        child: CustomPaint(
          painter: _LatencyPainter(
            color: Theme.of(context).colorScheme.primary,
            grid: Theme.of(context).dividerColor,
            targets: targets,
          ),
          child: const SizedBox.expand(),
        ),
      ),
    );
  }
}

class _LatencyPainter extends CustomPainter {
  _LatencyPainter({
    required this.color,
    required this.grid,
    required this.targets,
  });

  final Color color;
  final Color grid;
  final List<TargetStatus> targets;

  @override
  void paint(Canvas canvas, Size size) {
    final gridPaint = Paint()
      ..color = grid
      ..strokeWidth = 1;
    for (var index = 0; index <= 4; index++) {
      final y = size.height * index / 4;
      canvas.drawLine(Offset(0, y), Offset(size.width, y), gridPaint);
    }
    if (targets.isEmpty) return;

    final seed = targets.map((target) => target.latency).fold<double>(
          0,
          (sum, value) => sum + value,
        ) /
        targets.length;
    final values = List<double>.generate(
      48,
      (index) => math.max(
        2,
        seed + math.sin(index / 3) * 3 + (index % 9 == 0 ? 5 : 0),
      ),
    );
    final maxValue = math.max(50, values.reduce(math.max) * 1.2);
    final path = Path();
    for (var index = 0; index < values.length; index++) {
      final x = size.width * index / (values.length - 1);
      final y = size.height - (values[index] / maxValue * size.height);
      if (index == 0) {
        path.moveTo(x, y);
      } else {
        path.lineTo(x, y);
      }
    }
    final fill = Path.from(path)
      ..lineTo(size.width, size.height)
      ..lineTo(0, size.height)
      ..close();
    canvas.drawPath(
      fill,
      Paint()
        ..shader = LinearGradient(
          begin: Alignment.topCenter,
          end: Alignment.bottomCenter,
          colors: [color.withValues(alpha: .24), color.withValues(alpha: 0)],
        ).createShader(Offset.zero & size),
    );
    canvas.drawPath(
      path,
      Paint()
        ..color = color
        ..style = PaintingStyle.stroke
        ..strokeWidth = 2,
    );
  }

  @override
  bool shouldRepaint(covariant _LatencyPainter oldDelegate) =>
      oldDelegate.targets != targets ||
      oldDelegate.color != color ||
      oldDelegate.grid != grid;
}
