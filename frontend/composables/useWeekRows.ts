import type { CraLine, CraWeek } from '~/stores/cra'
import { computeWeekDays, hoursToMinutes, minutesToHoursLabel } from '~/composables/useWeekCalendar'

export type ActivityRow = {
  key: string
  id?: string
  sourceType: string
  sourceId: string
  day: string
  hours: string
  comment: string
  origin: string
  billable: boolean
  workRefType?: string
  workRefId?: string
}

export const buildKey = (sourceType: string, sourceId: string, day: string) => `${sourceType}:${sourceId}:${day}`

export const newRowKey = () => crypto.randomUUID()

export const rowKeyFromLine = (line: Pick<CraLine, 'id' | 'sourceType' | 'sourceId' | 'day'>, index: number) =>
  line.id || `${line.sourceType}:${line.sourceId}:${line.day}:${index}`

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
    for (const [index, line] of (week.value?.lines ?? []).entries()) {
      const day = line.day.slice(0, 10)
      if (!map.has(day)) continue
      map.get(day)!.push({
        key: rowKeyFromLine(line, index),
        id: line.id,
        sourceType: line.sourceType,
        sourceId: line.sourceId,
        day,
        hours: line.duration > 0 ? minutesToHoursLabel(line.duration) : '',
        comment: line.comment ?? '',
        origin: line.origin ?? 'manual',
        billable: line.billable ?? true,
        workRefType: line.workRefType,
        workRefId: line.workRefId
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

  const toSaveLines = (rows: ActivityRow[]): CraLine[] =>
    rows.flatMap((r) => {
      const duration = hoursToMinutes(r.hours)
      if (duration <= 0) return []
      return [{
        id: r.id,
        sourceType: r.sourceType,
        sourceId: r.sourceId,
        day: r.day,
        duration,
        comment: r.comment,
        origin: r.origin ?? 'manual',
        billable: r.billable,
        workRefType: r.workRefType,
        workRefId: r.workRefId
      }]
    })

  return { weekDays, rowsByDay, dayTotals, weekTotalMinutes, toSaveLines, buildKey }
}
