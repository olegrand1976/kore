type SetupResponse = {
  data?: {
    otpauthUrl?: string
    secret?: string
    qrCodeDataUrl?: string
  }
}

export function useTwoFactorSetup() {
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
      const res = await $fetch<SetupResponse>(endpoint, {
        method: options?.method ?? 'POST',
        body: options?.body
      })
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
