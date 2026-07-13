import 'package:flutter/material.dart';

import '../../../core/l10n/app_localizations.dart';

class KoreScaffold extends StatelessWidget {
  const KoreScaffold({
    super.key,
    required this.title,
    required this.body,
    this.actions,
    this.bottomNavIndex = 0,
    this.onNavTap,
    this.floatingActionButton,
  });

  final String title;
  final Widget body;
  final List<Widget>? actions;
  final int bottomNavIndex;
  final ValueChanged<int>? onNavTap;
  final Widget? floatingActionButton;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Scaffold(
      appBar: AppBar(
        title: Text(title),
        actions: actions,
      ),
      body: SafeArea(child: body),
      floatingActionButton: floatingActionButton,
      bottomNavigationBar: onNavTap == null
          ? null
          : BottomNavigationBar(
              currentIndex: bottomNavIndex,
              onTap: onNavTap,
              items: [
                BottomNavigationBarItem(
                  icon: const Icon(Icons.calendar_month_outlined),
                  label: l10n.t('navCra'),
                ),
                BottomNavigationBarItem(
                  icon: const Icon(Icons.beach_access_outlined),
                  label: l10n.t('navConges'),
                ),
              ],
            ),
    );
  }
}
