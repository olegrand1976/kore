import { isCraMonthIncomplete, type CraTimesheet } from '~/composables/useKpiMetrics'
import { readCachedCraGateMode, writeCraGateModeCache, type CraGateModeCache } from '~/utils/craGateCache'
import { isCraGateBlockedPath } from '~/utils/craGatePaths'

async function resolveCraGateMode(): Promise<string> {
  const cache = useState<CraGateModeCache | null>('cra-gate-mode', () => null)
  const cached = readCachedCraGateMode(cache.value)
  if (cached !== null) {
    return cached
  }

  const calendar = await $fetch<{ data?: { craGateMode?: string } }>('/api/org/users/me/calendar-settings').catch(
    () => null
  )
  const mode = calendar?.data?.craGateMode ?? 'warn'
  cache.value = writeCraGateModeCache(mode)
  return mode
}

export default defineNuxtRouteMiddleware(async (to) => {
  if (!isCraGateBlockedPath(to.path)) {
    return
  }

  const mode = await resolveCraGateMode()
  if (mode !== 'block') {
    return
  }

  const profile = await $fetch<{ data?: { craRequis?: boolean } }>('/api/org/users/me/profile').catch(() => null)
  if (!profile?.data?.craRequis) {
    return
  }

  const res = await $fetch<{ data?: CraTimesheet[] }>('/api/cra/timesheets/recent?limit=6').catch(() => null)
  const items = res?.data ?? []
  if (!isCraMonthIncomplete(items)) {
    return
  }

  const now = new Date()
  const monthKey = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
  const current = items.find((ts) => (ts.month ?? ts.Month) === monthKey)
  const id = String(current?.id ?? current?.ID ?? '')
  return navigateTo(id ? `/cra/${id}` : '/cra')
})
