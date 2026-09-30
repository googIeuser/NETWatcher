import 'dart:math' as math;

import 'package:flutter/material.dart';

import 'models.dart';
import 'theme.dart';

class _MapPalette {
  const _MapPalette({
    required this.background,
    required this.border,
    required this.nodeFill,
    required this.text,
    required this.muted,
    required this.inactive,
    required this.accent,
    required this.error,
  });

  final Color background;
  final Color border;
  final Color nodeFill;
  final Color text;
  final Color muted;
  final Color inactive;
  final Color accent;
  final Color error;

  factory _MapPalette.of(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    if (scheme.brightness == Brightness.dark) {
      return const _MapPalette(
        background: Color(0xFF14221B),
        border: Color(0xFF365345),
        nodeFill: Color(0xFF1F352A),
        text: Color(0xFFDAE4D9),
        muted: Color(0xFF9FB4A6),
        inactive: Color(0xFF8DA399),
        accent: NetWatcherTheme.signal,
        error: Color(0xFFFF8878),
      );
    }
    return _MapPalette(
      background: scheme.surface,
      border: scheme.outline,
      nodeFill: const Color(0xFFE9EFE6),
      text: scheme.onSurface,
      muted: scheme.onSurfaceVariant,
      inactive: const Color(0xFF75877B),
      accent: const Color(0xFF58751B),
      error: scheme.error,
    );
  }
}

/// A compact topology view of the actual endpoints in the current snapshot.
class NetworkMap extends StatelessWidget {
  const NetworkMap({super.key, required this.snapshot});

  final NetworkSnapshot snapshot;

  bool get _live => snapshot.monitoring && snapshot.samples > 0;

  Color _statusColor(TargetStatus target, _MapPalette palette) {
    if (!_live) return palette.inactive;
    if (target.state == 'online') return palette.accent;
    if (target.state == 'offline') return palette.error;
    return palette.inactive;
  }

