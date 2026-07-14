import 'package:flutter/material.dart';

import '../../../core/l10n/app_localizations.dart';
import '../data/leave_repository.dart';

class LeaveRequestScreen extends StatefulWidget {
  const LeaveRequestScreen({super.key, required this.repository});

  final LeaveRepository repository;

  @override
  State<LeaveRequestScreen> createState() => _LeaveRequestScreenState();
}

class _LeaveRequestScreenState extends State<LeaveRequestScreen> {
  final _typeCtrl = TextEditingController(text: 'CP');
  DateTime _from = DateTime.now();
  DateTime _to = DateTime.now();
  bool _submitting = false;

  @override
  void dispose() {
    _typeCtrl.dispose();
    super.dispose();
  }

  Future<void> _pickDate({required bool isFrom}) async {
    final picked = await showDatePicker(
      context: context,
      initialDate: isFrom ? _from : _to,
      firstDate: DateTime.now().subtract(const Duration(days: 365)),
      lastDate: DateTime.now().add(const Duration(days: 730)),
    );
    if (picked == null) return;
    setState(() {
      if (isFrom) {
        _from = picked;
        if (_to.isBefore(_from)) _to = _from;
      } else {
        _to = picked;
      }
    });
  }

  String _fmt(DateTime d) =>
      '${d.year.toString().padLeft(4, '0')}-${d.month.toString().padLeft(2, '0')}-${d.day.toString().padLeft(2, '0')}';

  Future<void> _submit() async {
    setState(() => _submitting = true);
    try {
      await widget.repository.createRequest(
        type: _typeCtrl.text.trim(),
        from: _from,
        to: _to,
      );
      if (mounted) Navigator.of(context).pop(true);
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(AppLocalizations.of(context).t('errorGeneric'))),
        );
      }
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Scaffold(
      appBar: AppBar(title: Text(l10n.t('congesNew'))),
      body: SafeArea(
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            TextField(
              controller: _typeCtrl,
              decoration: InputDecoration(labelText: l10n.t('congesType')),
            ),
            const SizedBox(height: 16),
            ListTile(
              title: Text(l10n.t('congesFrom')),
              subtitle: Text(_fmt(_from)),
              trailing: const Icon(Icons.calendar_today),
              onTap: () => _pickDate(isFrom: true),
            ),
            ListTile(
              title: Text(l10n.t('congesTo')),
              subtitle: Text(_fmt(_to)),
              trailing: const Icon(Icons.calendar_today),
              onTap: () => _pickDate(isFrom: false),
            ),
            const SizedBox(height: 24),
            ElevatedButton(
              onPressed: _submitting ? null : _submit,
              child: _submitting
                  ? const SizedBox(
                      height: 20,
                      width: 20,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : Text(l10n.t('congesSubmit')),
            ),
          ],
        ),
      ),
    );
  }
}
