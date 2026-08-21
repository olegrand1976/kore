import { beforeAll, describe, expect, it, vi } from 'vitest'
import { computed, reactive, ref, toValue } from 'vue'
import { currentMonthKey, useCraStatus } from '../composables/useCraStatus'
import { useCraMonthStats } from '../composables/useCraMonthStats'
import { minEditionPrice, matchEdition, parsePricingEditions, parsePricingModules, suggestUpgradeEdition } from '../composables/usePricingCatalog'
import { ALL_MODULES, useEntitlements } from '../composables/useEntitlements'
import { fetchWithRefresh } from '../composables/useApiFetch'
import { mapCraApiError } from '../composables/useCraError'
import { useReporting } from '../composables/useReporting'
import { buildKey, useWeekRows } from '../composables/useWeekRows'
import { decodeWorkRef, encodeWorkRef } from '../composables/useCraWorkRefs'
import {
  isManualPrestationEntry,
  missionPrestationPatch,
  unwrapMissionPayload
} from '../utils/craPrestation'
import { timesheetAdminAction, timesheetAdminConfirmKey } from '../utils/craTimesheetAdmin'
import {
  applyTextSearch,
  compareValues,
  groupByKey,
  useListControls
} from '../composables/useListControls'
import { resolveLogoUrl } from '../composables/useTenantBranding'

beforeAll(() => {
  vi.stubGlobal('useI18n', () => ({ t: (key: string) => key }))
  vi.stubGlobal('useState', (_key: string, init: () => unknown) => ref(init()))
  vi.stubGlobal('computed', computed)
})

describe('parseLineDurationMinutes', () => {
  it('parses API shapes without producing NaN', async () => {
    const { parseLineDurationMinutes } = await import('../utils/craDuration')
    expect(parseLineDurationMinutes({ Duration: { Minutes: 240 } })).toBe(240)
    expect(parseLineDurationMinutes({ duration: 120 })).toBe(120)
    expect(parseLineDurationMinutes({ duration: {} })).toBe(0)
    expect(parseLineDurationMinutes({})).toBe(0)
  })
})

describe('useCraMonthStats', () => {
  it('stays finite when a line duration is malformed', () => {
    const weeks = ref([{
      weekNumber: 1,
      lines: [{ sourceType: 'holiday', sourceId: '2026-07-14', day: '2026-07-14', duration: Number.NaN }],
      submittedAt: null
    }])
    const month = ref('2026-07')
    const weekStartDay = ref(1)
    const dayCapacity = ref(480)

    const stats = useCraMonthStats(weeks, month, weekStartDay, dayCapacity)
    expect(Number.isFinite(stats.totalMinutes.value)).toBe(true)
    expect(Number.isFinite(stats.capacityMinutes.value)).toBe(true)
  })

  it('recalculates capacity when dayCapacityMinutes ref changes', () => {
    const weeks = ref([{ weekNumber: 1, lines: [], submittedAt: null }])
    const month = ref('2026-07')
    const weekStartDay = ref(1)
    const dayCapacity = ref(480)

    const stats = useCraMonthStats(weeks, month, weekStartDay, dayCapacity)
    const before = stats.capacityMinutes.value

    dayCapacity.value = 420
    expect(stats.capacityMinutes.value).toBeLessThan(before)
    expect(stats.capacityMinutes.value).toBeGreaterThan(0)
  })
})

describe('useCraStatus', () => {
  it('returns YYYY-MM format', () => {
    expect(currentMonthKey()).toMatch(/^\d{4}-\d{2}$/)
  })

  it('maps status to i18n labels', () => {
    const { statusLabel } = useCraStatus()
    expect(statusLabel('Brouillon')).toBe('cra.status_draft')
    expect(statusLabel('ValidéSemaine')).toBe('cra.status_submitted')
    expect(statusLabel('Définitif')).toBe('cra.status_validated')
    expect(statusLabel('Unknown')).toBe('Unknown')
  })

  it('exposes currentMonthKey from the composable', () => {
    const { currentMonthKey: fromComposable } = useCraStatus()
    expect(typeof fromComposable).toBe('function')
    expect(fromComposable()).toMatch(/^\d{4}-\d{2}$/)
    expect(fromComposable()).toBe(currentMonthKey())
  })

  it('maps status to badge variants', () => {
    const { statusVariant } = useCraStatus()
    expect(statusVariant('Définitif')).toBe('success')
    expect(statusVariant('ValidéSemaine')).toBe('warning')
    expect(statusVariant('Brouillon')).toBe('default')
  })
})