  @override
  Widget build(BuildContext context) {
    final palette = _MapPalette.of(context);
    final targets = snapshot.targets.take(10).toList(growable: false);
    final rows = (targets.length + 1) ~/ 2;
    const rowPitch = 112.0;
    const nodeHeight = 92.0;
    final mapHeight = math.max(286.0, rows * rowPitch + 24);
    final firstTop =
        (mapHeight - nodeHeight - math.max(0, rows - 1) * rowPitch) / 2;
    final nodeTops = List<double>.generate(
        targets.length, (index) => firstTop + (index ~/ 2) * rowPitch);
    final status = !snapshot.monitoring
        ? 'Monitoring paused'
        : !_live
            ? 'Checking connection'
            : snapshot.connectionLabel;
    final statusColor = !_live
        ? palette.text
        : snapshot.connectionState == 'offline'
            ? palette.error
            : palette.accent;
    return Container(
      key: _live ? null : const ValueKey<String>('dashboard-empty-state'),
      decoration: BoxDecoration(
        color: palette.background,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: palette.border),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(24, 22, 24, 0),
            child: Row(
              children: [
                Container(
                  width: 8,
                  height: 8,
                  decoration: BoxDecoration(
                    color: statusColor,
                    shape: BoxShape.circle,
                  ),
                ),
                const SizedBox(width: 10),
                Text('NETWORK MAP',
                    style: TextStyle(
                      color: palette.accent,
                      fontFamily: 'Consolas',
                      fontSize: 11,
                      fontWeight: FontWeight.bold,
                      letterSpacing: 2,
                    )),
                const Spacer(),
                Text('${snapshot.targets.length} TARGETS',
                    style: TextStyle(
                      color: palette.muted,
                      fontFamily: 'Consolas',
                      fontSize: 11,
                      letterSpacing: 1.2,
                    )),
              ],
            ),
          ),
          Padding(
            padding: const EdgeInsets.fromLTRB(24, 12, 24, 0),
            child: Text(status,
                style: TextStyle(
                  color: statusColor,
                  fontSize: 28,
                  fontWeight: FontWeight.w800,
                  letterSpacing: -.5,
                )),
          ),
          Padding(
            padding: const EdgeInsets.fromLTRB(24, 2, 24, 14),
            child: Text(
              !_live
                  ? 'Start monitoring to see live connection paths.'
                  : 'Live paths from this computer to each monitored target.',
              style: TextStyle(color: palette.muted),
            ),
          ),
          LayoutBuilder(builder: (context, constraints) {
            if (constraints.maxWidth < 700) {
              return Padding(
                padding: const EdgeInsets.fromLTRB(24, 8, 24, 22),
                child: Column(
                  children: [
                    for (final target in targets)
                      _MapTargetRow(
                        target: target,
                        color: _statusColor(target, palette),
                        live: _live,
                      ),
                    if (targets.isEmpty)
                      Padding(
                        padding: const EdgeInsets.symmetric(vertical: 18),
                        child: Text('No targets available yet.',
                            style: TextStyle(color: palette.muted)),
                      ),
                    if (snapshot.targets.length > 10)
                      _MoreTargets(snapshot.targets.length - 10),
                  ],
                ),
              );
            }
            return SizedBox(
              height: mapHeight,
              child: Stack(
                children: [
                  Positioned.fill(
                    child: CustomPaint(
                      painter: _MapLines(
                        colors: targets
                            .map((target) => _statusColor(target, palette))
                            .toList(),
                        nodeTops: nodeTops,
                      ),
                    ),
                  ),
                  Align(
                    child: Container(
                      width: 104,
                      height: 104,
                      decoration: BoxDecoration(
                        color: NetWatcherTheme.signal,
                        shape: BoxShape.circle,
                        border: Border.all(color: palette.background, width: 6),
                      ),
                      child: const Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(Icons.hub_rounded,
                              size: 30, color: NetWatcherTheme.ink),
                          SizedBox(height: 2),
                          Text('YOU',
                              style: TextStyle(
                                color: NetWatcherTheme.ink,
                                fontFamily: 'Consolas',
                                fontSize: 11,
                                fontWeight: FontWeight.w900,
                                letterSpacing: 1,
                              )),
                        ],
                      ),
                    ),
                  ),
                  for (var i = 0; i < targets.length; i++)
                    Positioned(
                      top: nodeTops[i],
                      left: 0,
                      right: 0,
                      child: Align(
                        alignment: Alignment(i.isEven ? -.77 : .77, 0),
                        child: _MapNode(
                          target: targets[i],
                          color: _statusColor(targets[i], palette),
                          live: _live,
                        ),
                      ),
                    ),
                  if (targets.isEmpty)
                    Align(
                      alignment: const Alignment(0, .84),
                      child: Text('No targets available yet.',
                          style: TextStyle(color: palette.muted)),
                    ),
                ],
              ),
            );
          }),
          if (snapshot.targets.length > 10)
            LayoutBuilder(builder: (context, constraints) {
              if (constraints.maxWidth < 700) return const SizedBox.shrink();
              return Padding(
                padding: const EdgeInsets.only(bottom: 14),
                child:
                    Center(child: _MoreTargets(snapshot.targets.length - 10)),
              );
            }),
          Container(
            padding: const EdgeInsets.fromLTRB(24, 14, 24, 16),
            decoration: BoxDecoration(
              border: Border(top: BorderSide(color: palette.border)),
            ),
            child: Wrap(
              spacing: 24,
              runSpacing: 8,
              children: [
                _MapDatum(
                    'AVG LATENCY',
                    _live
                        ? '${snapshot.averageLatency.toStringAsFixed(1)} ms'
                        : '—'),
                _MapDatum('PACKET LOSS',
                    _live ? '${snapshot.packetLoss.toStringAsFixed(1)}%' : '—'),
                _MapDatum(
                    'QUALITY', _live ? '${snapshot.qualityScore}/100' : '—'),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _MapLines extends CustomPainter {
  _MapLines({required this.colors, required this.nodeTops});

  final List<Color> colors;
  final List<double> nodeTops;

  @override
  void paint(Canvas canvas, Size size) {
    final center = Offset(size.width / 2, size.height / 2);
    for (var i = 0; i < colors.length; i++) {
      final destination = Offset(
        size.width / 2 + (i.isEven ? -.77 : .77) * (size.width - 154) / 2,
        nodeTops[i] + 22,
      );
      final direction = destination - center;
      final length = direction.distance;
      if (length < 1) continue;
      final unit = direction / length;
      final paint = Paint()
        ..color = colors[i].withValues(alpha: .75)
        ..strokeWidth = 3
        ..strokeCap = StrokeCap.round;
      canvas.drawLine(center + unit * 52,
          destination - unit * math.min(22, length / 3), paint);
    }
  }

  @override
  bool shouldRepaint(covariant _MapLines oldDelegate) {
    if (colors.length != oldDelegate.colors.length ||
        nodeTops.length != oldDelegate.nodeTops.length) {
      return true;
    }
    for (var i = 0; i < colors.length; i++) {
      if (colors[i] != oldDelegate.colors[i] ||
          nodeTops[i] != oldDelegate.nodeTops[i]) {
        return true;
      }
    }
    return false;
  }
}

class _MapNode extends StatelessWidget {
  const _MapNode(
      {required this.target, required this.color, required this.live});

  final TargetStatus target;
  final Color color;
  final bool live;

  @override
  Widget build(BuildContext context) {
    final palette = _MapPalette.of(context);
    return SizedBox(
      width: 154,
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            width: 44,
            height: 44,
            decoration: BoxDecoration(
              color: palette.nodeFill,
              shape: BoxShape.circle,
              border: Border.all(color: color, width: 2),
            ),
            child: Icon(Icons.circle, size: 13, color: color),
          ),
          const SizedBox(height: 6),
          Text(target.target.name,
              textAlign: TextAlign.center,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style:
                  TextStyle(color: palette.text, fontWeight: FontWeight.w700)),
          Text(
            live && target.state == 'online'
                ? '${target.latency.toStringAsFixed(1)} ms'
                : live && target.state == 'offline'
                    ? 'OFFLINE'
                    : target.target.host,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            style:
                TextStyle(color: color, fontFamily: 'Consolas', fontSize: 11),
          ),
        ],
      ),
    );
  }
}

