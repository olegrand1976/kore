export type OrgUserSummary = {
  id?: string
  ID?: string
  login?: string
  Login?: string
  profil?: string
  Profil?: string
  profils?: string[]
  Profiles?: string[]
  active?: boolean
  Active?: boolean
  equipeId?: string
  EquipeID?: string
  equipeIds?: string[]
  EquipeIDs?: string[]
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

function pickUserProfiles(item: OrgUserSummary): string[] {
  const multi = item.profils ?? item.Profiles
  if (Array.isArray(multi) && multi.length > 0) {
    return multi.map(String)
  }
  const single = pickUserProfile(item)
  return single ? [single] : []
}

function pickUserActive(item: OrgUserSummary) {
  return item.active ?? item.Active ?? true
}

function pickUserEquipeId(item: OrgUserSummary) {
  return item.equipeId ?? item.EquipeID ?? ''
}

function pickUserEquipeIds(item: OrgUserSummary): string[] {
  const multi = item.equipeIds ?? item.EquipeIDs
  if (Array.isArray(multi) && multi.length > 0) {
    return multi.map(String)
  }
  const single = pickUserEquipeId(item)
  return single ? [single] : []
}

export function useUsers() {
  const { apiFetch } = useApiFetch()
  const list = async () => {
    const res = await apiFetch<{ data?: OrgUserSummary[] }>('/api/org/users')
    const payload = res?.data ?? res
    return Array.isArray(payload) ? payload : []
  }

  const create = async (body: {
    login: string
    password: string
    profils: string[]
    equipeIds?: string[]
  }) => {
    return apiFetch('/api/org/users', { method: 'POST', body })
  }

  const update = async (
    id: string,
    body: {
      profils?: string[]
      password?: string
      active?: boolean
      equipeIds?: string[]
    }
  ) => {
    return apiFetch(`/api/org/users/${id}`, { method: 'PUT', body })
  }

  const deactivate = async (id: string) => {
    return apiFetch(`/api/org/users/${id}/deactivate`, { method: 'PATCH' })
  }

  const remove = async (id: string) => {
    return apiFetch(`/api/org/users/${id}`, { method: 'DELETE' })
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
    pickUserProfiles,
    pickUserActive,
    pickUserEquipeId,
    pickUserEquipeIds
  }
}
