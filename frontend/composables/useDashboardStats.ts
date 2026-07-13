import {
  budgetConsumptionSeries,
  budgetMetrics,
  consumptionPct,
  craCurrentMonthStatus,
  craMonthSeries,
  countLeaveByStatus,
  countTmaOpen,
  leaveStatusSeries,
  tmaStatusSeries
} from '~/composables/useKpiMetrics'
import { currentMonthKey } from '~/composables/useCraStatus'
import type { BudgetItem } from '~/composables/useBudget'
import type { LeaveRequest } from '~/composables/useLeave'
import type { TmaDemand } from '~/composables/useTma'
import type { ModuleCode } from '~/composables/useEntitlements'

export type DashboardStats = {
  craCurrentStatus: string | null
  leavePending: number
  tmaOpen: number
  tmaTotal: number
  budgetOverrun: number
  budgetConsumptionPct: number
  pendingValidations: number
}

export type DashboardCharts = {
  tmaStatus: ReturnType<typeof tmaStatusSeries>
  budgetConsumption: ReturnType<typeof budgetConsumptionSeries>
  craMonths: ReturnType<typeof craMonthSeries>
  leaveStatus: ReturnType<typeof leaveStatusSeries>
}

export type DashboardStatErrors = {
  cra?: boolean
  conges?: boolean
  tma?: boolean
  budget?: boolean
}

export type DashboardLoadResult = {
  stats: DashboardStats
  charts: DashboardCharts
  errors: DashboardStatErrors
}

const emptyStats = (): DashboardStats => ({
  craCurrentStatus: null,
  leavePending: 0,
  tmaOpen: 0,
  tmaTotal: 0,
  budgetOverrun: 0,
  budgetConsumptionPct: 0,
  pendingValidations: 0
})

const emptyCharts = (): DashboardCharts => ({
  tmaStatus: [],
  budgetConsumption: [],
  craMonths: [],
  leaveStatus: []
})

export function useDashboardStats() {
  const { hasModule } = useEntitlements()
  const { canValidateConges } = usePermissions()
  const { statusLabel: craStatusLabel } = useCraStatus()
  const { list: listLeaves } = useLeave()
  const { list: listTma } = useTma()
  const { list: listBudgets, pickId: pickBudgetId } = useBudget()
  const { locale, t } = useI18n()

  const tmaStatusLabel = (status: string) => t(`dashboard.charts.status.tma.${status}`, status)
  const leaveStatusLabel = (status: string) => t(`dashboard.charts.status.leave.${status}`, status)
  const budgetLabel = (b: BudgetItem) => {
    const type = b.type ?? b.Type ?? 'budget'
    const id = pickBudgetId(b)
    return id ? `${type} · ${id.slice(0, 8)}` : type
  }

  const load = async (): Promise<DashboardLoadResult> => {
    const stats = emptyStats()
    const charts = emptyCharts()
    const errors: DashboardStatErrors = {}
    const tasks: Promise<void>[] = []

    if (hasModule('cra')) {
      tasks.push(
        $fetch<{ data?: unknown[] }>('/api/cra/timesheets/recent')
          .then((res) => {
            const items = (res?.data ?? []) as Array<{ status?: string; Status?: string; month?: string; Month?: string }>
            stats.craCurrentStatus = craCurrentMonthStatus(items)
            charts.craMonths = craMonthSeries(items, locale.value)
          })
          .catch(() => {
            errors.cra = true
          })
      )
    }

    if (hasModule('conges')) {
      tasks.push(
        listLeaves()
          .then((items) => {
            stats.leavePending = countLeaveByStatus(items, 'en_attente')
            if (canValidateConges.value) {
              stats.pendingValidations = stats.leavePending
            }
            charts.leaveStatus = leaveStatusSeries(items as LeaveRequest[], leaveStatusLabel)
          })
          .catch(() => {
            errors.conges = true
          })
      )
    }

    if (hasModule('tma')) {
      tasks.push(
        listTma()
          .then((items) => {
            stats.tmaTotal = items.length
            stats.tmaOpen = countTmaOpen(items)
            charts.tmaStatus = tmaStatusSeries(items as TmaDemand[], tmaStatusLabel)
          })
          .catch(() => {
            errors.tma = true
          })
      )
    }

    if (hasModule('budget')) {
      tasks.push(
        listBudgets()
          .then((items) => {
            const m = budgetMetrics(items)
            stats.budgetOverrun = m.overrun
            stats.budgetConsumptionPct = consumptionPct(m.consumedDays, m.plannedDays, false)
            charts.budgetConsumption = budgetConsumptionSeries(items as BudgetItem[], budgetLabel)
          })
          .catch(() => {
            errors.budget = true
          })
      )
    }

    await Promise.all(tasks)
    return { stats, charts, errors }
  }

  const craCurrentLabel = (status: string | null) => {
    if (!status) return '—'
    return craStatusLabel(status)
  }

  const showModule = (code: ModuleCode) => hasModule(code)

  return { load, emptyStats, emptyCharts, craCurrentLabel, showModule, currentMonthKey, canValidateConges }
}