describe('timesheetAdminAction', () => {
  it('asks to unvalidate a final timesheet before delete', () => {
    expect(timesheetAdminAction('Définitif')).toBe('unvalidate')
    expect(timesheetAdminAction('ValidéSemaine')).toBe('delete')
    expect(timesheetAdminAction('Brouillon')).toBe('delete')
  })

  it('picks confirm copy with or without a user name', () => {
    expect(timesheetAdminConfirmKey('unvalidate', true)).toBe('cra.unvalidate_confirm')
    expect(timesheetAdminConfirmKey('unvalidate', false)).toBe('cra.unvalidate_confirm_simple')
    expect(timesheetAdminConfirmKey('delete', true)).toBe('cra.delete_confirm')
    expect(timesheetAdminConfirmKey('delete', false)).toBe('cra.delete_confirm_simple')
  })
})

describe('useEntitlements.hasModule', () => {
  it('grants everything until loaded', () => {
    const { hasModule } = useEntitlements()
    expect(hasModule('cra')).toBe(true)
  })

  it('filters modules once loaded', () => {
    const ent = useEntitlements()
    ent.loaded.value = true
    ent.modules.value = ['cra', 'conges']
    expect(ent.hasModule('cra')).toBe(true)
    expect(ent.hasModule('tma')).toBe(false)
  })

  it('denies modules when loaded with empty list', () => {
    const ent = useEntitlements()
    ent.loaded.value = true
    ent.modules.value = []
    expect(ent.hasModule('billing')).toBe(false)
  })

  it('grants all modules on dev tenant fallback (404)', () => {
    const ent = useEntitlements()
    ent.loaded.value = true
    ent.modules.value = [...ALL_MODULES]
    expect(ent.hasModule('cra')).toBe(true)
    expect(ent.hasModule('tma')).toBe(true)
  })
})

describe('fetchWithRefresh', () => {
  it('retries once after a 401 then succeeds', async () => {
    const fetchFn = vi
      .fn()
      .mockRejectedValueOnce({ statusCode: 401 })
      .mockResolvedValueOnce({ ok: true })
    const refreshFn = vi.fn().mockResolvedValue(true)
    const onAuthFailure = vi.fn()

    const res = await fetchWithRefresh<{ ok: boolean }>(fetchFn, refreshFn, onAuthFailure, '/api/cra')

    expect(res).toEqual({ ok: true })
    expect(fetchFn).toHaveBeenCalledTimes(2)
    expect(refreshFn).toHaveBeenCalledTimes(1)
    expect(onAuthFailure).not.toHaveBeenCalled()
  })

  it('redirects to login when refresh fails', async () => {
    const fetchFn = vi.fn().mockRejectedValue({ statusCode: 401 })
    const refreshFn = vi.fn().mockResolvedValue(false)
    const onAuthFailure = vi.fn()

    await expect(
      fetchWithRefresh(fetchFn, refreshFn, onAuthFailure, '/api/cra')
    ).rejects.toMatchObject({ statusCode: 401 })
    expect(refreshFn).toHaveBeenCalledTimes(1)
    expect(onAuthFailure).toHaveBeenCalledTimes(1)
    expect(fetchFn).toHaveBeenCalledTimes(1)
  })

  it('does not retry on non-401 errors', async () => {
    const fetchFn = vi.fn().mockRejectedValue({ statusCode: 500 })
    const refreshFn = vi.fn()
    const onAuthFailure = vi.fn()

    await expect(
      fetchWithRefresh(fetchFn, refreshFn, onAuthFailure, '/api/cra')
    ).rejects.toMatchObject({ statusCode: 500 })
    expect(refreshFn).not.toHaveBeenCalled()
    expect(fetchFn).toHaveBeenCalledTimes(1)
  })
})

describe('parsePricingModules', () => {
  it('reads modules from data.catalog.modules', () => {
    const modules = parsePricingModules({
      data: {
        catalog: {
          modules: [{ code: 'cra', name: 'CRA', description: 'Timesheets', unitAmount: 1200 }]
        }
      }
    })
    expect(modules).toHaveLength(1)
    expect(modules[0]?.code).toBe('cra')
    expect(modules[0]?.unitAmount).toBe(1200)
  })
})

