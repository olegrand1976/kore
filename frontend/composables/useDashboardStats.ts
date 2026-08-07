import type { ChartBarItem, CraMonthItem } from '~/composables/useKpiMetrics'
import type { ModuleCode } from '~/composables/useEntitlements'
import { currentMonthKey } from '~/composables/useCraStatus'

export type DashboardStats = {
  craCurrentStatus: string | null
  craRequired: boolean
  craAlert: boolean
  craPrefillRatio: number | null
  craPrefillLow: boolean
  leavePending: number
  tmaOpen: number
  tmaTotal: number
  budgetOverrun: number
  budgetConsumptionPct: number
  pendingValidations: number
  billingAmountCents: number
  billingInvoiceCount: number
  billableHoursMonth: number
}

export type DashboardCharts = {
  tmaStatus: ChartBarItem[]
  budgetConsumption: ChartBarItem[]
  craMonths: CraMonthItem[]
  leaveStatus: ChartBarItem[]
}

export type DashboardStatErrors = {
  cra?: boolean
  conges?: boolean
  tma?: boolean
  budget?: boolean
  billing?: boolean
}

export type DashboardLoadResult = {
  stats: DashboardStats
  charts: DashboardCharts
  errors: DashboardStatErrors
}

type HomeStatusCount = { key?: string; Key?: string; value?: number; Value?: number }
type HomeBudgetBar = { key?: string; Key?: string; label?: string; Label?: string; value?: number; Value?: number }
type HomeCraMonth = { key?: string; Key?: string; status?: string | null; Status?: string | null }

type HomeDashboardPayload = {
  data?: {
    cra?: {
      required?: boolean
      alert?: boolean
      currentStatus?: string | null
      prefillRatio?: number | null
      prefillLow?: boolean
      months?: HomeCraMonth[]
    }
    leave?: {
      pending?: number
      pendingValidations?: number
      statusCounts?: HomeStatusCount[]
    }
    tma?: {
      open?: number
      total?: number
      statusCounts?: HomeStatusCount[]
    }
    budget?: {
      overrun?: number
      consumptionPct?: number
      bars?: HomeBudgetBar[]
    }
    billing?: {
      amountCents?: number
      invoiceCount?: number
      billableHours?: number
    }
    errors?: {
      cra?: boolean
      conges?: boolean
      tma?: boolean
      budget?: boolean
      billing?: boolean
    }
  }
}

const emptyStats = (): DashboardStats => ({
  craCurrentStatus: null,
  craRequired: false,
  craAlert: false,
  craPrefillRatio: null,
  craPrefillLow: false,
  leavePending: 0,
  tmaOpen: 0,
  tmaTotal: 0,
  budgetOverrun: 0,
  budgetConsumptionPct: 0,
  pendingValidations: 0,
  billingAmountCents: 0,
  billingInvoiceCount: 0,
  billableHoursMonth: 0
})

const emptyCharts = (): DashboardCharts => ({
  tmaStatus: [],
  budgetConsumption: [],
  craMonths: [],
  leaveStatus: []
})

const openTma = new Set(['ouverte', 'affectee', 'en_cours', 'rework'])

function statusToneTma(status: string): ChartBarItem['tone'] {
  if (status === 'resolue') return 'success'
  if (openTma.has(status)) return 'blue'
  if (status === 'en_attente_creation') return 'warn'
  return 'muted'
}

function statusToneLeave(status: string): ChartBarItem['tone'] {
  if (status === 'valide') return 'success'
  if (status === 'en_attente') return 'warn'
  if (status === 'refuse') return 'muted'
  return 'blue'
}

function budgetTone(pct: number): ChartBarItem['tone'] {
  if (pct > 100) return 'warn'
  if (pct >= 80) return 'gold'
  return 'success'
}

