const LEGACY_MONTH_FILTER_KEY = 'month'

/** Migrate persisted list filter `month` (YYYY-MM) → `periodYear` + `periodMonth`. */
export function migrateLegacyMonthFilter(filterValues: Record<string, string>): boolean {
  const legacy = filterValues[LEGACY_MONTH_FILTER_KEY]?.trim()
  if (!legacy) {
    if (LEGACY_MONTH_FILTER_KEY in filterValues) {
      delete filterValues[LEGACY_MONTH_FILTER_KEY]
    }
    return false
  }
  const match = legacy.match(/^(\d{4})-(\d{2})$/)
  if (!match) {
    delete filterValues[LEGACY_MONTH_FILTER_KEY]
    return false
  }
  filterValues.periodYear = match[1]
  filterValues.periodMonth = match[2]
  delete filterValues[LEGACY_MONTH_FILTER_KEY]
  return true
}

/** Match a CRA month key (YYYY-MM) against optional month (MM) and year (YYYY) filters. */
export function matchPeriodMonthYear(
  monthKey: string,
  periodMonth: string,
  periodYear: string
): boolean {
  const key = String(monthKey ?? '').trim()
  if (!/^\d{4}-\d{2}$/.test(key)) {
    return !periodMonth && !periodYear
  }
  if (periodYear && key.slice(0, 4) !== periodYear) return false
  if (periodMonth && key.slice(5, 7) !== periodMonth) return false
  return true
}

export function buildMonthFilterOptions(
  locale: string
): Array<{ value: string; label: string }> {
  const loc = locale === 'en' ? 'en-US' : 'fr-FR'
  return Array.from({ length: 12 }, (_, i) => {
    const value = String(i + 1).padStart(2, '0')
    const label = new Date(2000, i, 1).toLocaleDateString(loc, { month: 'long' })
    return { value, label }
  })
}

export function buildYearFilterOptions(
  monthKeys: string[],
  now: Date = new Date()
): Array<{ value: string; label: string }> {
  const years = new Set<string>()
  years.add(String(now.getFullYear()))
  for (const key of monthKeys) {
    const y = String(key ?? '').slice(0, 4)
    if (/^\d{4}$/.test(y)) years.add(y)
  }
  return [...years]
    .sort((a, b) => Number(b) - Number(a))
    .map((value) => ({ value, label: value }))
}
