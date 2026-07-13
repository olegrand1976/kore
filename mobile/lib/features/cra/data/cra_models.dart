class TimesheetSummary {
  const TimesheetSummary({
    required this.id,
    required this.month,
    required this.status,
    this.userLogin,
    this.totalMinutes = 0,
    this.weeksSubmitted = 0,
  });

  final String id;
  final String month;
  final String status;
  final String? userLogin;
  final int totalMinutes;
  final int weeksSubmitted;

  factory TimesheetSummary.fromJson(Map<String, dynamic> json) {
    return TimesheetSummary(
      id: json['id']?.toString() ?? '',
      month: json['month']?.toString() ?? '',
      status: json['status']?.toString() ?? '',
      userLogin: json['userLogin']?.toString(),
      totalMinutes: json['totalMinutes'] as int? ?? 0,
      weeksSubmitted: json['weeksSubmitted'] as int? ?? 0,
    );
  }
}

class Timesheet {
  const Timesheet({
    required this.id,
    required this.month,
    required this.status,
    this.weeks = const [],
  });

  final String id;
  final String month;
  final String status;
  final List<WeekEntry> weeks;

  factory Timesheet.fromJson(Map<String, dynamic> json) {
    final weeksJson = json['weeks'] as List<dynamic>? ?? [];
    return Timesheet(
      id: json['id']?.toString() ?? '',
      month: json['month']?.toString() ?? '',
      status: json['status']?.toString() ?? '',
      weeks: weeksJson
          .map((w) => WeekEntry.fromJson(w as Map<String, dynamic>))
          .toList(),
    );
  }
}

class WeekEntry {
  const WeekEntry({
    required this.weekNumber,
    this.lines = const [],
    this.submittedAt,
  });

  final int weekNumber;
  final List<TimeLine> lines;
  final String? submittedAt;

  factory WeekEntry.fromJson(Map<String, dynamic> json) {
    final linesJson = json['lines'] as List<dynamic>? ?? [];
    return WeekEntry(
      weekNumber: json['weekNumber'] as int? ?? json['WeekNumber'] as int? ?? 0,
      submittedAt: json['submittedAt']?.toString(),
      lines: linesJson
          .map((l) => TimeLine.fromJson(l as Map<String, dynamic>))
          .toList(),
    );
  }
}

class TimeLine {
  const TimeLine({
    required this.day,
    required this.duration,
    this.sourceType = 'manual',
    this.sourceId = '',
    this.comment = '',
  });

  final String day;
  final int duration;
  final String sourceType;
  final String sourceId;
  final String comment;

  factory TimeLine.fromJson(Map<String, dynamic> json) {
    return TimeLine(
      day: json['day']?.toString() ?? json['Day']?.toString() ?? '',
      duration: json['duration'] as int? ??
          json['Duration']?['Minutes'] as int? ??
          0,
      sourceType: json['sourceType']?.toString() ?? 'manual',
      sourceId: json['sourceId']?.toString() ?? '',
      comment: json['comment']?.toString() ?? '',
    );
  }

  Map<String, dynamic> toJson() => {
        'sourceType': sourceType,
        'sourceId': sourceId,
        'day': day,
        'duration': duration,
        'comment': comment,
      };
}
