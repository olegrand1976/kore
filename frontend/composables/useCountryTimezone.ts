export type CountryCode = 'FR' | 'BE' | 'MG' | 'MA' | 'TN' | 'CA'

const TIMEZONE_BY_PAYS: Record<CountryCode, string> = {
  FR: 'Europe/Paris',
  BE: 'Europe/Brussels',
  MG: 'Indian/Antananarivo',
  MA: 'Africa/Casablanca',
  TN: 'Africa/Tunis',
  CA: 'America/Toronto'
}

export function normalizeCountryCode(value: unknown): CountryCode {
  const code = typeof value === 'string' ? value.trim().toUpperCase() : ''
  switch (code) {
    case 'BE':
      return 'BE'
    case 'MG':
    case 'MD':
      return 'MG'
    case 'MA':
      return 'MA'
    case 'TN':
      return 'TN'
    case 'CA':
      return 'CA'
    case 'FR':
    case '':
      return 'FR'
    default:
      return 'FR'
  }
}

/** @deprecated Prefer normalizeCountryCode — kept for billing call sites. */
export const normalizeBillingCountry = normalizeCountryCode
export type BillingCountryCode = CountryCode

export function timezoneForPays(pays: unknown): string {
  return TIMEZONE_BY_PAYS[normalizeCountryCode(pays)]
}

/** i18n key `org.country_fr` … for a pays code. */
export function countryI18nKey(pays: unknown): `org.country_${Lowercase<CountryCode>}` {
  const code = normalizeCountryCode(pays).toLowerCase() as Lowercase<CountryCode>
  return `org.country_${code}`
}

export type ClockMode = 'time' | 'datetime'

function localeTag(locale: string): string {
  return locale === 'en' || locale.startsWith('en-') ? 'en-US' : 'fr-FR'
}

export function formatInTimezone(
  iso: string | Date,
  timeZone: string,
  locale: string,
  mode: ClockMode = 'time'
): string {
  const date = typeof iso === 'string' ? new Date(iso) : iso
  if (Number.isNaN(date.getTime())) return ''
  const options: Intl.DateTimeFormatOptions =
    mode === 'time'
      ? { hour: '2-digit', minute: '2-digit', hour12: false, timeZone }
      : {
          day: '2-digit',
          month: '2-digit',
          year: 'numeric',
          hour: '2-digit',
          minute: '2-digit',
          hour12: false,
          timeZone
        }
  return new Intl.DateTimeFormat(localeTag(locale), options).format(date)
}

/** Single clock in the organisation timezone (support, ETT, maintenance). */
export function formatOrgClock(
  iso: string | Date | null | undefined,
  orgPays: unknown,
  locale: string,
  mode: ClockMode = 'time'
): string {
  if (iso == null || iso === '') return '—'
  const date = typeof iso === 'string' ? new Date(iso) : iso
  if (Number.isNaN(date.getTime())) return '—'
  return formatInTimezone(date, timezoneForPays(orgPays), locale, mode) || '—'
}

/**
 * Long “now” label in org timezone (weekday + month name).
 * Used by ETT live clock — not a dual-TZ surface.
 */
export function formatOrgClockLong(date: Date, orgPays: unknown, locale: string): string {
  if (Number.isNaN(date.getTime())) return '—'
  return new Intl.DateTimeFormat(localeTag(locale), {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
    timeZone: timezoneForPays(orgPays)
  }).format(date)
}

/**
 * Org + client clocks when wall times differ.
 * Example: `14:00 (MG) · 11:00 (FR)`
 */
export function formatDualClock(
  iso: string | Date | null | undefined,
  orgPays: unknown,
  clientPays: CountryCode,
  locale: string,
  mode: ClockMode = 'time'
): string {
  if (iso == null || iso === '') return '—'
  const date = typeof iso === 'string' ? new Date(iso) : iso
  if (Number.isNaN(date.getTime())) return '—'

  const orgCode = normalizeCountryCode(orgPays)
  const clientCode = normalizeCountryCode(clientPays)
  const orgLabel = formatInTimezone(date, timezoneForPays(orgCode), locale, mode)
  if (!orgLabel) return '—'

  if (orgCode === clientCode) {
    return orgLabel
  }

  const clientLabel = formatInTimezone(date, timezoneForPays(clientCode), locale, mode)
  if (!clientLabel || clientLabel === orgLabel) {
    return orgLabel
  }
  return `${orgLabel} (${orgCode}) · ${clientLabel} (${clientCode})`
}
