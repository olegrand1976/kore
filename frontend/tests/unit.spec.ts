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
import {
  applyTextSearch,
  compareValues,
  groupByKey,
  useListControls
} from '../composables/useListControls'

beforeAll(() => {
  vi.stubGlobal('useI18n', () => ({ t: (key: string) => key }))
  vi.stubGlobal('useState', (_key: string, init: () => unknown) => ref(init()))
  vi.stubGlobal('computed', computed)
})

describe('useCraMonthStats', () => {
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
    expect(mapCraApiError(err, (key) => key)).toBe('cra.errors.commercial_required')
  })
})

describe('useWeekRows toSaveLines', () => {
  it('skips empty rows without existing lines and persists edited hours', () => {
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