describe('parsePricingEditions', () => {
  it('reads editions from data.catalog.editions', () => {
    const editions = parsePricingEditions({
      data: {
        catalog: {
          editions: [
            {
              code: 'starter',
              name: 'Starter',
              description: 'Entry',
              unitAmount: 1200,
              modules: ['cra', 'conges']
            },
            {
              code: 'pro',
              name: 'Pro',
              description: 'ESN',
              unitAmount: 2500,
              modules: ['cra', 'tma'],
              highlight: true
            }
          ]
        }
      }
    })
    expect(editions).toHaveLength(2)
    expect(editions[0]?.code).toBe('starter')
    expect(editions[1]?.highlight).toBe(true)
    expect(minEditionPrice({
      data: { catalog: { editions: [{ code: 'starter', unitAmount: 1200, modules: [] }] } }
    })).toBe(1200)
  })
})

describe('matchEdition', () => {
  const editions = parsePricingEditions({
    data: {
      catalog: {
        editions: [
          { code: 'starter', unitAmount: 1200, modules: ['org', 'cra', 'conges', 'budget'] },
          { code: 'pro', unitAmount: 2500, modules: ['org', 'cra', 'conges', 'budget', 'tma', 'workflow'] },
          { code: 'enterprise', unitAmount: 4900, modules: ['org', 'cra', 'conges', 'budget', 'tma', 'workflow', 'notifications', 'billing'] }
        ]
      }
    }
  })

  it('matches starter modules', () => {
    expect(matchEdition(['org', 'cra', 'conges', 'budget'], editions)?.code).toBe('starter')
  })

  it('matches pro when tma is active', () => {
    expect(matchEdition(['org', 'cra', 'conges', 'budget', 'tma', 'workflow'], editions)?.code).toBe('pro')
  })

  it('suggests pro after starter', () => {
    const current = matchEdition(['org', 'cra', 'conges', 'budget'], editions)
    expect(suggestUpgradeEdition(current, editions)?.code).toBe('pro')
  })
})

describe('auth session shape', () => {
  it('validates minimal session payload', () => {
    const session = { ok: true, profile: 'Administrateur', userId: 'u1', tenantId: 't1' }
    expect(session.ok).toBe(true)
    expect(session.profile).toBe('Administrateur')
  })
})

describe('resolveLogoUrl', () => {
  const tenant = 'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee'
  const apiLogo = `/api/v1/branding/logo/${tenant}`

  it('maps API logo path to BFF URL', () => {
    expect(resolveLogoUrl(apiLogo, tenant)).toBe(`/api/org/branding/logo/${tenant}`)
  })

  it('ignores empty-object tenantId from legacy Go JSON', () => {
    expect(resolveLogoUrl(apiLogo, {})).toBe(`/api/org/branding/logo/${tenant}`)
    expect(resolveLogoUrl(apiLogo, { value: tenant })).toBe(`/api/org/branding/logo/${tenant}`)
  })

  it('uses string tenantId when logo has no path id', () => {
    expect(resolveLogoUrl('stored', tenant)).toBe(`/api/org/branding/logo/${tenant}`)
  })

  it('keeps blob and BFF urls as-is', () => {
    expect(resolveLogoUrl('blob:http://localhost/x')).toBe('blob:http://localhost/x')
    expect(resolveLogoUrl(`/api/org/branding/logo/${tenant}`)).toBe(`/api/org/branding/logo/${tenant}`)
  })

  it('returns null when logo is missing', () => {
    expect(resolveLogoUrl(null, tenant)).toBeNull()
    expect(resolveLogoUrl(undefined)).toBeNull()
  })
})

describe('TenantLogo load fallback', () => {
  it('exposes error handler pattern via component source', async () => {
    const { readFileSync } = await import('node:fs')
    const { join } = await import('node:path')
    const src = readFileSync(join(__dirname, '../components/brand/TenantLogo.vue'), 'utf8')
    expect(src).toContain('@error="onLogoError"')
    expect(src).toContain('loadFailed')
    expect(src).toContain("emit('error')")
    expect(src).toContain('tenant-logo--framed')
    expect(src).not.toContain('tenant-logo--${size}')
    expect(src).not.toContain('logoUrl!')
  })
})

describe('rollingWindow60', () => {
  it('returns a 60-day inclusive window', () => {
    const { rollingWindow60 } = useReporting()
    const period = rollingWindow60(new Date('2026-07-15T12:00:00Z'))
    expect(period.window).toBe('60')
    expect(period.start).toBe('2026-07-15')
    expect(period.end).toBe('2026-09-12')
  })
})

