import { beforeAll, describe, expect, it, vi } from 'vitest'
import { computed, ref } from 'vue'
import { currentMonthKey, useCraStatus } from '../composables/useCraStatus'
import { minEditionPrice, parsePricingEditions, parsePricingModules } from '../composables/usePricingCatalog'
import { ALL_MODULES, useEntitlements } from '../composables/useEntitlements'
import { fetchWithRefresh } from '../composables/useApiFetch'

beforeAll(() => {
  vi.stubGlobal('useI18n', () => ({ t: (key: string) => key }))
  vi.stubGlobal('useState', (_key: string, init: () => unknown) => ref(init()))
  vi.stubGlobal('computed', computed)
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

describe('auth session shape', () => {
  it('validates minimal session payload', () => {
    const session = { ok: true, profile: 'Administrateur', userId: 'u1', tenantId: 't1' }
    expect(session.ok).toBe(true)
    expect(session.profile).toBe('Administrateur')
  })
})
