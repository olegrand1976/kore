export function mapCraApiError(err: unknown, t: (key: string) => string, fallback?: string): string {
  if (err && typeof err === 'object') {
    const e = err as {
      statusCode?: number
      data?: { message?: string; error?: { message?: string; code?: string } }
      statusMessage?: string
      message?: string
    }
    const message = (
      e.data?.error?.message ??
      e.data?.message ??
      e.statusMessage ??
      e.message ??
      ''
    ).toLowerCase()
    const status = e.statusCode ?? 0

    if (message.includes('commercial') || message.includes('commercial info')) {
      return t('cra.errors.commercial_required')
    }
    if (message.includes('capacity') || message.includes('capacité')) {
      return t('cra.errors.day_capacity')
    }
    if (message.includes('absence') || message.includes('conflict')) {
      return t('cra.errors.conflict_absence')
    }
    if (message.includes('already validated') || message.includes('définitif')) {
      return t('cra.errors.already_validated')
    }
    if (message.includes('incomplete') || message.includes('incomplet')) {
      return t('cra.errors.week_incomplete')
    }
    if (status === 409) {
      return t('cra.errors.conflict')
    }
    if (status === 422) {
      return t('cra.errors.validation')
    }
  }
  return fallback ?? t('cra.errors.generic')
}

export function useCraError() {
  const { t } = useI18n()
  return {
    mapCraError: (err: unknown, fallback?: string) => mapCraApiError(err, t, fallback)
  }
}