export function useDashboardStats() {
  const { apiFetch } = useApiFetch()
  const { hasModule } = useEntitlements()
  const { canValidateConges } = usePermissions()
  const { statusLabel: craStatusLabel } = useCraStatus()
  const { locale, t } = useI18n()

  const tmaStatusLabel = (status: string) => t(`dashboard.charts.status.tma.${status}`, status)
  const leaveStatusLabel = (status: string) => t(`dashboard.charts.status.leave.${status}`, status)

  const load = async (): Promise<DashboardLoadResult> => {
    const stats = emptyStats()
    const charts = emptyCharts()
    const errors: DashboardStatErrors = {}

    try {
      const res = await apiFetch<HomeDashboardPayload>('/api/dashboards/home')
      const home = res?.data
      if (!home) return { stats, charts, errors }

      if (home.errors?.cra) errors.cra = true
      if (home.errors?.conges) errors.conges = true
      if (home.errors?.tma) errors.tma = true
      if (home.errors?.budget) errors.budget = true
      if (home.errors?.billing) errors.billing = true

      if (home.cra) {
        stats.craRequired = !!home.cra.required
        stats.craAlert = !!home.cra.alert
        stats.craCurrentStatus = home.cra.currentStatus ?? null
        stats.craPrefillRatio = home.cra.prefillRatio ?? null
        stats.craPrefillLow = !!home.cra.prefillLow
        const loc = locale.value === 'en' ? 'en-US' : 'fr-FR'
        charts.craMonths = (home.cra.months ?? []).map((m) => {
          const key = m.key ?? m.Key ?? ''
          const [y, mo] = key.split('-').map(Number)
          const d = new Date(y || 2026, (mo || 1) - 1, 1)
          return {
            key,
            label: d.toLocaleDateString(loc, { month: 'short' }),
            status: m.status ?? m.Status ?? null
          } satisfies CraMonthItem
        })
      }

      if (home.leave) {
        stats.leavePending = home.leave.pending ?? 0
        stats.pendingValidations = home.leave.pendingValidations ?? 0
        charts.leaveStatus = (home.leave.statusCounts ?? []).map((c) => {
          const key = c.key ?? c.Key ?? 'unknown'
          return {
            key,
            label: leaveStatusLabel(key),
            value: c.value ?? c.Value ?? 0,
            tone: statusToneLeave(key)
          }
        })
      }

      if (home.tma) {
        stats.tmaOpen = home.tma.open ?? 0
        stats.tmaTotal = home.tma.total ?? 0
        charts.tmaStatus = (home.tma.statusCounts ?? []).map((c) => {
          const key = c.key ?? c.Key ?? 'unknown'
          return {
            key,
            label: tmaStatusLabel(key),
            value: c.value ?? c.Value ?? 0,
            tone: statusToneTma(key)
          }
        })
      }

      if (home.budget) {
        stats.budgetOverrun = home.budget.overrun ?? 0
        stats.budgetConsumptionPct = home.budget.consumptionPct ?? 0
        charts.budgetConsumption = (home.budget.bars ?? []).map((b) => {
          const value = b.value ?? b.Value ?? 0
          return {
            key: b.key ?? b.Key ?? '',
            label: b.label ?? b.Label ?? '',
            value,
            tone: budgetTone(value)
          }
        })
      }

      if (home.billing) {
        stats.billingAmountCents = home.billing.amountCents ?? 0
        stats.billingInvoiceCount = home.billing.invoiceCount ?? 0
        stats.billableHoursMonth = home.billing.billableHours ?? 0
      }
    } catch {
      errors.cra = true
      errors.conges = true
      errors.tma = true
      errors.budget = true
      errors.billing = true
    }

    return { stats, charts, errors }
  }

  const craCurrentLabel = (status: string | null) => {
    if (!status) return '—'
    return craStatusLabel(status)
  }

  const showModule = (code: ModuleCode) => hasModule(code)

  return { load, emptyStats, emptyCharts, craCurrentLabel, showModule, currentMonthKey, canValidateConges }
}
