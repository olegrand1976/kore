import '../../../core/api/api_client.dart';
import 'cra_models.dart';

class CraRepository {
  CraRepository(this._api);

  final ApiClient _api;

  Future<List<TimesheetSummary>> listRecent({int limit = 24}) async {
    return _api.getData(
      '/timesheets/recent',
      query: {'limit': limit.toString()},
      parser: (json) => (json as List<dynamic>)
          .map((e) => TimesheetSummary.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  Future<Timesheet> getByMonth(String month) async {
    return _api.getData(
      '/timesheets',
      query: {'month': month},
      parser: (json) => Timesheet.fromJson(json as Map<String, dynamic>),
    );
  }

  Future<Timesheet> saveWeek({
    required String timesheetId,
    required int week,
    required List<TimeLine> lines,
  }) async {
    return _api.putData(
      '/timesheets/$timesheetId/weeks/$week',
      body: {'lines': lines.map((l) => l.toJson()).toList()},
      parser: (json) => Timesheet.fromJson(json as Map<String, dynamic>),
    );
  }

  Future<void> submitWeek({
    required String timesheetId,
    required int week,
  }) async {
    await _api.post('/timesheets/$timesheetId/weeks/$week/submit');
  }

  Future<void> validateFinal(String timesheetId) async {
    await _api.post('/timesheets/$timesheetId/validate');
  }
}
