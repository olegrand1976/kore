export function mapCraApiError(err: unknown, t: (key: string) => string, fallback?: string): string {
  if (err && typeof err === 'object') {
    const e = err as {
      statusCode?: number
      data?: { message?: string; error?: { message?: string; code?: string } }
      statusMessage?: string
      message?: string
    }
    const code = e.data?.error?.code ?? ''
    const message = (
      e.data?.error?.message ??
      e.data?.message ??
      e.statusMessage ??
      e.message ??
      ''
    ).toLowerCase()
    const status = e.statusCode ?? 0

    switch (code) {
      case 'COMMERCIAL_INFO_REQUIRED':
        return t('cra.errors.commercial_required')
      case 'DAY_CAPACITY_EXCEEDED':
        return t('cra.errors.day_capacity')
      case 'CRA_CONFLICT_ABSENCE':
        return t('cra.errors.conflict_absence')
      case 'CRA_ALREADY_VALIDATED':
        return t('cra.errors.already_validated')
      case 'WEEK_INCOMPLETE':
        return t('cra.errors.week_incomplete')
      default:
        break
    }

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
    mapCraError: (err: unknown, fallback?: string) => mapCraApiError(err, t, fallback),
    mapInvoiceDraftMessage: (
      draft: { status?: string; reason?: string } | null | undefined,
      skippedKey = 'cra.invoice_skipped'
    ) => mapInvoiceDraftMessage(draft, t, skippedKey)
  }
}

const invoiceReasonKeys: Record<string, string> = {
  client_unresolved: 'cra.invoice_reason.client_unresolved',
  no_billable_hours: 'cra.invoice_reason.no_billable_hours',
  billable_hours_error: 'cra.invoice_reason.billable_hours_error',
  publish_failed: 'cra.invoice_reason.publish_failed',
  already_exists_or_empty: 'cra.invoice_reason.already_exists_or_empty',
  invoicing_not_configured: 'cra.invoice_reason.invoicing_not_configured'
}

export function mapInvoiceDraftReason(
  reason: string | undefined,
  t: (key: string) => string
): string {
  const reasonKey = invoiceReasonKeys[reason ?? '']
  return reasonKey ? t(reasonKey) : (reason ?? t('cra.invoice_reason.unknown'))
}

export function mapInvoiceDraftMessage(
  draft: { status?: string; reason?: string } | null | undefined,
  t: (key: string, params?: Record<string, unknown>) => string,
  skippedKey = 'cra.invoice_skipped'
): string {
  if (!draft?.status || draft.status === 'created') {
    return draft?.status === 'created' ? t('cra.invoice_created') : t('cra.validated_ok')
  }
  if (draft.status === 'unavailable') {
    return t('cra.invoice_unavailable')
  }
  return t(skippedKey, { reason: mapInvoiceDraftReason(draft.reason, t) })
}
