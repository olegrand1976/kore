import type { CraLine, CraWeek } from '~/stores/cra'
import { computeWeekDays, hoursToMinutes, minutesToHoursLabel } from '~/composables/useWeekCalendar'

export type ActivityRow = {
  key: string
  sourceType: string
  sourceId: string
  day: string
  hours: string
  comment: string
  origin: string
  billable: boolean
}

export const buildKey = (sourceType: string, sourceId: string, day: string) => `${sourceType}:${sourceId}:${day}`

export function useWeekRows(
  week: Ref<CraWeek | undefined>,
  weekNumber: Ref<number>,
  month: Ref<string>,
  weekStartDay: Ref<number>
) {
  const weekDays = computed(() => computeWeekDays(month.value, weekNumber.value, weekStartDay.value))

  const rowsByDay = computed(() => {
    const map = new Map<string, ActivityRow[]>()
    for (const day of weekDays.value) {
      map.set(day, [])
    }
    for (const line of week.value?.lines ?? []) {
      const day = line.day.slice(0, 10)
      if (!map.has(day)) continue
      map.get(day)!.push({
        key: buildKey(line.sourceType, line.sourceId, day),
        sourceType: line.sourceType,
        sourceId: line.sourceId,
        day,
        hours: line.duration > 0 ? minutesToHoursLabel(line.duration) : '',
        comment: line.comment ?? '',
        origin: line.origin ?? 'manual',
        billable: line.billable ?? true
      })
    }
    for (const [day, rows] of map) {
      if (rows.length === 0) {
        map.set(day, [{
          key: buildKey('manual', 'default', day),
          sourceType: 'manual',
          sourceId: 'default',
          day,
          hours: '',
          comment: '',
          origin: 'manual',
          billable: true
        }])
      }
    }
    return map
  })

  const dayTotals = computed(() => {
    const totals = new Map<string, number>()
    for (const [day, rows] of rowsByDay.value) {
      totals.set(day, rows.reduce((sum, r) => sum + hoursToMinutes(r.hours), 0))
    }
    return totals
  })

  const weekTotalMinutes = computed(() => {
    let total = 0
    for (const mins of dayTotals.value.values()) total += mins
    return total
  })

  const toSaveLines = (rows: ActivityRow[]): CraLine[] => {
    const existing = week.value?.lines ?? []
    return rows.flatMap((r) => {
      const duration = hoursToMinutes(r.hours)
      const hadLine = existing.some(
        (line) =>
          line.sourceType === r.sourceType &&
          line.sourceId === r.sourceId &&
          line.day.slice(0, 10) === r.day
      )
      if (duration <= 0 && !hadLine) return []
      return [{
        sourceType: r.sourceType,
        sourceId: r.sourceId,
        day: r.day,
        duration,
        comment: r.comment,
        origin: 'manual',
        billable: r.billable
      }]
    })
  }

  return { weekDays, rowsByDay, dayTotals, weekTotalMinutes, toSaveLines, buildKey }
}