describe('mapCraApiError', () => {
  it('maps CRA business error codes', () => {
    const err = {
      statusCode: 422,
      data: { error: { code: 'COMMERCIAL_INFO_REQUIRED', message: 'commercial info required' } }
    }
    expect(mapCraApiError(err, (key) => key)).toBe('cra.errors.prestation_required')
  })

  it('maps the backend sentinel message without a code', () => {
    const err = {
      statusCode: 422,
      data: { error: { message: 'commercial info required' } }
    }
    expect(mapCraApiError(err, (key) => key)).toBe('cra.errors.prestation_required')
  })

  it('does not treat unrelated commercial wording as a CRA prestation error', () => {
    const err = {
      statusCode: 422,
      data: { error: { code: 'VALIDATION', message: 'invalid commercialId' } }
    }
    expect(mapCraApiError(err, (key) => key)).toBe('cra.errors.validation')
  })

  it('maps already invoiced conflict', () => {
    const err = {
      statusCode: 409,
      data: { error: { code: 'CRA_ALREADY_INVOICED', message: 'cra already invoiced' } }
    }
    expect(mapCraApiError(err, (key) => key)).toBe('cra.errors.already_invoiced')
  })
})

describe('mapInvoiceDraftReason', () => {
  it('maps known invoice skip reasons', async () => {
    const { mapInvoiceDraftReason, mapInvoiceDraftMessage } = await import('../composables/useCraError')
    const t = (key: string) => key
    expect(mapInvoiceDraftReason('client_unresolved', t)).toBe('cra.invoice_reason.client_unresolved')
    expect(mapInvoiceDraftReason('zero_unit_price', t)).toBe('cra.invoice_reason.zero_unit_price')
    expect(mapInvoiceDraftReason('invoicing_disabled', t)).toBe('cra.invoice_reason.invoicing_disabled')
    expect(mapInvoiceDraftReason('billing_mode_unresolved', t)).toBe('cra.invoice_reason.billing_mode_unresolved')
    expect(mapInvoiceDraftMessage({ status: 'unavailable' }, t)).toBe('cra.invoice_unavailable')
  })
})

describe('craDayState', () => {
  it('detects full absence day and unlocks holiday prefill', async () => {
    const { isFullAbsenceDay, partialAbsenceHoursLabel, unlockHolidayPrefillRows } = await import('../utils/craDayState')
    const toMinutes = (hours: string) => Number(hours) * 60

    expect(isFullAbsenceDay([{ sourceType: 'holiday', hours: '', origin: 'prefill' }], toMinutes)).toBe(true)
    expect(isFullAbsenceDay([{ sourceType: 'holiday', hours: '4', origin: 'manual' }], toMinutes)).toBe(false)
    expect(isFullAbsenceDay([
      { sourceType: 'holiday', hours: '', origin: 'prefill' },
      { sourceType: 'manual', hours: '', origin: 'manual' },
    ], toMinutes)).toBe(false)
    expect(isFullAbsenceDay([
      { sourceType: 'holiday', hours: '', origin: 'prefill' },
      { sourceType: 'mission', hours: '4', origin: 'manual' },
    ], toMinutes)).toBe(false)

    expect(partialAbsenceHoursLabel(480)).toBe('4')
    expect(partialAbsenceHoursLabel(420)).toBe('3.5')

    const unlocked = unlockHolidayPrefillRows([
      { sourceType: 'holiday', hours: '', origin: 'prefill' },
      { sourceType: 'leave', hours: '', origin: 'prefill' },
    ])
    expect(unlocked[0]?.origin).toBe('manual')
    expect(unlocked[1]?.origin).toBe('prefill')
  })
})

