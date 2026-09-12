import { buildMonthFilterOptions } from '~/utils/craPeriodFilter'

/** Previous calendar month as YYYY-MM. */
export function previousMonthKey(now: Date = new Date()): string {
  const d = new Date(now.getFullYear(), now.getMonth() - 1, 1)
  const yy = d.getFullYear()
  const mm = String(d.getMonth() + 1).padStart(2, '0')
  return `${yy}-${mm}`
}

export function monthKeyFromParts(year: string, monthMM: string): string {
  return `${year}-${monthMM}`
}

function currentMonthKeyFromDate(now: Date): string {
  const y = now.getFullYear()
  const m = String(now.getMonth() + 1).padStart(2, '0')
  return `${y}-${m}`
}

/** True when monthKey is strictly before the current calendar month. */
export function isStrictlyPastMonth(monthKey: string, now: Date = new Date()): boolean {
  if (!/^\d{4}-\d{2}$/.test(monthKey)) return false
  return monthKey < currentMonthKeyFromDate(now)
}

export function buildCreatePeriodYearOptions(
  now: Date = new Date(),
  yearsBack = 5
): Array<{ value: string; label: string }> {
  const current = now.getFullYear()
  const years: string[] = []
  for (let y = current; y >= current - yearsBack; y--) {
    years.push(String(y))
  }
  return years.map((value) => ({ value, label: value }))
}

/** Month options for create-period modal: past months only for the selected year. */
export function buildCreatePeriodMonthOptions(
  year: string,
  locale: string,
  now: Date = new Date()
): Array<{ value: string; label: string }> {
  const all = buildMonthFilterOptions(locale)
  const current = currentMonthKeyFromDate(now)
  const currentYear = current.slice(0, 4)
  const currentMM = current.slice(5, 7)
  if (year > currentYear) return []
  if (year < currentYear) return all
  return all.filter((opt) => opt.value < currentMM)
}

/** Ensure year/month form a selectable past period; fallback to previous month. */
export function clampCreatePeriodSelection(
  year: string,
  monthMM: string,
  now: Date = new Date()
): { year: string; month: string } {
  const key = monthKeyFromParts(year, monthMM)
  if (isStrictlyPastMonth(key, now)) {
    return { year, month: monthMM }
  }
  const prev = previousMonthKey(now)
  return { year: prev.slice(0, 4), month: prev.slice(5, 7) }
}

/**
 * Resolve which userId to pass to GetOrCreate.
 * Validators may open/create a CRA for the list filter collaborator;
 * otherwise always the session user (empty = omit query param).
 */
export function resolveOpenTimesheetUserId(
  canValidate: boolean,
  sessionUserId: string,
  filterUserId: string
): string {
  if (!canValidate) return ''
  const target = String(filterUserId ?? '').trim()
  if (!target) return ''
  const session = String(sessionUserId ?? '').trim()
  if (session && target === session) return ''
  return target
}
