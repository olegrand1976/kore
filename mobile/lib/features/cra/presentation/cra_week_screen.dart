import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import '../../../core/l10n/app_localizations.dart';
import '../data/cra_models.dart';
import '../data/cra_repository.dart';

class CraWeekScreen extends StatefulWidget {
  const CraWeekScreen({
    super.key,
    required this.repository,
    required this.timesheet,
  });

  final CraRepository repository;
  final Timesheet timesheet;

  @override
  State<CraWeekScreen> createState() => _CraWeekScreenState();
}

class _CraWeekScreenState extends State<CraWeekScreen> {
  late int _selectedWeek;
  bool _saving = false;

  @override
  void initState() {
    super.initState();
    _selectedWeek = widget.timesheet.weeks.isNotEmpty
        ? widget.timesheet.weeks.first.weekNumber
        : 1;
  }

  WeekEntry? get _weekEntry {
    for (final w in widget.timesheet.weeks) {
      if (w.weekNumber == _selectedWeek) return w;
    }
    return null;
  }

  Future<void> _saveSampleLine() async {
    setState(() => _saving = true);
    final today = DateFormat('yyyy-MM-dd').format(DateTime.now());
    try {
      await widget.repository.saveWeek(
        timesheetId: widget.timesheet.id,
        week: _selectedWeek,
        lines: [
          TimeLine(
            day: today,
            duration: 480,
            sourceType: 'manual',
            sourceId: 'mobile',
            comment: 'Mobile entry',
          ),
        ],
      );
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(AppLocalizations.of(context).t('craSave'))),
        );
      }
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  Future<void> _submitWeek() async {
    setState(() => _saving = true);
    try {
      await widget.repository.submitWeek(
        timesheetId: widget.timesheet.id,
        week: _selectedWeek,
      );
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(AppLocalizations.of(context).t('craSubmit')),
          ),
        );
      }
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final week = _weekEntry;
    final lines = week?.lines ?? [];

    return Scaffold(
      appBar: AppBar(
        title: Text('${widget.timesheet.month} — ${l10n.t('craWeek')} $_selectedWeek'),
      ),
      body: SafeArea(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Padding(
              padding: const EdgeInsets.all(16),
              child: Row(
                children: [
                  IconButton(
                    onPressed: _selectedWeek > 1
                        ? () => setState(() => _selectedWeek--)
                        : null,
                    icon: const Icon(Icons.chevron_left),
                  ),
                  Expanded(
                    child: Text(
                      '${l10n.t('craStatus')}: ${widget.timesheet.status}',
                      textAlign: TextAlign.center,
                    ),
                  ),
                  IconButton(
                    onPressed: () => setState(() => _selectedWeek++),
                    icon: const Icon(Icons.chevron_right),
                  ),
                ],
              ),
            ),
            Expanded(
              child: lines.isEmpty
                  ? Center(child: Text(l10n.t('craEmpty')))
                  : ListView.builder(
                      padding: const EdgeInsets.symmetric(horizontal: 16),
                      itemCount: lines.length,
                      itemBuilder: (context, index) {
                        final line = lines[index];
                        return Card(
                          child: ListTile(
                            title: Text(line.day),
                            subtitle: Text(line.comment),
                            trailing: Text('${line.duration ~/ 60}h'),
                          ),
                        );
                      },
                    ),
            ),
            Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  ElevatedButton(
                    onPressed: _saving ? null : _saveSampleLine,
                    child: Text(l10n.t('craSave')),
                  ),
                  const SizedBox(height: 8),
                  OutlinedButton(
                    onPressed: _saving ? null : _submitWeek,
                    child: Text(l10n.t('craSubmit')),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
