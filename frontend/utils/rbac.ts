/** Miroir de internal/modules/org/app/service.go DefaultPermissions — garder synchronisé. */
export type RbacModule = 'org' | 'cra' | 'conges' | 'budget' | 'tma' | 'workflow' | 'billing' | 'notifications' | 'reporting' | 'support' | 'maintenance'
export type RbacAction = 'L' | 'E' | 'V'

type ProfilePerms = Partial<Record<RbacModule, Partial<Record<RbacAction, boolean>>>>

const read: Partial<Record<RbacAction, boolean>> = { L: true }
const readWrite: Partial<Record<RbacAction, boolean>> = { L: true, E: true }
const readWriteValidate: Partial<Record<RbacAction, boolean>> = { L: true, E: true, V: true }

const mvpAdmin: ProfilePerms = {
  org: readWriteValidate,
  cra: readWriteValidate,
  tma: readWriteValidate,
  conges: readWriteValidate,
  budget: readWriteValidate,
  workflow: readWriteValidate,
  billing: readWrite,
  notifications: readWrite,
  reporting: read,
  support: readWriteValidate,
  maintenance: readWriteValidate
}

const PROFILE_PERMISSIONS: Record<string, ProfilePerms> = {
  Administrateur: mvpAdmin,
  Collaborateur: {
    cra: readWrite,
    tma: readWrite,
    conges: readWrite,
    budget: read
  },
  "Chef d'équipe": {
    org: read,
    cra: readWriteValidate,
    tma: readWriteValidate,
    conges: read,
    budget: readWrite,
    reporting: read
  },
  'Responsable de service': {
    org: read,
    cra: readWriteValidate,
    tma: readWriteValidate,
    conges: readWriteValidate,
    budget: readWriteValidate,
    reporting: read
  }
}

export function rbacCan(profile: string | undefined, module: RbacModule, action: RbacAction): boolean {
  if (!profile) return false
  return PROFILE_PERMISSIONS[profile]?.[module]?.[action] ?? false
}
