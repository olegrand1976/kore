import 'package:flutter/material.dart';

/// Kore design tokens — mirror of `--kore-*` CSS (documentation/CHARTE_GRAPHIQUE.md).
abstract final class KoreColors {
  // Brand (invariant)
  static const brandBlue = Color(0xFF2B6CB0);
  static const brandBlueLight = Color(0xFF3B82F6);
  static const brandNavy = Color(0xFF1E3A5F);
  static const brandGold = Color(0xFFC9A227);
  static const brandGoldLight = Color(0xFFE8C547);
  static const brandGoldDark = Color(0xFFA68520);

  // Dark semantic
  static const darkBg = Color(0xFF1A1F2E);
  static const darkBgElevated = Color(0xFF252B3B);
  static const darkBgSubtle = Color(0xFF1E2433);
  static const darkBorder = Color(0xFF3D4559);
  static const darkText = Color(0xFFE8EAED);
  static const darkTextMuted = Color(0xFF9CA3AF);
  static const darkLink = Color(0xFF60A5FA);

  // Light semantic
  static const lightBg = Color(0xFFF8F9FB);
  static const lightBgElevated = Color(0xFFFFFFFF);
  static const lightBgSubtle = Color(0xFFF1F3F6);
  static const lightBorder = Color(0xFFE2E6ED);
  static const lightText = Color(0xFF1A1F2E);
  static const lightTextMuted = Color(0xFF6B7280);

  static const error = Color(0xFFF87171);
  static const success = Color(0xFF4ADE80);
}

abstract final class KoreTheme {
  static ThemeData light() {
    const scheme = ColorScheme(
      brightness: Brightness.light,
      primary: KoreColors.brandGold,
      onPrimary: KoreColors.lightText,
      secondary: KoreColors.brandBlue,
      onSecondary: Colors.white,
      surface: KoreColors.lightBgElevated,
      onSurface: KoreColors.lightText,
      error: KoreColors.error,
      onError: Colors.white,
    );
    return _base(scheme).copyWith(
      scaffoldBackgroundColor: KoreColors.lightBg,
      dividerColor: KoreColors.lightBorder,
      appBarTheme: const AppBarTheme(
        backgroundColor: KoreColors.lightBgElevated,
        foregroundColor: KoreColors.brandNavy,
        elevation: 0,
        centerTitle: false,
      ),
      cardTheme: CardThemeData(
        color: KoreColors.lightBgElevated,
        elevation: 0,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(12),
          side: const BorderSide(color: KoreColors.lightBorder),
        ),
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: KoreColors.lightBgSubtle,
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: const BorderSide(color: KoreColors.lightBorder),
        ),
      ),
      bottomNavigationBarTheme: const BottomNavigationBarThemeData(
        backgroundColor: KoreColors.lightBgElevated,
        selectedItemColor: KoreColors.brandBlue,
        unselectedItemColor: KoreColors.lightTextMuted,
        type: BottomNavigationBarType.fixed,
      ),
    );
  }

  static ThemeData dark() {
    const scheme = ColorScheme(
      brightness: Brightness.dark,
      primary: KoreColors.brandGold,
      onPrimary: KoreColors.darkBg,
      secondary: KoreColors.brandBlueLight,
      onSecondary: KoreColors.darkBg,
      surface: KoreColors.darkBgElevated,
      onSurface: KoreColors.darkText,
      error: KoreColors.error,
      onError: KoreColors.darkBg,
    );
    return _base(scheme).copyWith(
      scaffoldBackgroundColor: KoreColors.darkBg,
      dividerColor: KoreColors.darkBorder,
      appBarTheme: const AppBarTheme(
        backgroundColor: KoreColors.darkBgElevated,
        foregroundColor: KoreColors.darkText,
        elevation: 0,
        centerTitle: false,
      ),
      cardTheme: CardThemeData(
        color: KoreColors.darkBgElevated,
        elevation: 0,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(12),
          side: const BorderSide(color: KoreColors.darkBorder),
        ),
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: KoreColors.darkBgSubtle,
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: const BorderSide(color: KoreColors.darkBorder),
        ),
      ),
      bottomNavigationBarTheme: const BottomNavigationBarThemeData(
        backgroundColor: KoreColors.darkBgElevated,
        selectedItemColor: KoreColors.brandGold,
        unselectedItemColor: KoreColors.darkTextMuted,
        type: BottomNavigationBarType.fixed,
      ),
    );
  }

  static ThemeData _base(ColorScheme scheme) {
    return ThemeData(
      useMaterial3: true,
      colorScheme: scheme,
      fontFamily: 'Roboto',
      textTheme: const TextTheme(
        headlineLarge: TextStyle(fontSize: 36, fontWeight: FontWeight.w700),
        headlineMedium: TextStyle(fontSize: 28, fontWeight: FontWeight.w600),
        titleLarge: TextStyle(fontSize: 20, fontWeight: FontWeight.w600),
        titleMedium: TextStyle(fontSize: 16, fontWeight: FontWeight.w600),
        bodyLarge: TextStyle(fontSize: 16),
        bodyMedium: TextStyle(fontSize: 14),
        bodySmall: TextStyle(fontSize: 12),
        labelLarge: TextStyle(fontSize: 14, fontWeight: FontWeight.w600),
      ),
      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          minimumSize: const Size.fromHeight(48),
          backgroundColor: scheme.primary,
          foregroundColor: scheme.onPrimary,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
        ),
      ),
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          minimumSize: const Size.fromHeight(48),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
        ),
      ),
    );
  }
}
