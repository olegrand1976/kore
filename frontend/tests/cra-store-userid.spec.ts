import { describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useCraStore } from '../stores/cra'

describe('cra store normalizeTimesheet userId', () => {
  it('persists userId from API payload', () => {
    setActivePinia(createPinia())
    const store = useCraStore()
    store.setTimesheet({
      id: 'ts-1',
      userId: 'owner-uuid',
      month: '2026-09',
      status: 'Brouillon',
      weeks: []
    })
    expect(store.timesheet?.userId).toBe('owner-uuid')
  })

  it('accepts PascalCase UserID', () => {
    setActivePinia(createPinia())
    const store = useCraStore()
    store.setTimesheet({
      ID: 'ts-2',
      UserID: 'owner-2',
      Month: '2026-08',
      Status: 'Brouillon',
      Weeks: []
    })
    expect(store.timesheet?.id).toBe('ts-2')
    expect(store.timesheet?.userId).toBe('owner-2')
  })
})
