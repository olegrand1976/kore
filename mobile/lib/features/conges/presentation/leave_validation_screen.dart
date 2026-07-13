import 'package:flutter/material.dart';

import '../../../core/l10n/app_localizations.dart';
import '../data/leave_models.dart';
import '../data/leave_repository.dart';

class LeaveValidationScreen extends StatefulWidget {
  const LeaveValidationScreen({super.key, required this.repository});

  final LeaveRepository repository;

  @override
  State<LeaveValidationScreen> createState() => _LeaveValidationScreenState();
}

class _LeaveValidationScreenState extends State<LeaveValidationScreen> {
  late Future<List<LeaveRequest>> _future;

  @override
  void initState() {
    super.initState();
    _reload();
  }

  void _reload() {
    setState(() => _future = widget.repository.listPendingForValidation());
  }

  Future<void> _decide(LeaveRequest request, {required bool approve}) async {
    if (approve) {
      await widget.repository.approve(request.id);
    } else {
      await widget.repository.reject(request.id);
    }
    _reload();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Scaffold(
      appBar: AppBar(title: Text(l10n.t('congesValidation'))),
      body: SafeArea(
        child: FutureBuilder<List<LeaveRequest>>(
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
                final item = items[index];
                return Card(
                  child: Padding(
                    padding: const EdgeInsets.all(12),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(item.type, style: Theme.of(context).textTheme.titleMedium),
                        const SizedBox(height: 4),
                        Text('${item.from ?? ''} → ${item.to ?? ''}'),
                        if (item.motif.isNotEmpty) Text(item.motif),
                        const SizedBox(height: 12),
                        Row(
                          children: [
                            Expanded(
                              child: OutlinedButton(
                                onPressed: () => _decide(item, approve: false),
                                child: Text(l10n.t('congesReject')),
                              ),
                            ),
                            const SizedBox(width: 8),
                            Expanded(
                              child: ElevatedButton(
                                onPressed: () => _decide(item, approve: true),
                                child: Text(l10n.t('congesApprove')),
                              ),
                            ),
                          ],
                        ),
                      ],
                    ),
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
