import { formatHoursValue } from '~/utils/craDuration'

/** Parse "8:30", "8h30", "08.30" → minutes from midnight, or null. */
export function parseClockToMinutes(raw: string): number | null {
  const cleaned = raw.trim().toLowerCase().replace(/\s+/g, '')
  if (!cleaned) return null
  const match = cleaned.match(/^(\d{1,2})[:h.](\d{2})$/) ?? cleaned.match(/^(\d{1,2})$/)
  if (!match) return null
  const hours = Number(match[1])
  const minutes = match[2] != null ? Number(match[2]) : 0
  if (!Number.isFinite(hours) || !Number.isFinite(minutes)) return null
  if (hours < 0 || hours > 23 || minutes < 0 || minutes > 59) return null
  return hours * 60 + minutes
}

export type HoursCalculatorInput = {
  start: string
  end: string
  /** Pause in minutes (non-negative). */
  breakMinutes?: number
}

export type HoursCalculatorResult =
  | { ok: true; minutes: number; hours: number; hoursLabel: string }
  | { ok: false; reason: 'invalid_time' | 'end_before_start' | 'break_exceeds' | 'zero_or_negative' }

/**
 * Compute worked duration from a single time range minus an optional break.
 * Does not wrap past midnight (overnight shifts are rejected as end_before_start).
 */
export function calculateHoursFromRange(input: HoursCalculatorInput): HoursCalculatorResult {
  const start = parseClockToMinutes(input.start)
  const end = parseClockToMinutes(input.end)
  if (start == null || end == null) return { ok: false, reason: 'invalid_time' }
  if (end <= start) return { ok: false, reason: 'end_before_start' }
  const breakMinutes = Math.max(0, Math.round(Number(input.breakMinutes) || 0))
  if (breakMinutes >= end - start) return { ok: false, reason: 'break_exceeds' }
  const minutes = end - start - breakMinutes
  if (minutes <= 0) return { ok: false, reason: 'zero_or_negative' }
  const hours = minutes / 60
  return { ok: true, minutes, hours, hoursLabel: formatHoursValue(hours) }
}

/** Theoretical week capacity for 5 working days (minutes → hours label). */
export function theoreticalWeekHoursLabel(dayCapacityMinutes: number, workingDays = 5): string {
  const day = Math.max(0, Number(dayCapacityMinutes) || 0)
  return formatHoursValue((day * workingDays) / 60)
}

/** Clamp hours to [0, maxHours] and format for the CRA hours field. */
export function clampHoursLabel(hours: number, maxHours: number): string {
  const capped = Math.max(0, Math.min(maxHours, hours))
  return formatHoursValue(capped)
}
