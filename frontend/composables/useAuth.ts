export type AuthProfile = 'Administrateur' | 'Collaborateur' | 'Utilisateur' | string

type SessionUser = {
  ok: boolean
  profile?: AuthProfile
  profiles?: AuthProfile[]
  userId?: string
  tenantId?: string
  isPlatformAdmin?: boolean
}

export function useAuth() {
  const user = useState<SessionUser | null>('auth-user', () => null)
  // Forward incoming Cookie header during SSR (plain $fetch does not).
  const requestFetch = useRequestFetch()

  const fetchSession = async () => {
    try {
      user.value = await requestFetch<SessionUser>('/api/auth/session')
    } catch {
      user.value = null
    }
    return user.value
  }

  const effectiveProfiles = computed(() => {
    const multi = user.value?.profiles
    if (Array.isArray(multi) && multi.length > 0) return multi
    const single = user.value?.profile
    return single ? [single] : []
  })

  const isAdmin = computed(() => effectiveProfiles.value.includes('Administrateur'))
  const isPlatformAdmin = computed(() => user.value?.isPlatformAdmin === true)

  const isManager = computed(() =>
    effectiveProfiles.value.some(
      (profile) =>
        profile === 'Administrateur' || profile.includes('Chef') || profile.includes('Responsable')
    )
  )

  return { user, fetchSession, isAdmin, isManager, isPlatformAdmin }
}
