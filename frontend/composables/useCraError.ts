import { extractFetchErrorCode } from '~/composables/useApiError'

function extractErrorMessage(err: unknown): string {
  if (!err || typeof err !== 'object') return ''
  const e = err as {
    data?: { message?: string; error?: string | { message?: string; code?: string } }
    statusMessage?: string
    message?: string
  }
  const nested = e.data?.error
  if (typeof nested === 'string' && nested.trim()) return nested
  if (nested && typeof nested === 'object' && typeof nested.message === 'string') {
    return nested.message
  }
  return e.data?.message ?? e.statusMessage ?? e.message ?? ''
}

export function isCommercialInfoRequiredError(err: unknown): boolean {
  const code = extractFetchErrorCode(err) ?? ''
  if (code === 'COMMERCIAL_INFO_REQUIRED') return true
  const message = extractErrorMessage(err).toLowerCase()
  return message.includes('commercial info required')
}

export function mapCraApiError(err: unknown, t: (key: string) => string, fallback?: string): string {
  if (err && typeof err === 'object') {
    const e = err as {
      statusCode?: number
      data?: { message?: string; error?: { message?: string; code?: string } }
      statusMessage?: string
      message?: string
    }
    const code = extractFetchErrorCode(err) ?? e.data?.error?.code ?? ''
    const message = extractErrorMessage(err).toLowerCase()
    const status = e.statusCode ?? 0

    switch (code) {
      case 'COMMERCIAL_INFO_REQUIRED':
        return t('cra.errors.prestation_required')
      case 'DAY_CAPACITY_EXCEEDED':
        return t('cra.errors.day_capacity')
      case 'CRA_CONFLICT_ABSENCE':
        return t('cra.errors.conflict_absence')
      case 'CRA_ALREADY_VALIDATED':
        return t('cra.errors.already_validated')
      case 'CRA_ALREADY_INVOICED':
        return t('cra.errors.already_invoiced')
      case 'WEEK_INCOMPLETE':
        return t('cra.errors.week_incomplete')
      case 'CRA_NOT_SUBMITTED':
        return t('cra.errors.not_submitted')
      case 'CRA_NO_LOGGED_TIME':
        return t('cra.errors.no_logged_time')
      default:
        break
    }

    if (message.includes('commercial info required')) {
      return t('cra.errors.prestation_required')
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
    if (message.includes('already invoiced')) {
      return t('cra.errors.already_invoiced')
    }
    if (message.includes('incomplete') || message.includes('incomplet')) {
      return t('cra.errors.week_incomplete')
    }
    if (message.includes('no submitted week') || message.includes('not submitted')) {
      return t('cra.errors.not_submitted')
    }
    if (message.includes('no logged time')) {
      return t('cra.errors.no_logged_time')
    }
    if (status === 409) {
      return t('cra.errors.conflict')
    }
    if (status === 422) {
      const raw = extractErrorMessage(err).trim()
      if (raw && raw.toLowerCase() !== 'invalid data' && raw.toLowerCase() !== 'données invalides.') {
        return `${t('cra.errors.validation')} (${raw})`
      }
      return t('cra.errors.validation')
    }
  }
  return fallback ?? t('cra.errors.generic')
}

export function useCraError() {
  const { t } = useI18n()
  return {
    mapCraError: (err: unknown, fallback?: string) => mapCraApiError(err, t, fallback),
    isCommercialInfoRequiredError,
    mapInvoiceDraftMessage: (
      draft: { status?: string; reason?: string } | null | undefined,
      skippedKey = 'cra.invoice_skipped'
    ) => mapInvoiceDraftMessage(draft, t, skippedKey),
    mapInvoiceDraftReason: (reason: string | undefined) => mapInvoiceDraftReason(reason, t)
  }
}

const invoiceReasonKeys: Record<string, string> = {
  client_unresolved: 'cra.invoice_reason.client_unresolved',
  no_billable_hours: 'cra.invoice_reason.no_billable_hours',
  billable_hours_error: 'cra.invoice_reason.billable_hours_error',
  publish_failed: 'cra.invoice_reason.publish_failed',
  already_exists_or_empty: 'cra.invoice_reason.already_exists_or_empty',
  already_exists: 'cra.invoice_reason.already_exists',
  already_exists_check_failed: 'cra.invoice_reason.publish_failed',
  invoicing_not_configured: 'cra.invoice_reason.invoicing_not_configured',
  invoicing_disabled: 'cra.invoice_reason.invoicing_disabled',
  zero_unit_price: 'cra.invoice_reason.zero_unit_price',
  billing_mode_disabled: 'cra.invoice_reason.billing_mode_disabled',
  billing_mode_forfait: 'cra.invoice_reason.billing_mode_forfait',
  billing_mode_unresolved: 'cra.invoice_reason.billing_mode_unresolved',
  not_definitive: 'cra.invoice_reason.not_definitive',
  timesheet_not_found: 'cra.invoice_reason.timesheet_not_found'
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
