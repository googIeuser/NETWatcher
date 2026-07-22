import 'package:flutter/material.dart';

abstract final class NetWatcherMotion {
  static const fast = Duration(milliseconds: 150);
  static const normal = Duration(milliseconds: 240);
  static const slow = Duration(milliseconds: 360);
  static const curve = Curves.easeOutCubic;
  static const emphasizedCurve = Curves.easeInOutCubicEmphasized;
}

class FadeSlideSwitcher extends StatelessWidget {
  const FadeSlideSwitcher({
    super.key,
    required this.child,
    this.duration = NetWatcherMotion.normal,
  });

  final Widget child;
  final Duration duration;

  @override
  Widget build(BuildContext context) {
    return AnimatedSwitcher(
      duration: duration,
      reverseDuration: NetWatcherMotion.fast,
      switchInCurve: NetWatcherMotion.curve,
      switchOutCurve: Curves.easeInCubic,
      layoutBuilder: (currentChild, previousChildren) => Stack(
        fit: StackFit.expand,
        children: [
          ...previousChildren,
          if (currentChild != null) currentChild,
        ],
      ),
      transitionBuilder: (child, animation) {
        final curved = CurvedAnimation(
          parent: animation,
          curve: NetWatcherMotion.curve,
          reverseCurve: Curves.easeInCubic,
        );
        final offset = Tween<Offset>(
          begin: const Offset(.018, 0),
          end: Offset.zero,
        ).animate(curved);
        return FadeTransition(
          opacity: curved,
          child: SlideTransition(position: offset, child: child),
        );
      },
      child: child,
    );
  }
}

class AnimatedValue extends StatelessWidget {
  const AnimatedValue({
    super.key,
    required this.value,
    required this.builder,
  });

  final Object value;
  final Widget Function(BuildContext context) builder;

  @override
  Widget build(BuildContext context) {
    return AnimatedSwitcher(
      duration: NetWatcherMotion.normal,
      switchInCurve: NetWatcherMotion.curve,
      switchOutCurve: Curves.easeInCubic,
      transitionBuilder: (child, animation) => FadeTransition(
        opacity: animation,
        child: ScaleTransition(
          scale: Tween<double>(begin: .97, end: 1).animate(animation),
          child: child,
        ),
      ),
      child: KeyedSubtree(
        key: ValueKey<Object>(value),
        child: Builder(builder: builder),
      ),
    );
  }
}
