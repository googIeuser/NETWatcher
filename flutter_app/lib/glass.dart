import 'dart:ui' as ui;

import 'package:flutter/material.dart';

/// A clipped, translucent surface for chrome and floating controls.
///
/// Keep it on small areas: backdrop blur is expensive over large regions.
class GlassSurface extends StatelessWidget {
  const GlassSurface({
    super.key,
    required this.child,
    this.radius = 18,
    this.blur = 18,
    this.padding = EdgeInsets.zero,
    this.accent,
  });

  final Widget child;
  final double radius;
  final double blur;
  final EdgeInsets padding;
  final Color? accent;

  @override
  Widget build(BuildContext context) {
    final dark = Theme.of(context).brightness == Brightness.dark;
    final shape = BorderRadius.circular(radius);
    final edge =
        accent ?? (dark ? const Color(0xFFB9ECDD) : const Color(0xFF537E71));
    return DecoratedBox(
      decoration: BoxDecoration(
        borderRadius: shape,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: dark ? .28 : .12),
            blurRadius: 24,
            offset: const Offset(0, 12),
          ),
        ],
      ),
      child: ClipRRect(
        borderRadius: shape,
        child: BackdropFilter(
          filter: ui.ImageFilter.blur(sigmaX: blur, sigmaY: blur),
          child: DecoratedBox(
            decoration: BoxDecoration(
              borderRadius: shape,
              border:
                  Border.all(color: edge.withValues(alpha: dark ? .31 : .35)),
              gradient: LinearGradient(
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
                colors: dark
                    ? const [
                        Color(0xE22C443A),
                        Color(0xD91A2B26),
                        Color(0xE3213930),
                      ]
                    : const [
                        Color(0xF2FFFFFF),
                        Color(0xEAF7FAF4),
                        Color(0xEFF0F7EF),
                      ],
                stops: const [0, .42, 1],
              ),
            ),
            child: Stack(
              children: [
                Padding(padding: padding, child: child),
                Positioned(
                  top: 0,
                  left: 14,
                  right: 14,
                  child: IgnorePointer(
                    child: Container(
                      height: 1,
                      decoration: BoxDecoration(
                        gradient: LinearGradient(colors: [
                          Colors.transparent,
                          Colors.white.withValues(alpha: dark ? .54 : .9),
                          Colors.transparent,
                        ]),
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class GlassAmbient extends StatelessWidget {
  const GlassAmbient({super.key});

  @override
  Widget build(BuildContext context) => IgnorePointer(
        child: CustomPaint(
          painter: _GlassAmbientPainter(
            dark: Theme.of(context).brightness == Brightness.dark,
          ),
          child: const SizedBox.expand(),
        ),
      );
}

class _GlassAmbientPainter extends CustomPainter {
  const _GlassAmbientPainter({required this.dark});

  final bool dark;

  @override
  void paint(Canvas canvas, Size size) {
    final areas = [
      (
        Offset(size.width * .08, size.height * .08),
        size.width * .48,
        const Color(0xFFD9F46A)
      ),
      (
        Offset(size.width * .94, size.height * .20),
        size.width * .48,
        const Color(0xFF78C9CC)
      ),
    ];
    for (final (center, radius, color) in areas) {
      canvas.drawCircle(
        center,
        radius,
        Paint()
          ..shader = ui.Gradient.radial(center, radius, [
            color.withValues(alpha: dark ? .085 : .10),
            color.withValues(alpha: 0),
          ]),
      );
    }
  }

  @override
  bool shouldRepaint(covariant _GlassAmbientPainter oldDelegate) =>
      dark != oldDelegate.dark;
}
