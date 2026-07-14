import type { CraWeek } from '~/stores/cra'
import { computeMonthWeeks } from '~/composables/useWeekCalendar'
import { safeMinutes } from '~/utils/craDuration'

const DEFAULT_DAY_CAPACITY_MINUTES = 8 * 60

export function useCraMonthStats(
  weeks: Ref<CraWeek[]>,
  month: Ref<string>,
  weekStartDay: Ref<number>,
  dayCapacityMinutes: Ref<number> | number = DEFAULT_DAY_CAPACITY_MINUTES
) {
  const capacityPerDay = computed(() => {
    const raw = typeof dayCapacityMinutes === 'number' ? dayCapacityMinutes : dayCapacityMinutes.value
    const value = safeMinutes(raw)
    return value > 0 ? value : DEFAULT_DAY_CAPACITY_MINUTES
  })

  const monthWeekTabs = computed(() => computeMonthWeeks(month.value, weekStartDay.value))

  const totalMinutes = computed(() => {
    let total = 0
    for (const week of weeks.value) {
      for (const line of week.lines) {
        total += safeMinutes(line.duration)
      }
    }
    return total
  })

  const capacityMinutes = computed(() =>
    monthWeekTabs.value.reduce((sum, tab) => sum + tab.days.length * capacityPerDay.value, 0)
  )

  const weeksSubmitted = computed(() => weeks.value.filter((w) => w.submittedAt).length)
  const weeksTotal = computed(() => monthWeekTabs.value.length)

  const prefillRatio = computed(() => {
    let prefill = 0
    let total = 0
    for (const week of weeks.value) {
      for (const line of week.lines) {
        const minutes = safeMinutes(line.duration)
        if (minutes <= 0) continue
        total += minutes
        if (line.origin === 'prefill') prefill += minutes
      }
    }
    return total > 0 ? Math.round((prefill / total) * 100) : 0
  })

  const progress = computed(() => {
    if (capacityMinutes.value <= 0) return 0
    return Math.min(100, Math.round((totalMinutes.value / capacityMinutes.value) * 100))
  })

  return { totalMinutes, capacityMinutes, weeksSubmitted, weeksTotal, prefillRatio, progress }
}
