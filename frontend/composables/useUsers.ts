export type OrgUserSummary = {
  id?: string
  ID?: string
  login?: string
  Login?: string
  profil?: string
  Profil?: string
  active?: boolean
  Active?: boolean
}

export const USER_PROFILES = [
  'Administrateur',
  'Collaborateur',
  "Chef d'équipe",
  'Responsable de service'
] as const

export type UserProfile = (typeof USER_PROFILES)[number]

function pickUserId(item: OrgUserSummary) {
  return item.id ?? item.ID ?? ''
}

function pickUserLogin(item: OrgUserSummary) {
  return item.login ?? item.Login ?? ''
}

function pickUserProfile(item: OrgUserSummary) {
  return item.profil ?? item.Profil ?? ''
}

function pickUserActive(item: OrgUserSummary) {
  return item.active ?? item.Active ?? true
}

export function useUsers() {
  const list = async () => {
    const res = await $fetch<{ data?: OrgUserSummary[] }>('/api/org/users')
    const payload = res?.data ?? res
    return Array.isArray(payload) ? payload : []
  }

  const create = async (body: { login: string; password: string; profil: string }) => {
    return $fetch('/api/org/users', { method: 'POST', body })
  }

  const update = async (
    id: string,
    body: { profil?: string; password?: string; active?: boolean }
  ) => {
    return $fetch(`/api/org/users/${id}`, { method: 'PUT', body })
  }

  const deactivate = async (id: string) => {
    return $fetch(`/api/org/users/${id}/deactivate`, { method: 'PATCH' })
  }

  const remove = async (id: string) => {
    return $fetch(`/api/org/users/${id}`, { method: 'DELETE' })
  }

  return {
    list,
    create,
    update,
    deactivate,
    remove,
    pickUserId,
    pickUserLogin,
    pickUserProfile,
    pickUserActive
  }
}
