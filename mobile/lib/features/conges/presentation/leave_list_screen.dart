import 'package:flutter/material.dart';

import '../../../core/l10n/app_localizations.dart';
import '../../shared/widgets/kore_scaffold.dart';
import '../data/leave_models.dart';
import '../data/leave_repository.dart';
import 'leave_request_screen.dart';
import 'leave_balances_screen.dart';
import 'leave_validation_screen.dart';

class LeaveListScreen extends StatefulWidget {
  const LeaveListScreen({
    super.key,
    required this.repository,
    required this.onNavigateCra,
    this.canValidate = false,
  });

  final LeaveRepository repository;
  final VoidCallback onNavigateCra;
  final bool canValidate;

  @override
  State<LeaveListScreen> createState() => _LeaveListScreenState();
}

class _LeaveListScreenState extends State<LeaveListScreen> {
  late Future<List<LeaveRequest>> _future;

  @override
  void initState() {
    super.initState();
    _future = widget.repository.listRequests();
  }

  void _reload() => setState(() => _future = widget.repository.listRequests());

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return KoreScaffold(
      title: l10n.t('congesTitle'),
      bottomNavIndex: 1,
      onNavTap: (index) {
        if (index == 0) widget.onNavigateCra();
      },
      actions: [
        IconButton(
          icon: const Icon(Icons.account_balance_wallet_outlined),
          tooltip: l10n.t('congesBalances'),
          onPressed: () {
            Navigator.of(context).push(
              MaterialPageRoute<void>(
                builder: (_) =>
                    LeaveBalancesScreen(repository: widget.repository),
              ),
            );
          },
        ),
        if (widget.canValidate)
          IconButton(
            icon: const Icon(Icons.rule_outlined),
            tooltip: l10n.t('congesValidation'),
            onPressed: () {
              Navigator.of(context).push(
                MaterialPageRoute<void>(
                  builder: (_) =>
                      LeaveValidationScreen(repository: widget.repository),
                ),
              );
            },
          ),
      ],
      floatingActionButton: FloatingActionButton(
        onPressed: () async {
          final created = await Navigator.of(context).push<bool>(
            MaterialPageRoute<bool>(
              builder: (_) => LeaveRequestScreen(repository: widget.repository),
            ),
          );
          if (created == true) _reload();
        },
        child: const Icon(Icons.add),
      ),
      body: RefreshIndicator(
        onRefresh: () async => _reload(),
        child: FutureBuilder<List<LeaveRequest>>(
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
                  Center(child: Text(l10n.t('congesEmpty'))),
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
                    title: Text(item.type),
                    subtitle: Text('${item.from ?? ''} → ${item.to ?? ''}'),
                    trailing: Text(item.status),
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
