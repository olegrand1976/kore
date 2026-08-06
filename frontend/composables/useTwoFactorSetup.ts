type SetupResponse = {
  data?: {
    otpauthUrl?: string
    secret?: string
    qrCodeDataUrl?: string
  }
}

export function useTwoFactorSetup() {
  const { apiFetch } = useApiFetch()
  const otpauthUrl = ref('')
  const manualSecret = ref('')
  const qrCodeDataUrl = ref('')
  const totpCode = ref('')
  const backupCodes = ref<string[]>([])
  const loading = ref(false)
  const error = ref('')

  async function loadSetup(endpoint: string, options?: { method?: string; body?: Record<string, unknown> }) {
    loading.value = true
    error.value = ''
    try {
      const fetchOptions = {
        method: options?.method ?? 'POST',
        body: options?.body
      }
      // Auth enrollment stays on raw $fetch (no session refresh loop on login).
      const res = endpoint.startsWith('/api/auth/')
        ? await $fetch<SetupResponse>(endpoint, fetchOptions)
        : await apiFetch<SetupResponse>(endpoint, fetchOptions)
      otpauthUrl.value = res?.data?.otpauthUrl ?? ''
      manualSecret.value = res?.data?.secret ?? ''
      qrCodeDataUrl.value = res?.data?.qrCodeDataUrl ?? ''
    } catch (e: unknown) {
      const err = e as { data?: { error?: { message?: string } } }
      error.value = err?.data?.error?.message ?? 'setup failed'
      throw e
    } finally {
      loading.value = false
    }
  }

  function reset() {
    otpauthUrl.value = ''
    manualSecret.value = ''
    qrCodeDataUrl.value = ''
    totpCode.value = ''
    backupCodes.value = []
    error.value = ''
  }

  return {
    otpauthUrl,
    manualSecret,
    qrCodeDataUrl,
    totpCode,
    backupCodes,
    loading,
    error,
    loadSetup,
    reset
  }
}
