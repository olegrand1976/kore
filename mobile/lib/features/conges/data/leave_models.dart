class LeaveRequest {
  const LeaveRequest({
    required this.id,
    required this.type,
    required this.status,
    this.from,
    this.to,
    this.motif = '',
  });

  final String id;
  final String type;
  final String status;
  final String? from;
  final String? to;
  final String motif;

  factory LeaveRequest.fromJson(Map<String, dynamic> json) {
    final period = json['Period'] as Map<String, dynamic>? ??
        json['period'] as Map<String, dynamic>?;
    return LeaveRequest(
      id: (json['id'] ?? json['ID'])?.toString() ?? '',
      type: (json['type'] ?? json['Type'])?.toString() ?? '',
      status: (json['status'] ?? json['Status'])?.toString() ?? '',
      from: period?['From']?.toString() ?? period?['from']?.toString(),
      to: period?['To']?.toString() ?? period?['to']?.toString(),
      motif: (json['motif'] ?? json['Motif'])?.toString() ?? '',
    );
  }
}

class LeaveBalance {
  const LeaveBalance({
    required this.type,
    required this.acquired,
    required this.taken,
    required this.remaining,
  });

  final String type;
  final double acquired;
  final double taken;
  final double remaining;

  factory LeaveBalance.fromJson(Map<String, dynamic> json) {
    return LeaveBalance(
      type: (json['type'] ?? json['Type'])?.toString() ?? '',
      acquired: _toDouble(json['acquired'] ?? json['Acquired']),
      taken: _toDouble(json['taken'] ?? json['Taken']),
      remaining: _toDouble(json['remaining'] ?? json['Remaining']),
    );
  }

  static double _toDouble(Object? value) {
    if (value is num) return value.toDouble();
    return double.tryParse(value?.toString() ?? '') ?? 0;
  }
}