describe('useWeekRows toSaveLines', () => {
  it('skips empty rows and persists edited hours', () => {
    const week = ref({
      weekNumber: 1,
      lines: [{
        sourceType: 'manual',
        sourceId: 'default',
        day: '2026-07-07',
        duration: 240,
        comment: '',
        origin: 'manual',
        billable: true
      }],
      submittedAt: null
    })
    const { toSaveLines } = useWeekRows(week, ref(1), ref('2026-07'), ref(1))
    const lines = toSaveLines([
      {
        key: buildKey('manual', 'default', '2026-07-07'),
        sourceType: 'manual',
        sourceId: 'default',
        day: '2026-07-07',
        hours: '7.5',
        comment: 'done',
        origin: 'manual',
        billable: true
      },
      {
        key: buildKey('manual', 'extra', '2026-07-08'),
        sourceType: 'manual',
        sourceId: 'extra',
        day: '2026-07-08',
        hours: '',
        comment: '',
        origin: 'manual',
        billable: true
      }
    ])
    expect(lines).toHaveLength(1)
    expect(lines[0].duration).toBe(450)
    expect(lines[0].comment).toBe('done')
  })

  it('persists comment-only rows with zero duration', () => {
    const week = ref({
      weekNumber: 1,
      lines: [],
      submittedAt: null
    })
    const { toSaveLines } = useWeekRows(week, ref(1), ref('2026-07'), ref(1))
    const lines = toSaveLines([
      {
        key: buildKey('manual', 'default', '2026-07-07'),
        sourceType: 'manual',
        sourceId: 'default',
        day: '2026-07-07',
        hours: '',
        comment: '  note sans heures  ',
        origin: 'manual',
        billable: true
      },
      {
        key: buildKey('manual', 'extra', '2026-07-08'),
        sourceType: 'manual',
        sourceId: 'extra',
        day: '2026-07-08',
        hours: '',
        comment: '   ',
        origin: 'manual',
        billable: true
      }
    ])
    expect(lines).toHaveLength(1)
    expect(lines[0].duration).toBe(0)
    expect(lines[0].comment).toBe('  note sans heures  ')
  })

  it('persists duplicate activity types on the same day', () => {
    const week = ref({
      weekNumber: 1,
      lines: [],
      submittedAt: null
    })
    const { toSaveLines } = useWeekRows(week, ref(1), ref('2026-07'), ref(1))
    const lines = toSaveLines([
      {
        key: 'row-1',
        sourceType: 'manual',
        sourceId: 'default',
        day: '2026-07-07',
        hours: '5',
        comment: '',
        origin: 'manual',
        billable: true
      },
      {
        key: 'row-2',
        sourceType: 'manual',
        sourceId: 'default',
        day: '2026-07-07',
        hours: '3',
        comment: 'interne',
        origin: 'manual',
        billable: false
      }
    ])
    expect(lines).toHaveLength(2)
    expect(lines[0].duration).toBe(300)
    expect(lines[1].duration).toBe(180)
    expect(lines[1].comment).toBe('interne')
    expect(lines[1].billable).toBe(false)
  })

  it('persists work reference on a line', () => {
    const week = ref({ weekNumber: 1, lines: [], submittedAt: null })
    const { toSaveLines } = useWeekRows(week, ref(1), ref('2026-07'), ref(1))
    const lines = toSaveLines([{
      key: 'row-1',
      sourceType: 'manual',
      sourceId: 'default',
      day: '2026-07-07',
      hours: '5',
      comment: '',
      origin: 'manual',
      billable: true,
      workRefType: 'tma',
      workRefId: 'abc-123'
    }])
    expect(lines[0].workRefType).toBe('tma')
    expect(lines[0].workRefId).toBe('abc-123')
  })
})

describe('useCraWorkRefs encoding', () => {
  it('round-trips work ref values', () => {
    const encoded = encodeWorkRef('ticket', 'uuid-1')
    expect(decodeWorkRef(encoded)).toEqual({ type: 'ticket', id: 'uuid-1' })
    expect(decodeWorkRef('')).toEqual({ type: '', id: '' })
  })
})

describe('craPrestation', () => {
  it('treats blank mission id as manual entry', () => {
    expect(isManualPrestationEntry('')).toBe(true)
    expect(isManualPrestationEntry('   ')).toBe(true)
    expect(isManualPrestationEntry(undefined)).toBe(true)
    expect(isManualPrestationEntry('mission-1')).toBe(false)
  })

  it('unwraps BFF { data } and flat payloads', () => {
    expect(unwrapMissionPayload({ data: { clientName: 'ACME' } })).toEqual({ clientName: 'ACME' })
    expect(unwrapMissionPayload({ clientName: 'ACME' })).toEqual({ clientName: 'ACME' })
    expect(unwrapMissionPayload(null)).toEqual({})
  })

  it('copies technologies and contact from a mission, including empty contact', () => {
    const filled = missionPrestationPatch({
      clientName: 'ACME',
      clientId: 'c1',
      technologies: ['Go', ' Vue '],
      clientContact: 'Jane'
    })
    expect(filled).toEqual({
      client: 'ACME',
      clientId: 'c1',
      technologies: ['Go', 'Vue'],
      responsableClient: 'Jane'
    })

    const cleared = missionPrestationPatch({
      ClientName: 'Initech',
      Technologies: [],
      ClientContact: '  '
    })
    expect(cleared.client).toBe('Initech')
    expect(cleared.technologies).toEqual([])
    expect(cleared.responsableClient).toBe('')
  })
})

