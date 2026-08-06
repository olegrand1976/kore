/** Miroir de internal/modules/org/app/service.go DefaultPermissions — garder synchronisé. */
export type RbacModule =
  | 'org'
  | 'cra'
  | 'conges'
  | 'budget'
  | 'tma'
  | 'workflow'
  | 'billing'
  | 'notifications'
  | 'reporting'
  | 'support'
  | 'maintenance'
  | 'integrations'
  | 'invoicing'
  | 'admin'
  | 'ssii'
  | 'ett'
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
  maintenance: readWriteValidate,
  integrations: readWriteValidate,
  invoicing: readWriteValidate,
  admin: readWriteValidate,
  ssii: readWriteValidate,
  ett: readWriteValidate
}

/** Profils avec permissions MVP branchées (API + front). */
export const IMPLEMENTED_RBAC_PROFILES = [
  'Administrateur',
  'Collaborateur',
  "Chef d'équipe",
  'Responsable de service'
] as const

export type ImplementedRbacProfile = (typeof IMPLEMENTED_RBAC_PROFILES)[number]

/** Profils catalogue SFD pas encore branchés dans DefaultPermissions. */
export const PLANNED_RBAC_PROFILES = [
  'Utilisateur',
  'Commercial',
  'Support',
  'Chef utilisateur',
  'Client externe',
  'Sous-traitant'
] as const

/** Modules affichés dans le guide d'accès in-app (alignés DefaultPermissions). */
export const HELP_MATRIX_MODULES: RbacModule[] = [
  'org',
  'cra',
  'tma',
  'conges',
  'budget',
  'reporting',
  'support',
  'maintenance',
  'workflow',
  'billing',
  'notifications',
  'integrations',
  'invoicing',
  'admin',
  'ssii',
  'ett'
]

export const PROFILE_PERMISSIONS: Record<string, ProfilePerms> = {
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

export function isImplementedRbacProfile(profile: string): profile is ImplementedRbacProfile {
  return (IMPLEMENTED_RBAC_PROFILES as readonly string[]).includes(profile)
}

export function rbacCan(
  profile: string | string[] | undefined,
  module: RbacModule,
  action: RbacAction
): boolean {
  const profiles = Array.isArray(profile) ? profile : profile ? [profile] : []
  return profiles.some((p) => PROFILE_PERMISSIONS[p]?.[module]?.[action] ?? false)
}

/** Cellule matrice L/E/V ou tiret si aucun droit. */
export function formatRbacCell(profile: string, module: RbacModule): string {
  const perms = PROFILE_PERMISSIONS[profile]?.[module]
  if (!perms) return '—'
  const parts: RbacAction[] = []
  if (perms.L) parts.push('L')
  if (perms.E) parts.push('E')
  if (perms.V) parts.push('V')
  return parts.length > 0 ? parts.join('/') : '—'
}
