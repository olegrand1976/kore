export type AuthProfile = 'Administrateur' | 'Collaborateur' | 'Utilisateur'

type SessionUser = {
  ok: boolean
  profile?: AuthProfile
  userId?: string
  tenantId?: string
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

  return { user, fetchSession, isAdmin }
}
