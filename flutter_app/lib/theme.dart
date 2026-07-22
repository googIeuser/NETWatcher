import 'package:flutter/material.dart';

import 'motion.dart';

class NetWatcherTheme {
  static ButtonStyle _buttonStyle() => ButtonStyle(
        mouseCursor: WidgetStateMouseCursor.clickable,
        animationDuration: NetWatcherMotion.fast,
        shape: WidgetStatePropertyAll(
          RoundedRectangleBorder(borderRadius: BorderRadius.circular(11)),
        ),
      );

  static ThemeData dark() {
    const background = Color(0xFF0D1119);
    const panel = Color(0xFF161D29);
    const border = Color(0xFF293547);
    const blue = Color(0xFF4C8DFF);
    return ThemeData(
      useMaterial3: true,
      brightness: Brightness.dark,
      scaffoldBackgroundColor: background,
      colorScheme: ColorScheme.fromSeed(
        seedColor: blue,
        brightness: Brightness.dark,
        surface: panel,
      ),
      cardTheme: const CardThemeData(
        color: panel,
        elevation: 0,
        margin: EdgeInsets.zero,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.all(Radius.circular(18)),
          side: BorderSide(color: border),
        ),
      ),
      navigationRailTheme: const NavigationRailThemeData(
        backgroundColor: Color(0xFF111722),
        indicatorColor: Color(0xFF1A3154),
      ),
      navigationBarTheme: const NavigationBarThemeData(
        height: 68,
        indicatorColor: Color(0xFF1A3154),
      ),
      filledButtonTheme: FilledButtonThemeData(style: _buttonStyle()),
      elevatedButtonTheme: ElevatedButtonThemeData(style: _buttonStyle()),
      outlinedButtonTheme: OutlinedButtonThemeData(style: _buttonStyle()),
      textButtonTheme: TextButtonThemeData(style: _buttonStyle()),
      iconButtonTheme: IconButtonThemeData(style: _buttonStyle()),
      tooltipTheme: const TooltipThemeData(
        waitDuration: Duration(milliseconds: 350),
      ),
      dividerColor: border,
      inputDecorationTheme: const InputDecorationTheme(
        filled: true,
        fillColor: Color(0xFF1B2432),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.all(Radius.circular(12)),
          borderSide: BorderSide(color: border),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.all(Radius.circular(12)),
          borderSide: BorderSide(color: border),
        ),
      ),
    );
  }

  static ThemeData light() {
    const blue = Color(0xFF2677EC);
    return ThemeData(
      useMaterial3: true,
      brightness: Brightness.light,
      scaffoldBackgroundColor: const Color(0xFFF3F6FA),
      colorScheme: ColorScheme.fromSeed(
        seedColor: blue,
        brightness: Brightness.light,
        surface: Colors.white,
      ),
      cardTheme: const CardThemeData(
        color: Colors.white,
        elevation: 0,
        margin: EdgeInsets.zero,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.all(Radius.circular(18)),
          side: BorderSide(color: Color(0xFFDBE3EE)),
        ),
      ),
      navigationRailTheme: const NavigationRailThemeData(
        backgroundColor: Colors.white,
        indicatorColor: Color(0xFFDDEAFF),
      ),
      navigationBarTheme: const NavigationBarThemeData(
        height: 68,
        indicatorColor: Color(0xFFDDEAFF),
      ),
      filledButtonTheme: FilledButtonThemeData(style: _buttonStyle()),
      elevatedButtonTheme: ElevatedButtonThemeData(style: _buttonStyle()),
      outlinedButtonTheme: OutlinedButtonThemeData(style: _buttonStyle()),
      textButtonTheme: TextButtonThemeData(style: _buttonStyle()),
      iconButtonTheme: IconButtonThemeData(style: _buttonStyle()),
      tooltipTheme: const TooltipThemeData(
        waitDuration: Duration(milliseconds: 350),
      ),
      dividerColor: const Color(0xFFDBE3EE),
      inputDecorationTheme: const InputDecorationTheme(
        filled: true,
        fillColor: Color(0xFFF6F8FB),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.all(Radius.circular(12)),
        ),
      ),
    );
  }
}
