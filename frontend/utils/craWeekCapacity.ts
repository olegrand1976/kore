/**
 * Week capacity for CRA validation: mission planned week minutes when set,
 * otherwise org day capacity × days present in the week tab.
 */
export function resolveWeekCapacityMinutes(opts: {
  weekDayCount: number
  dayCapacityMinutes: number
  plannedWeekMinutes?: number | null
}): number {
  const planned = opts.plannedWeekMinutes
  if (planned != null && Number.isFinite(planned) && planned > 0) {
    return Math.round(planned)
  }
  const dayCap = opts.dayCapacityMinutes > 0 ? opts.dayCapacityMinutes : 8 * 60
  const days = Math.max(0, opts.weekDayCount)
  return days * dayCap
}

export function hoursInputToMinutes(raw: string | number | null | undefined): number | null {
  if (raw === null || raw === undefined || raw === '') return null
  const n = typeof raw === 'number' ? raw : Number(String(raw).replace(',', '.'))
  if (!Number.isFinite(n) || n <= 0) return null
  return Math.round(n * 60)
}

export function minutesToHoursInput(minutes: number | null | undefined): string {
  if (minutes == null || !Number.isFinite(minutes) || minutes <= 0) return ''
  const hours = minutes / 60
  return Number.isInteger(hours) ? String(hours) : String(Math.round(hours * 100) / 100)
}
