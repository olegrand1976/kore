import { isCraMonthIncomplete, type CraTimesheet } from '~/composables/useKpiMetrics'

function currentMonthKey() {
  const now = new Date()
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
}

export function useCraGate() {
  const blocked = ref(false)
  const loading = ref(true)
  const currentTimesheetId = ref<string | null>(null)

  const refresh = async () => {
    loading.value = true
    try {
      const profile = await $fetch<{ data?: { craRequis?: boolean } }>('/api/org/users/me/profile')
      const required = profile?.data?.craRequis ?? false
      if (!required) {
        blocked.value = false
        currentTimesheetId.value = null
        return
      }
      const res = await $fetch<{ data?: CraTimesheet[] }>('/api/cra/timesheets/recent?limit=6')
      const items = (res?.data ?? []) as CraTimesheet[]
      blocked.value = isCraMonthIncomplete(items)
      const current = items.find((ts) => (ts.month ?? ts.Month) === currentMonthKey())
      currentTimesheetId.value = String(current?.id ?? current?.ID ?? '') || null
    } catch {
      blocked.value = false
      currentTimesheetId.value = null
    } finally {
      loading.value = false
    }
  }

  onMounted(refresh)

  return { blocked, loading, currentTimesheetId, refresh }
}
