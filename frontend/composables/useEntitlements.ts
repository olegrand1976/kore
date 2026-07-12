export type ModuleCode = 'org' | 'cra' | 'conges' | 'budget' | 'tma' | 'workflow' | 'notifications' | 'billing'

export const ALL_MODULES: ModuleCode[] = [
  'org', 'cra', 'conges', 'budget', 'tma', 'workflow', 'notifications', 'billing'
]

function httpStatus(err: unknown): number | undefined {
  const e = err as { statusCode?: number; status?: number; response?: { status?: number } }
  return e?.statusCode ?? e?.status ?? e?.response?.status
}

type SubscriptionModule = {
  moduleCode?: string
  ModuleCode?: string
  enabled?: boolean
  Enabled?: boolean
}

type SubscriptionResponse = {
  data?: {
    modules?: SubscriptionModule[]
    Modules?: SubscriptionModule[]
    status?: string
    Status?: string
  }
}

export function useEntitlements() {
  const modules = useState<ModuleCode[]>('entitlements-modules', () => [])
  const status = useState<string>('subscription-status', () => 'active')
  const loaded = useState<boolean>('entitlements-loaded', () => false)

  const fetchEntitlements = async () => {
    const { apiFetch } = useApiFetch()
    try {
      const res = await apiFetch<SubscriptionResponse>('/api/billing/subscription')
      const sub = res.data ?? {}
      const raw = sub.modules ?? sub.Modules ?? []
      modules.value = raw
        .filter((m) => m.enabled ?? m.Enabled ?? true)
        .map((m) => String(m.moduleCode ?? m.ModuleCode ?? '').toLowerCase() as ModuleCode)
        .filter(Boolean)
      status.value = String(sub.status ?? sub.Status ?? 'active')
      loaded.value = true
    } catch (err) {
      const code = httpStatus(err)
      if (code === 401 || code === 403) {
        // useApiFetch already attempted refresh + redirect to /login
        return
      }
      if (code === 404) {
        modules.value = [...ALL_MODULES]
        status.value = 'active'
      } else {
        modules.value = []
        status.value = 'active'
      }
      loaded.value = true
    }
  }

  const hasModule = (code: ModuleCode) => {
    if (!loaded.value) return true
    if (modules.value.length === 0) return false
    return modules.value.includes(code)
  }

  const isPastDue = computed(() => status.value === 'past_due')

  return { modules, status, loaded, fetchEntitlements, hasModule, isPastDue }
}
