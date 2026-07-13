export type TenantUsageSummary = {
  id: string
  name: string
  societeName: string
  createdAt: string
  subscriptionStatus: string
  seatLimit: number
  activeUsers: number
  seatUsagePct: number
  modulesEnabled: number
  craCount: number
  tmaCount: number
  tmaOpen: number
  budgetCount: number
  leaveCount: number
  aiRequests30d: number
  lastActivityAt: string | null
  activeLast30d: boolean
}

export type PlatformOverviewSummary = {
  totalTenants: number
  activeTenants30d: number
  totalActiveUsers: number
  totalSeatLimit: number
  tenantsByStatus: Record<string, number>
}

export type PlatformOverview = {
  summary: PlatformOverviewSummary
  tenants: TenantUsageSummary[]
}

type ApiEnvelope<T> = { data?: T }

export function usePlatformOverview() {
  const overview = ref<PlatformOverview | null>(null)
  const pending = ref(false)
  const error = ref(false)
  const forbidden = ref(false)

  const fetchOverview = async () => {
    pending.value = true
    error.value = false
    forbidden.value = false
    try {
      const res = await $fetch<ApiEnvelope<PlatformOverview>>('/api/platform/overview')
      overview.value = res.data ?? null
    } catch (e: unknown) {
      overview.value = null
      const status = (e as { statusCode?: number })?.statusCode
      if (status === 403) {
        forbidden.value = true
      } else {
        error.value = true
      }
    } finally {
      pending.value = false
    }
  }

  return { overview, pending, error, forbidden, fetchOverview }
}
