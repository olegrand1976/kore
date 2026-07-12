export function useBilling() {
  const { apiFetch } = useApiFetch()
  const loading = ref(false)
  const error = ref<string | null>(null)

  const startCheckout = async (payload: {
    modules: string[]
    seats: number
    successUrl: string
    cancelUrl: string
    email?: string
  }) => {
    loading.value = true
    error.value = null
    try {
      const res = await apiFetch<{ data?: { url?: string; URL?: string } }>('/api/billing/checkout-session', {
        method: 'POST',
        body: payload
      })
      const url = res.data?.url ?? res.data?.URL
      if (url && import.meta.client) {
        window.location.href = url
      }
      return res
    } catch {
      error.value = 'checkout_failed'
      throw new Error('checkout_failed')
    } finally {
      loading.value = false
    }
  }

  const openPortal = async (returnUrl: string) => {
    loading.value = true
    try {
      const res = await apiFetch<{ data?: { url?: string; URL?: string } }>('/api/billing/portal-session', {
        method: 'POST',
        body: { returnUrl }
      })
      const url = res.data?.url ?? res.data?.URL
      if (url && import.meta.client) {
        window.location.href = url
      }
      return res
    } finally {
      loading.value = false
    }
  }

  const cancelSubscription = () => apiFetch('/api/billing/cancel', { method: 'POST' })

  return { loading, error, startCheckout, openPortal, cancelSubscription }
}
