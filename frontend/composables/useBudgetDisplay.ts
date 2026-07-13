import type { BadgeVariant } from '~/components/ui/AppBadge.vue'

export type BudgetStatus = 'ok' | 'warn' | 'overrun'

export function useBudgetDisplay() {
  const { t, locale } = useI18n()

  const normalizeBudgetType = (type: string) => type.toLowerCase()

  const budgetTypeLabel = (type: string) => {
    switch (normalizeBudgetType(type)) {
      case 'defaut':
        return t('budget.type_defaut')
      case 'specifique':
        return t('budget.type_specifique')
      default:
        return type
    }
  }

  const budgetStatus = (consumed: number, planned: number): BudgetStatus => {
    if (planned <= 0) return consumed > 0 ? 'overrun' : 'ok'
    const pct = (consumed / planned) * 100
    if (pct > 100) return 'overrun'
    if (pct >= 90) return 'warn'
    return 'ok'
  }

  const consumptionPercent = (consumed: number, planned: number) => {
    if (planned <= 0) return consumed > 0 ? 100 : 0
    return Math.round((consumed / planned) * 100)
  }

  const statusLabel = (status: BudgetStatus) => {
    switch (status) {
      case 'ok':
        return t('budget.status_ok')
      case 'warn':
        return t('budget.status_warn')
      case 'overrun':
        return t('budget.status_overrun')
      default: {
        const _exhaustive: never = status
        return _exhaustive
      }
    }
  }

  const statusBadgeVariant = (status: BudgetStatus): BadgeVariant => {
    switch (status) {
      case 'ok':
        return 'success'
      case 'warn':
        return 'warning'
      case 'overrun':
        return 'error'
      default: {
        const _exhaustive: never = status
        return _exhaustive
      }
    }
  }

  const formatBudgetAmount = (centimes: number, currency = 'EUR') => {
    const amount = centimes / 100
    return new Intl.NumberFormat(locale.value === 'en' ? 'en-US' : 'fr-FR', {
      style: 'currency',
      currency
    }).format(amount)
  }

  const eurosToCentimes = (euros: number) => Math.round(euros * 100)
  const centimesToEuros = (centimes: number) => centimes / 100

  const currentMonthPeriod = () => {
    const now = new Date()
    const start = new Date(now.getFullYear(), now.getMonth(), 1)
    const end = new Date(now.getFullYear(), now.getMonth() + 1, 0)
    return {
      start: start.toISOString().slice(0, 10),
      end: end.toISOString().slice(0, 10)
    }
  }

  const budgetPageTitle = (appLabel: string, budgetId: string) =>
    appLabel || `${t('budget.title')} — ${budgetId.slice(0, 8)}`

  const pickApplicationId = (budget: { applicationId?: string; ApplicationID?: string }) =>
    budget.applicationId ?? budget.ApplicationID ?? ''

  return {
    budgetTypeLabel,
    budgetStatus,
    consumptionPercent,
    statusLabel,
    statusBadgeVariant,
    formatBudgetAmount,
    eurosToCentimes,
    centimesToEuros,
    currentMonthPeriod,
    budgetPageTitle,
    pickApplicationId,
    normalizeBudgetType
  }
}
