import { isCraMonthIncomplete, type CraTimesheet } from '~/composables/useKpiMetrics'
import { isCraGateBlockedPath } from '~/utils/craGatePaths'

export default defineNuxtRouteMiddleware(async (to) => {
  if (!isCraGateBlockedPath(to.path)) {
    return
  }

  const calendar = await $fetch<{ data?: { craGateMode?: string } }>('/api/org/users/me/calendar-settings').catch(() => null)
  if ((calendar?.data?.craGateMode ?? 'warn') !== 'block') {
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
