import 'package:flutter/material.dart';

import 'motion.dart';

/// The v5 visual system: warm paper, dark ink and a single signal accent.
class NetWatcherTheme {
  static const ink = Color(0xFF172821);
  static const signal = Color(0xFFD9F46A);

  static ButtonStyle _buttonStyle() => ButtonStyle(
        mouseCursor: WidgetStateMouseCursor.clickable,
        animationDuration: NetWatcherMotion.fast,
        shape: WidgetStatePropertyAll(
          RoundedRectangleBorder(borderRadius: BorderRadius.circular(5)),
        ),
        padding: const WidgetStatePropertyAll(
          EdgeInsets.symmetric(horizontal: 18, vertical: 15),
        ),
      );

  static ThemeData _build({required bool dark}) {
    final canvas = dark ? const Color(0xFF101A16) : const Color(0xFFF2F1E8);
    final surface = dark ? const Color(0xFF1B2721) : const Color(0xFFFCFBF5);
    final border = dark ? const Color(0xFF35443B) : const Color(0xFFCFD2C4);
    final primary = dark ? signal : ink;
    final text = dark ? const Color(0xFFF1F3E8) : ink;
    final muted = dark ? const Color(0xFFADBBAE) : const Color(0xFF64736A);
    final scheme = ColorScheme.fromSeed(
      seedColor: primary,
      brightness: dark ? Brightness.dark : Brightness.light,
    ).copyWith(
      primary: primary,
      onPrimary: dark ? ink : Colors.white,
      surface: surface,
      onSurface: text,
      onSurfaceVariant: muted,
      outline: border,
      error: dark ? const Color(0xFFFF8878) : const Color(0xFFAA443C),
    );
    final base = ThemeData(
      useMaterial3: true,
      brightness: dark ? Brightness.dark : Brightness.light,
      colorScheme: scheme,
      fontFamily: 'Bahnschrift',
    );
    return base.copyWith(
      scaffoldBackgroundColor: canvas,
      textTheme: base.textTheme.apply(bodyColor: text, displayColor: text),
      cardTheme: CardThemeData(
        color: surface,
        elevation: 0,
        margin: EdgeInsets.zero,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(6),
          side: BorderSide(color: border),
        ),
      ),
      filledButtonTheme: FilledButtonThemeData(style: _buttonStyle()),
      elevatedButtonTheme: ElevatedButtonThemeData(style: _buttonStyle()),
      outlinedButtonTheme: OutlinedButtonThemeData(style: _buttonStyle()),
      textButtonTheme: TextButtonThemeData(style: _buttonStyle()),
      iconButtonTheme: IconButtonThemeData(style: _buttonStyle()),
      dividerColor: border,
      tooltipTheme: const TooltipThemeData(
        waitDuration: Duration(milliseconds: 350),
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: dark ? const Color(0xFF24332A) : Colors.white,
        contentPadding:
            const EdgeInsets.symmetric(horizontal: 16, vertical: 15),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(5),
          borderSide: BorderSide(color: border),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(5),
          borderSide: BorderSide(color: border),
        ),
      ),
      switchTheme: SwitchThemeData(
        thumbColor: WidgetStateProperty.resolveWith(
          (states) => states.contains(WidgetState.selected) ? primary : muted,
        ),
      ),
    );
  }

  static ThemeData dark() => _build(dark: true);
  static ThemeData light() => _build(dark: false);
}
