export type AuthProfile = 'Administrateur' | 'Collaborateur' | 'Utilisateur' | string

type SessionUser = {
  ok: boolean
  profile?: AuthProfile
  userId?: string
  tenantId?: string
  isPlatformAdmin?: boolean
}

export function useAuth() {
  const user = useState<SessionUser | null>('auth-user', () => null)

  const fetchSession = async () => {
    try {
      user.value = await $fetch<SessionUser>('/api/auth/session')
    } catch {
      user.value = null
    }
    return user.value
  }

  const isAdmin = computed(() => user.value?.profile === 'Administrateur')
  const isPlatformAdmin = computed(() => user.value?.isPlatformAdmin === true)

  const isManager = computed(() => {
    const profile = user.value?.profile ?? ''
    return profile === 'Administrateur' || profile.includes('Chef') || profile.includes('Responsable')
  })

  return { user, fetchSession, isAdmin, isManager, isPlatformAdmin }
}