describe('useListControls helpers', () => {
  it('applyTextSearch is case insensitive', () => {
    expect(applyTextSearch('foo', 'Hello FOO World')).toBe(true)
    expect(applyTextSearch('bar', 'Hello FOO World')).toBe(false)
    expect(applyTextSearch('', 'anything')).toBe(true)
  })

  it('compareValues handles string, number and date', () => {
    expect(compareValues('b', 'a', 'string')).toBeGreaterThan(0)
    expect(compareValues(2, 10, 'number')).toBeLessThan(0)
    expect(compareValues('2026-07-01', '2026-06-01', 'date')).toBeGreaterThan(0)
    expect(compareValues(null, 'a', 'string')).toBeGreaterThan(0)
  })

  it('groupByKey buckets items', () => {
    const grouped = groupByKey(
      [
        { id: '1', status: 'open' },
        { id: '2', status: 'done' },
        { id: '3', status: 'open' }
      ],
      (item) => item.status
    )
    expect(grouped.open).toHaveLength(2)
    expect(grouped.done).toHaveLength(1)
  })
})

describe('useListControls', () => {
  type Row = { id: string; status: string; title: string; month: string; createdAt: string }

  const sample: Row[] = [
    { id: '1', status: 'open', title: 'Alpha', month: '2026-06', createdAt: '2026-06-10' },
    { id: '2', status: 'done', title: 'Beta', month: '2026-07', createdAt: '2026-07-01' },
    { id: '3', status: 'open', title: 'Gamma', month: '2026-07', createdAt: '2026-07-15' }
  ]

  it('filters by status and sorts by date desc', () => {
    const items = ref<Row[]>(sample)
    const controls = useListControls(items, {
      defaultSort: { key: 'createdAt', dir: 'desc' },
      filters: {
        status: {
          type: 'select',
          label: 'Status',
          options: [
            { value: 'open', label: 'Open' },
            { value: 'done', label: 'Done' }
          ],
          match: (row, value) => row.status === value
        },
        month: {
          type: 'month',
          label: 'Month',
          match: (row, value) => row.month === value
        }
      },
      sortKeys: [
        { key: 'title', label: 'Title', type: 'string', accessor: (row) => row.title },
        { key: 'createdAt', label: 'Created', type: 'date', accessor: (row) => row.createdAt }
      ]
    })

    controls.setFilter('status', 'open')
    controls.setFilter('month', '2026-07')
    expect(controls.filteredItems.value).toHaveLength(1)
    expect(controls.sortedItems.value[0]?.id).toBe('3')

    controls.resetFilters()
    expect(controls.filteredItems.value).toHaveLength(3)
    expect(controls.hasActiveFilters.value).toBe(false)
  })

  it('supports search filter and sort direction', () => {
    const items = ref<Row[]>(sample)
    const controls = useListControls(items, {
      filters: {
        q: {
          type: 'search',
          label: 'Search',
          match: (row, query) => applyTextSearch(query, row.title)
        }
      },
      sortKeys: [{ key: 'title', label: 'Title', type: 'string', accessor: (row) => row.title }]
    })

    controls.setFilter('q', 'beta')
    expect(controls.filteredItems.value).toHaveLength(1)
    controls.setFilter('q', '')
    controls.setSort('title', 'desc')
    expect(controls.sortedItems.value[0]?.title).toBe('Gamma')
  })
})

describe('rbac formatRbacCell / rbacCan', () => {
  it('formats L/E/V cells from PROFILE_PERMISSIONS', async () => {
    const { formatRbacCell, rbacCan, isImplementedRbacProfile } = await import('../utils/rbac')
    expect(formatRbacCell('Collaborateur', 'cra')).toBe('L/E')
    expect(formatRbacCell("Chef d'équipe", 'conges')).toBe('L')
    expect(formatRbacCell('Administrateur', 'org')).toBe('L/E/V')
    expect(formatRbacCell('Administrateur', 'ett')).toBe('L/E/V')
    expect(formatRbacCell('Collaborateur', 'org')).toBe('—')
    expect(rbacCan('Collaborateur', 'budget', 'L')).toBe(true)
    expect(rbacCan('Collaborateur', 'budget', 'E')).toBe(false)
    expect(rbacCan(['Collaborateur', "Chef d'équipe"], 'cra', 'V')).toBe(true)
    expect(isImplementedRbacProfile('Administrateur')).toBe(true)
    expect(isImplementedRbacProfile('Commercial')).toBe(false)
  })
})
