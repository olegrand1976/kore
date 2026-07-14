import type { BudgetItem } from '~/composables/useBudget'
import type { LeaveBalance, LeaveRequest, LeaveTypeConfig } from '~/composables/useLeave'
import { leaveDayCount, pickFrom, pickLeaveTypeCode, pickLeaveTypeLabel, pickSortOrder, pickTracksBalance, pickTo } from '~/composables/useLeave'
import type { TmaDemand } from '~/composables/useTma'
import { currentMonthKey } from '~/composables/useCraStatus'

type CraTimesheet = {
  month?: string
  Month?: string
  status?: string
  Status?: string
  totalMinutes?: number
  TotalMinutes?: number
  weeksSubmitted?: number
  WeeksSubmitted?: number
  weeksTotal?: number
  WeeksTotal?: number
  prefillRatio?: number
  PrefillRatio?: number
}

export function pickLeaveStatus(item: LeaveRequest) {
  return item.status ?? item.Status ?? ''
}

export function countLeaveByStatus(items: LeaveRequest[], status: string) {
  return items.filter((item) => pickLeaveStatus(item) === status).length
}

function findLeaveBalance(balances: LeaveBalance[], type: string) {
  const row = balances.find((b) => (b.type ?? b.Type) === type)
  if (!row) {
    return { acquired: 0, taken: 0, remaining: null as number | null }
  }
  return {
    acquired: row.acquired ?? row.Acquired ?? 0,
    taken: row.taken ?? row.Taken ?? 0,
    remaining: row.remaining ?? row.Remaining ?? row.balance ?? row.Balance ?? null
  }
}

export function leaveMetrics(items: LeaveRequest[], balances: LeaveBalance[], leaveTypes: LeaveTypeConfig[] = []) {
  const today = new Date()
  today.setHours(0, 0, 0, 0)

  let approvedDays = 0
  let upcomingDays = 0

  for (const item of items) {
    if (pickLeaveStatus(item) !== 'valide') continue
    const from = pickFrom(item)
    const to = pickTo(item)
    const days = leaveDayCount(from, to)
    approvedDays += days

    const end = new Date(to)
    if (Number.isNaN(end.getTime()) || end < today) continue
    const start = new Date(from)
    const effectiveFrom = start < today ? today : start
    upcomingDays += leaveDayCount(effectiveFrom.toISOString().slice(0, 10), to)
  }

  const trackTypes = leaveTypes.length > 0
    ? leaveTypes
        .filter((item) => item.active ?? item.Active ?? true)
        .filter((item) => pickTracksBalance(item))
        .sort((a, b) => pickSortOrder(a) - pickSortOrder(b))
    : []

  const balanceKpis = (trackTypes.length > 0 ? trackTypes : [
    { code: 'conges_payes', Code: 'conges_payes' },
    { code: 'rtt', Code: 'rtt' }
  ] as LeaveTypeConfig[]).slice(0, 2).map((item) => {
    const code = pickLeaveTypeCode(item)
    const balance = findLeaveBalance(balances, code)
    return {
      code,
      label: pickLeaveTypeLabel(item),
      ...balance
    }
  })

  const cp = balanceKpis[0] ?? { acquired: 0, taken: 0, remaining: null as number | null }
  const rtt = balanceKpis[1] ?? { acquired: 0, taken: 0, remaining: null as number | null }

  return {
    total: items.length,
    pending: countLeaveByStatus(items, 'en_attente'),
    approved: countLeaveByStatus(items, 'valide'),
    rejected: countLeaveByStatus(items, 'refuse'),
    approvedDays,
    upcomingDays,
    balanceKpis,
    cpRemaining: cp.remaining,
    cpTaken: cp.taken,
    cpAcquired: cp.acquired,
    rttRemaining: rtt.remaining,
    rttTaken: rtt.taken,
    rttAcquired: rtt.acquired
  }
}

export function pickTmaStatus(d: TmaDemand) {
  return d.status ?? d.Status ?? ''
}

export function countTmaOpen(items: TmaDemand[]) {
  const open = new Set(['ouverte', 'affectee', 'en_cours', 'rework'])
  return items.filter((d) => open.has(pickTmaStatus(d))).length
}

export function countTmaByStatus(items: TmaDemand[], status: string) {
  return items.filter((d) => pickTmaStatus(d) === status).length
}

export function countCraByStatus(items: CraTimesheet[], status: string) {
  return items.filter((ts) => (ts.status ?? ts.Status) === status).length
}

export function craCurrentMonthStatus(items: CraTimesheet[]) {
  const key = currentMonthKey()
  const current = items.find((ts) => (ts.month ?? ts.Month) === key)
  return current?.status ?? current?.Status ?? null
}