class _MapTargetRow extends StatelessWidget {
  const _MapTargetRow(
      {required this.target, required this.color, required this.live});
  final TargetStatus target;
  final Color color;
  final bool live;

  @override
  Widget build(BuildContext context) {
    final palette = _MapPalette.of(context);
    return Container(
      padding: const EdgeInsets.symmetric(vertical: 10),
      decoration: BoxDecoration(
        border: Border(top: BorderSide(color: palette.border)),
      ),
      child: Row(
        children: [
          Icon(Icons.radio_button_checked, size: 16, color: color),
          const SizedBox(width: 12),
          Expanded(
            child: Text(target.target.name,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: TextStyle(color: palette.text)),
          ),
          Text(
            live && target.state == 'online'
                ? '${target.latency.toStringAsFixed(1)} ms'
                : live && target.state == 'offline'
                    ? 'OFFLINE'
                    : '—',
            style: TextStyle(color: color, fontFamily: 'Consolas'),
          ),
        ],
      ),
    );
  }
}

class _MapDatum extends StatelessWidget {
  const _MapDatum(this.label, this.value);
  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    final palette = _MapPalette.of(context);
    return Text.rich(TextSpan(children: [
      TextSpan(
          text: '$label  ',
          style: TextStyle(color: palette.muted, fontSize: 11)),
      TextSpan(
          text: value,
          style: TextStyle(
              color: palette.text, fontSize: 13, fontWeight: FontWeight.bold)),
    ]));
  }
}

class _MoreTargets extends StatelessWidget {
  const _MoreTargets(this.count);
  final int count;

  @override
  Widget build(BuildContext context) => Text('+$count more targets below',
      style: TextStyle(color: _MapPalette.of(context).muted, fontSize: 11));
}
