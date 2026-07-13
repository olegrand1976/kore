export type PlatformSettings = {
  geminiModel: string
  llmProvider: string
  updatedAt?: string
}

type ApiEnvelope<T> = { data?: T }

export function usePlatformSettings() {
  const settings = ref<PlatformSettings | null>(null)
  const pending = ref(false)
  const saving = ref(false)
  const error = ref(false)
  const saveError = ref(false)

  const fetchSettings = async () => {
    pending.value = true
    error.value = false
    try {
      const res = await $fetch<ApiEnvelope<PlatformSettings>>('/api/platform/settings')
      settings.value = res.data ?? null
    } catch {
      settings.value = null
      error.value = true
    } finally {
      pending.value = false
    }
  }

  const saveSettings = async (geminiModel: string) => {
    saving.value = true
    saveError.value = false
    try {
      const res = await $fetch<ApiEnvelope<PlatformSettings>>('/api/platform/settings', {
        method: 'PUT',
        body: { geminiModel }
      })
      settings.value = res.data ?? null
      return true
    } catch {
      saveError.value = true
      return false
    } finally {
      saving.value = false
    }
  }

  return { settings, pending, saving, error, saveError, fetchSettings, saveSettings }
}
