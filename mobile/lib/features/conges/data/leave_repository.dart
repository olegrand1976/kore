import '../../../core/api/api_client.dart';
import 'leave_models.dart';

class LeaveRepository {
  LeaveRepository(this._api);

  final ApiClient _api;

  Future<List<LeaveRequest>> listRequests({String? status}) async {
    return _api.getData(
      '/leave-requests',
      query: status != null ? {'status': status} : null,
      parser: (json) => (json as List<dynamic>)
          .map((e) => LeaveRequest.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  Future<List<LeaveRequest>> listPendingForValidation() async {
    return listRequests(status: 'en_attente');
  }

  Future<LeaveRequest> createRequest({
    required String type,
    required DateTime from,
    required DateTime to,
    String motif = '',
  }) async {
    return _api.postData(
      '/leave-requests',
      body: {
        'type': type,
        'from': from.toUtc().toIso8601String(),
        'to': to.toUtc().toIso8601String(),
        'motif': motif,
      },
      parser: (json) => LeaveRequest.fromJson(json as Map<String, dynamic>),
    );
  }

  Future<void> approve(String id) async {
    await _api.post('/leave-requests/$id/approve');
  }

  Future<void> reject(String id) async {
    await _api.post('/leave-requests/$id/reject');
  }

  Future<List<LeaveBalance>> listBalances() async {
    return _api.getData(
      '/leave-balances',
      parser: (json) => (json as List<dynamic>)
          .map((e) => LeaveBalance.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }
}
