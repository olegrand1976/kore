import 'package:flutter/material.dart';

import '../../../core/l10n/app_localizations.dart';
import '../data/leave_models.dart';
import '../data/leave_repository.dart';

class LeaveBalancesScreen extends StatefulWidget {
  const LeaveBalancesScreen({super.key, required this.repository});

  final LeaveRepository repository;

  @override
  State<LeaveBalancesScreen> createState() => _LeaveBalancesScreenState();
}

class _LeaveBalancesScreenState extends State<LeaveBalancesScreen> {
  late Future<List<LeaveBalance>> _future;

  @override
  void initState() {
    super.initState();
    _future = widget.repository.listBalances();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Scaffold(
      appBar: AppBar(title: Text(l10n.t('congesBalances'))),
      body: SafeArea(
        child: FutureBuilder<List<LeaveBalance>>(
          future: _future,
          builder: (context, snapshot) {
            if (snapshot.connectionState == ConnectionState.waiting) {
              return Center(child: Text(l10n.t('loading')));
            }
            if (snapshot.hasError) {
              return Center(child: Text(l10n.t('errorGeneric')));
            }
            final items = snapshot.data ?? [];
            if (items.isEmpty) {
              return Center(child: Text(l10n.t('congesEmpty')));
            }
            return ListView.separated(
              padding: const EdgeInsets.all(16),
              itemCount: items.length,
              separatorBuilder: (context, _) => const SizedBox(height: 8),
              itemBuilder: (context, index) {
                final b = items[index];
                return Card(
                  child: ListTile(
                    title: Text(b.type),
                    subtitle: Text(
                      '${b.taken.toStringAsFixed(1)} / ${b.acquired.toStringAsFixed(1)}',
                    ),
                    trailing: Text(b.remaining.toStringAsFixed(1)),
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
