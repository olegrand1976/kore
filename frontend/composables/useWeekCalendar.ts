export type WeekTab = {
  weekNumber: number
  start: string
  end: string
  days: string[]
}

const DEFAULT_WEEK_START_DAY = 1

export function formatLocalDate(date: Date): string {
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, '0')
  const d = String(date.getDate()).padStart(2, '0')
  return `${y}-${m}-${d}`
}

function normalizeWeekStartDay(day: number): number {
  const value = Number(day)
  if (!Number.isFinite(value) || value < 0 || value > 6) return DEFAULT_WEEK_START_DAY
  return value
}

function parseMonth(month: string): { year: number; monthIndex: number } {
  const [y, m] = month.split('-').map(Number)
  return { year: y, monthIndex: m - 1 }
}

function weekRangeStart(year: number, monthIndex: number, weekNumber: number, weekStartDay: number): Date {
  const start = new Date(year, monthIndex, 1)
  while (start.getDay() !== weekStartDay) {
    start.setDate(start.getDate() - 1)
  }
  if (weekNumber > 1) {
    start.setDate(start.getDate() + (weekNumber - 1) * 7)
  }
  return start
}

export function computeWeekDays(month: string, weekNumber: number, weekStartDay = DEFAULT_WEEK_START_DAY): string[] {
  const { year, monthIndex } = parseMonth(month)
  const ws = normalizeWeekStartDay(weekStartDay)
  const start = weekRangeStart(year, monthIndex, weekNumber, ws)
  const days: string[] = []
  for (let i = 0; i < 7; i++) {
    const d = new Date(start)
    d.setDate(start.getDate() + i)
    if (d.getMonth() !== monthIndex) continue
    days.push(formatLocalDate(d))
  }
  return days
}

export function computeMonthWeeks(month: string, weekStartDay = DEFAULT_WEEK_START_DAY): WeekTab[] {
  const { year, monthIndex } = parseMonth(month)
  const ws = normalizeWeekStartDay(weekStartDay)
  const lastDay = new Date(year, monthIndex + 1, 0).getDate()
  const weeks: WeekTab[] = []
  let weekNumber = 1
  let start = weekRangeStart(year, monthIndex, 1, ws)

  while (weekNumber <= 6) {
    const days: string[] = []
    for (let i = 0; i < 7; i++) {
      const d = new Date(start)
      d.setDate(start.getDate() + i)
      if (d.getMonth() !== monthIndex) continue
      if (d.getDate() > lastDay) continue
      days.push(formatLocalDate(d))
    }
    if (days.length === 0) break
    weeks.push({
      weekNumber,
      start: days[0],
      end: days[days.length - 1],
      days
    })
    weekNumber++
    start = new Date(start)
    start.setDate(start.getDate() + 7)
  }

  return weeks.length > 0 ? weeks : [{ weekNumber: 1, start: `${month}-01`, end: `${month}-01`, days: [`${month}-01`] }]
}

export function weekNumberForDay(month: string, day: string, weekStartDay = DEFAULT_WEEK_START_DAY): number {
  const normalizedDay = day.slice(0, 10)
  const weeks = computeMonthWeeks(month, weekStartDay)
  for (const week of weeks) {
    if (week.days.includes(normalizedDay)) {
      return week.weekNumber
    }
  }
  return weeks[0]?.weekNumber ?? 1
}

export function minutesToHoursLabel(minutes: number): string {
  const value = Number(minutes)
  if (!Number.isFinite(value)) return '0'
  const h = value / 60
  return Number.isInteger(h) ? String(h) : h.toFixed(1)
}

export function hoursToMinutes(hours: string | number): number {
  const n = Number(hours)
  if (!Number.isFinite(n) || n <= 0) return 0
  return Math.round(n * 60)
}
