import { beforeEach, describe, expect, it, vi } from 'vitest'
import { computed, ref } from 'vue'
import { useAuth } from '~/composables/useAuth'

const requestFetch = vi.fn()
const stateByKey = new Map<string, ReturnType<typeof ref>>()

beforeEach(() => {
  requestFetch.mockReset()
  stateByKey.clear()
  vi.stubGlobal('useRequestFetch', () => requestFetch)
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

describe('useAuth fetchSession', () => {
  it('skips network when session already loaded', async () => {
    const { fetchSession, user } = useAuth()
    user.value = {
      ok: true,
      profile: 'Collaborateur',
      userId: 'u1',
      tenantId: 't1'
    }

    const result = await fetchSession()

    expect(result?.userId).toBe('u1')
    expect(requestFetch).not.toHaveBeenCalled()
  })

  it('fetches when force is true even if user is loaded', async () => {
    const { fetchSession, user } = useAuth()
    user.value = { ok: true, profile: 'Collaborateur', userId: 'u1' }
    requestFetch.mockResolvedValueOnce({
      ok: true,
      profile: 'Administrateur',
      userId: 'u1'
    })

    const result = await fetchSession({ force: true })

    expect(requestFetch).toHaveBeenCalledWith('/api/auth/session')
    expect(result?.profile).toBe('Administrateur')
  })

  it('fetches when user is empty', async () => {
    requestFetch.mockResolvedValueOnce({
      ok: true,
      profile: 'Utilisateur',
      userId: 'u2'
    })
    const { fetchSession } = useAuth()

    const result = await fetchSession()

    expect(requestFetch).toHaveBeenCalledOnce()
    expect(result?.userId).toBe('u2')
  })
})
