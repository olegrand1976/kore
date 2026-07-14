export default defineNuxtRouteMiddleware(async (to) => {
  const blockedPrefixes = ['/conges', '/tma']
  if (!blockedPrefixes.some((prefix) => to.path.startsWith(prefix))) {
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

  const res = await $fetch<{ data?: Array<{ id?: string; month?: string; Month?: string; status?: string; Status?: string }> }>(
    '/api/cra/timesheets/recent?limit=6'
  ).catch(() => null)
  const items = res?.data ?? []
  const now = new Date()
  const monthKey = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
  const current = items.find((ts) => (ts.month ?? ts.Month) === monthKey)
  const status = String(current?.status ?? current?.Status ?? '')
  const incomplete = !current || (status !== 'ValidéSemaine' && status !== 'Définitif')
  if (!incomplete) {
    return
  }

  const id = String(current?.id ?? '')
  return navigateTo(id ? `/cra/${id}` : '/cra')
})
