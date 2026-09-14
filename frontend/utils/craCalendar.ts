/** Saturday / Sunday for an ISO date `YYYY-MM-DD` (noon UTC avoids DST edge cases). */
export function isWeekend(day: string): boolean {
  const dow = new Date(`${day.slice(0, 10)}T12:00:00Z`).getUTCDay()
  return dow === 0 || dow === 6
}

export type IncompleteDayRow = {
  hours: string
  sourceType: string
}

/**
 * Working days (Mon–Fri) in the week tab with zero logged minutes.
 * Weekends and days that already have a holiday source line are excluded.
 */
export function collectIncompleteWorkingDays(
  days: string[],
  rowsByDay: Map<string, IncompleteDayRow[]>,
  toMinutes: (hours: string) => number
): string[] {
  const missing: string[] = []
  for (const day of days) {
    if (isWeekend(day)) continue
    const rows = rowsByDay.get(day) ?? []
    if (rows.some((row) => row.sourceType === 'holiday')) continue
    const total = rows.reduce((sum, row) => sum + toMinutes(row.hours), 0)
    if (total <= 0) missing.push(day)
  }
  return missing
}

export function countWorkingDays(days: string[]): number {
  return days.reduce((n, day) => (isWeekend(day) ? n : n + 1), 0)
}
