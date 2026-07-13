import 'package:flutter/material.dart';

import '../../../core/l10n/app_localizations.dart';
import '../../shared/widgets/kore_scaffold.dart';
import '../data/cra_models.dart';
import '../data/cra_repository.dart';
import 'cra_week_screen.dart';

class CraListScreen extends StatefulWidget {
  const CraListScreen({
    super.key,
    required this.repository,
    required this.onNavigateConges,
  });

  final CraRepository repository;
  final VoidCallback onNavigateConges;

  @override
  State<CraListScreen> createState() => _CraListScreenState();
}

class _CraListScreenState extends State<CraListScreen> {
  late Future<List<TimesheetSummary>> _future;

  @override
  void initState() {
    super.initState();
    _future = widget.repository.listRecent();
  }

  void _reload() {
    setState(() => _future = widget.repository.listRecent());
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return KoreScaffold(
      title: l10n.t('craTitle'),
      bottomNavIndex: 0,
      onNavTap: (index) {
        if (index == 1) widget.onNavigateConges();
      },
      body: RefreshIndicator(
        onRefresh: () async => _reload(),
        child: FutureBuilder<List<TimesheetSummary>>(
          future: _future,
          builder: (context, snapshot) {
            if (snapshot.connectionState == ConnectionState.waiting) {
              return ListView(
                children: [
                  const SizedBox(height: 120),
                  Center(child: Text(l10n.t('loading'))),
                ],
              );
            }
            if (snapshot.hasError) {
              return ListView(
                children: [
                  const SizedBox(height: 120),
                  Center(child: Text(l10n.t('errorGeneric'))),
                ],
              );
            }
            final items = snapshot.data ?? [];
            if (items.isEmpty) {
              return ListView(
                children: [
                  const SizedBox(height: 120),
                  Center(child: Text(l10n.t('craEmpty'))),
                ],
              );
            }
            return ListView.separated(
              padding: const EdgeInsets.all(16),
              itemCount: items.length,
              separatorBuilder: (context, _) => const SizedBox(height: 8),
              itemBuilder: (context, index) {
                final item = items[index];
                return Card(
                  child: ListTile(
                    title: Text(item.month),
                    subtitle: Text('${l10n.t('craStatus')}: ${item.status}'),
                    trailing: Text('${item.totalMinutes ~/ 60}h'),
                    onTap: () async {
                      final ts = await widget.repository.getByMonth(item.month);
                      if (!context.mounted) return;
                      await Navigator.of(context).push(
                        MaterialPageRoute<void>(
                          builder: (_) => CraWeekScreen(
                            repository: widget.repository,
                            timesheet: ts,
                          ),
                        ),
                      );
                      _reload();
                    },
                  ),
                );
              },
            );
          },
        ),
      ),
    );
  }
}
