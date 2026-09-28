import 'dart:math' as math;

import 'package:flutter/material.dart';

import 'models.dart';
import 'theme.dart';

/// A compact topology view of the actual endpoints in the current snapshot.
class NetworkMap extends StatelessWidget {
  const NetworkMap({super.key, required this.snapshot});

  final NetworkSnapshot snapshot;

  bool get _live => snapshot.monitoring && snapshot.samples > 0;

  Color _statusColor(TargetStatus target) {
    if (!_live) return const Color(0xFF8DA399);
    if (target.state == 'online') return NetWatcherTheme.signal;
    if (target.state == 'offline') return const Color(0xFFFF8878);
    return const Color(0xFF8DA399);
  }

  @override
  Widget build(BuildContext context) {
    final targets = snapshot.targets.take(4).toList(growable: false);
    final status = !snapshot.monitoring
        ? 'Monitoring paused'
        : !_live
            ? 'Checking connection'
            : snapshot.connectionLabel;
    final statusColor = !_live
        ? const Color(0xFFDAE4D9)
        : snapshot.connectionState == 'offline'
            ? const Color(0xFFFF8878)
            : NetWatcherTheme.signal;
    return Container(
      key: _live ? null : const ValueKey<String>('dashboard-empty-state'),
      decoration: BoxDecoration(
        color: const Color(0xFF14221B),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: const Color(0xFF365345)),
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
                const Text('NETWORK MAP',
                    style: TextStyle(
                      color: NetWatcherTheme.signal,
                      fontFamily: 'Consolas',
                      fontSize: 11,
                      fontWeight: FontWeight.bold,
                      letterSpacing: 2,
                    )),
                const Spacer(),
                Text('${snapshot.targets.length} TARGETS',
                    style: const TextStyle(
                      color: Color(0xFF9FB4A6),
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
              style: const TextStyle(color: Color(0xFF9FB4A6)),
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
                        color: _statusColor(target),
                        live: _live,
                      ),
                    if (targets.isEmpty)
                      const Padding(
                        padding: EdgeInsets.symmetric(vertical: 18),
                        child: Text('No targets available yet.',
                            style: TextStyle(color: Color(0xFF9FB4A6))),
                      ),
                    if (snapshot.targets.length > 4)
                      _MoreTargets(snapshot.targets.length - 4),
                  ],
                ),
              );
            }
            return SizedBox(
              height: 286,
              child: Stack(
                children: [
                  Positioned.fill(
                    child: CustomPaint(
                      painter: _MapLines(
                        colors: targets.map(_statusColor).toList(),
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
                        border: Border.all(
                            color: const Color(0xFF14221B), width: 6),
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
                    Align(
                      alignment: _MapLines.alignments[i],
                      child: _MapNode(
                        target: targets[i],
                        color: _statusColor(targets[i]),
                        live: _live,
                      ),
                    ),
                  if (targets.isEmpty)
                    const Align(
                      alignment: Alignment(0, .84),
                      child: Text('No targets available yet.',
                          style: TextStyle(color: Color(0xFF9FB4A6))),
                    ),
                  if (snapshot.targets.length > 4)
                    Align(
                      alignment: const Alignment(0, .95),
                      child: _MoreTargets(snapshot.targets.length - 4),
                    ),
                ],
              ),
            );
          }),
          Container(
            padding: const EdgeInsets.fromLTRB(24, 14, 24, 16),
            decoration: const BoxDecoration(
              border: Border(top: BorderSide(color: Color(0xFF365345))),
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
  _MapLines({required this.colors});

  final List<Color> colors;
  static const alignments = [
    Alignment(-.77, -.53),
    Alignment(.77, -.53),
    Alignment(-.77, .58),
    Alignment(.77, .58),
  ];

  @override
  void paint(Canvas canvas, Size size) {
    final center = Offset(size.width / 2, size.height / 2);
    for (var i = 0; i < colors.length; i++) {
      final alignment = alignments[i];
      final destination = Offset(
        size.width / 2 + alignment.x * (size.width - 154) / 2,
        size.height / 2 + alignment.y * (size.height - 85) / 2 - 20,
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
    if (colors.length != oldDelegate.colors.length) return true;
    for (var i = 0; i < colors.length; i++) {
      if (colors[i] != oldDelegate.colors[i]) return true;
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
  Widget build(BuildContext context) => SizedBox(
        width: 154,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              width: 44,
              height: 44,
              decoration: BoxDecoration(
                color: const Color(0xFF1F352A),
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
                style: const TextStyle(
                    color: Colors.white, fontWeight: FontWeight.w700)),
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

class _MapTargetRow extends StatelessWidget {
  const _MapTargetRow(
      {required this.target, required this.color, required this.live});
  final TargetStatus target;
  final Color color;
  final bool live;

  @override
  Widget build(BuildContext context) => Container(
        padding: const EdgeInsets.symmetric(vertical: 10),
        decoration: const BoxDecoration(
          border: Border(top: BorderSide(color: Color(0xFF365345))),
        ),
        child: Row(
          children: [
            Icon(Icons.radio_button_checked, size: 16, color: color),
            const SizedBox(width: 12),
            Expanded(
              child: Text(target.target.name,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(color: Colors.white)),
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

class _MapDatum extends StatelessWidget {
  const _MapDatum(this.label, this.value);
  final String label;
  final String value;

  @override
  Widget build(BuildContext context) => Text.rich(TextSpan(children: [
        TextSpan(
            text: '$label  ',
            style: const TextStyle(color: Color(0xFF9FB4A6), fontSize: 11)),
        TextSpan(
            text: value,
            style: const TextStyle(
                color: Colors.white,
                fontSize: 13,
                fontWeight: FontWeight.bold)),
      ]));
}

class _MoreTargets extends StatelessWidget {
  const _MoreTargets(this.count);
  final int count;

  @override
  Widget build(BuildContext context) => Text('+$count more targets below',
      style: const TextStyle(color: Color(0xFF9FB4A6), fontSize: 11));
}
