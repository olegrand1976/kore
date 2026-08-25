import { beforeEach, describe, expect, it, vi } from 'vitest'
import { computed, ref } from 'vue'
import { useLeaveTypeConfigs } from '~/composables/useLeave'

const apiFetch = vi.fn()
const stateByKey = new Map<string, ReturnType<typeof ref>>()

beforeEach(() => {
  apiFetch.mockReset()
  stateByKey.clear()
  vi.stubGlobal('useApiFetch', () => ({ apiFetch }))
  vi.stubGlobal('useState', (key: string, init: () => unknown) => {
    let state = stateByKey.get(key)
    if (!state) {
      state = ref(init())
      stateByKey.set(key, state)
    }
    return state
  })
  vi.stubGlobal('computed', computed)
})

describe('useLeaveTypeConfigs fetchMine', () => {
  it('skips network when types are already loaded', async () => {
    const { fetchMine, types } = useLeaveTypeConfigs()
    types.value = [{ code: 'CP', label: 'Congés', sortOrder: 1, active: true, tracksBalance: true }]

    const result = await fetchMine()

    expect(result).toHaveLength(1)
    expect(apiFetch).not.toHaveBeenCalled()
  })

  it('refetches when force is true', async () => {
    const { fetchMine, types } = useLeaveTypeConfigs()
    types.value = [{ code: 'CP', label: 'Old', sortOrder: 1, active: true, tracksBalance: true }]
    apiFetch.mockResolvedValueOnce({
      data: [{ code: 'RTT', label: 'RTT', sortOrder: 2, active: true, tracksBalance: true }]
    })

    const result = await fetchMine({ force: true })

    expect(apiFetch).toHaveBeenCalledWith('/api/conges/leave-type-configs/mine')
    expect(result.map((t) => t.code)).toEqual(['RTT'])
  })

  it('fetches when cache is empty', async () => {
    apiFetch.mockResolvedValueOnce({
      data: [
        { code: 'RTT', label: 'RTT', sortOrder: 2, active: true, tracksBalance: true },
        { code: 'CP', label: 'CP', sortOrder: 1, active: true, tracksBalance: true }
      ]
    })
    const { fetchMine } = useLeaveTypeConfigs()

    const result = await fetchMine()

    expect(apiFetch).toHaveBeenCalledOnce()
    expect(result.map((t) => t.code)).toEqual(['CP', 'RTT'])
  })
})