export function isCraMonthIncomplete(items: CraTimesheet[]) {
  const key = currentMonthKey()
  const current = items.find((ts) => (ts.month ?? ts.Month) === key)
  if (!current) return true
  const status = current.status ?? current.Status ?? 'Brouillon'
  if (status === 'Définitif') return false
  const submitted = current.weeksSubmitted ?? current.WeeksSubmitted ?? 0
  const total = current.weeksTotal ?? current.WeeksTotal ?? 0
  const minutes = current.totalMinutes ?? current.TotalMinutes ?? 0
  if (total > 0 && submitted < total) return true
  return minutes <= 0
}

export function craPrefillRatioForMonth(items: CraTimesheet[]) {
  const key = currentMonthKey()
  const current = items.find((ts) => (ts.month ?? ts.Month) === key)
  if (!current) return null
  const ratio = current.prefillRatio ?? current.PrefillRatio
  return typeof ratio === 'number' ? ratio : null
}

export function budgetMetrics(budgets: BudgetItem[]) {
  let plannedDays = 0
  let consumedDays = 0
  let overrun = 0
  for (const b of budgets) {
    const planned = b.planned?.days ?? b.Planned?.Days ?? 0
    const consumed = b.consumed?.days ?? b.Consumed?.Days ?? 0
    plannedDays += planned
    consumedDays += consumed
    if (planned > 0 && consumed > planned) overrun += 1
  }
  return { total: budgets.length, plannedDays, consumedDays, overrun }
}

export function consumptionPct(consumed: number, planned: number, cap = true) {
  if (planned <= 0) return consumed > 0 ? 100 : 0
  const pct = Math.round((consumed / planned) * 100)
  return cap ? Math.min(100, pct) : pct
}

export type ChartBarItem = {
  key: string
  label: string
  value: number
  tone?: 'gold' | 'blue' | 'success' | 'warn' | 'muted'
}

export type CraMonthItem = {
  key: string
  label: string
  status: string | null
}

function groupByStatus<T>(
  items: T[],
  pickStatus: (item: T) => string,
  labelFn: (status: string) => string,
  toneFn?: (status: string) => ChartBarItem['tone']
): ChartBarItem[] {
  const counts = new Map<string, number>()
  for (const item of items) {
    const status = pickStatus(item) || 'unknown'
    counts.set(status, (counts.get(status) ?? 0) + 1)
  }
  return [...counts.entries()]
    .map(([key, value]) => ({
      key,
      label: labelFn(key),
      value,
      tone: toneFn?.(key)
    }))
    .sort((a, b) => b.value - a.value)
}

export function tmaStatusSeries(
  items: TmaDemand[],
  labelFn: (status: string) => string
): ChartBarItem[] {
  const open = new Set(['ouverte', 'affectee', 'en_cours', 'rework'])
  return groupByStatus(items, pickTmaStatus, labelFn, (status) => {
    if (status === 'resolue') return 'success'
    if (open.has(status)) return 'blue'
    if (status === 'en_attente_creation') return 'warn'
    return 'muted'
  })
}

export function leaveStatusSeries(
  items: LeaveRequest[],
  labelFn: (status: string) => string
): ChartBarItem[] {
  return groupByStatus(items, pickLeaveStatus, labelFn, (status) => {
    if (status === 'valide') return 'success'
    if (status === 'en_attente') return 'warn'
    if (status === 'refuse') return 'muted'
    return 'blue'
  })
}

export function budgetConsumptionSeries(
  budgets: BudgetItem[],
  labelFn: (budget: BudgetItem) => string,
  limit = 6
): ChartBarItem[] {
  return budgets
    .map((b) => {
      const planned = b.planned?.days ?? b.Planned?.Days ?? 0
      const consumed = b.consumed?.days ?? b.Consumed?.Days ?? 0
      const pct = consumptionPct(consumed, planned, false)
      return {
        key: b.id ?? b.ID ?? '',
        label: labelFn(b),
        value: pct,
        tone: (pct > 100 ? 'warn' : pct >= 80 ? 'gold' : 'success') as ChartBarItem['tone']
      }
    })
    .sort((a, b) => b.value - a.value)
    .slice(0, limit)
}

export function craMonthSeries(
  items: CraTimesheet[],
  locale: string,
  months = 6
): CraMonthItem[] {
  const loc = locale === 'en' ? 'en-US' : 'fr-FR'
  const now = new Date()
  const result: CraMonthItem[] = []

  for (let i = months - 1; i >= 0; i--) {
    const d = new Date(now.getFullYear(), now.getMonth() - i, 1)
    const key = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`
    const match = items.find((ts) => (ts.month ?? ts.Month) === key)
    result.push({
      key,
      label: d.toLocaleDateString(loc, { month: 'short' }),
      status: match?.status ?? match?.Status ?? null
    })
  }

  return result
}
