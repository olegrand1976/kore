import { isAbsenceSourceType } from '~/utils/craAbsence'

export type DayRowLike = {
  key?: string
  sourceType: string
  hours: string
  origin?: string
}

export function withManualOrigin<T extends { origin: string }>(row: T): T {
  return row.origin === 'manual' ? row : { ...row, origin: 'manual' }
}

export function unlockHolidayPrefillRows<T extends DayRowLike & { origin: string }>(rows: T[]): T[] {
  return rows.map((row) =>
    row.sourceType === 'holiday' && row.origin === 'prefill' ? withManualOrigin(row) : row
  )
}

/** Journée entièrement en absence (bandeau « jour non presté »). */
export function isFullAbsenceDay(
  rows: DayRowLike[],
  hoursToMinutes: (hours: string) => number
): boolean {
  const absenceRows = rows.filter((row) => isAbsenceSourceType(row.sourceType))
  if (absenceRows.length === 0) return false
  if (rows.some((row) => !isAbsenceSourceType(row.sourceType) && hoursToMinutes(row.hours) > 0)) {
    return false
  }
  if (rows.some((row) => !isAbsenceSourceType(row.sourceType))) return false
  if (absenceRows.some((row) => hoursToMinutes(row.hours) > 0)) return false
  return true
}

export function partialAbsenceHoursLabel(capacityMinutes: number): string {
  const halfDayHours = capacityMinutes / 2 / 60
  return Number.isInteger(halfDayHours) ? String(halfDayHours) : halfDayHours.toFixed(1)
}

export function rowsSnapshot(rows: DayRowLike[]): string {
  return rows
    .map((row) => {
      const workRef = 'workRefType' in row && 'workRefId' in row
        ? `${(row as { workRefType?: string }).workRefType ?? ''}:${(row as { workRefId?: string }).workRefId ?? ''}`
        : ''
      return `${row.key ?? row.sourceType}:${row.hours}:${row.origin ?? ''}:${workRef}`
    })
    .join('|')
}
