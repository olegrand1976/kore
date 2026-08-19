import type { RbacAction, RbacModule } from '~/utils/rbac'

export type NavSectionId =
  | 'home'
  | 'time'
  | 'requests'
  | 'ops'
  | 'account'
  | 'organisation'
  | 'automation'
  | 'system'

export type NavSectionDef = {
  id: NavSectionId
  labelKey: string | null
  secondary: boolean
}

export type NavSectionView<T extends { section: NavSectionId }> = {
  id: NavSectionId
  label: string | null
  items: T[]
  showDivider: boolean
}

/** Single source of truth for sidebar / drawer section order and labels. */
export const NAV_SECTIONS: readonly NavSectionDef[] = [
  { id: 'home', labelKey: null, secondary: false },
  { id: 'time', labelKey: 'nav.section_time', secondary: false },
  { id: 'requests', labelKey: 'nav.section_requests', secondary: false },
  { id: 'ops', labelKey: 'nav.section_ops', secondary: false },
  { id: 'account', labelKey: 'nav.section_account', secondary: true },
  { id: 'organisation', labelKey: 'nav.section_organisation', secondary: true },
  { id: 'automation', labelKey: 'nav.section_automation', secondary: true },
  { id: 'system', labelKey: 'nav.section_system', secondary: true }
] as const

/**
 * Organisation admin sidebar paths, ordered basic → detailed
 * (Société → Site → Service → Application → Équipe → Utilisateur → SSO).
 */
export const ORG_ADMIN_NAV_PATHS = [
  '/admin/organisation',
  '/admin/sites',
  '/admin/services',
  '/admin/applications',
  '/admin/equipes',
  '/admin/users',
  '/admin/identity-providers'
] as const

export type OrgAdminNavPath = (typeof ORG_ADMIN_NAV_PATHS)[number]

const ORG_ADMIN_NAV_RANK = new Map<string, number>(
  ORG_ADMIN_NAV_PATHS.map((path, index) => [path, index])
)

/** Sort organisation nav items by {@link ORG_ADMIN_NAV_PATHS}; unknown paths stay at the end. */
export function orderOrgAdminNavItems<T extends { to: string }>(items: T[]): T[] {
  return [...items].sort((a, b) => {
    const rankA = ORG_ADMIN_NAV_RANK.get(a.to)
    const rankB = ORG_ADMIN_NAV_RANK.get(b.to)
    if (rankA === undefined && rankB === undefined) return 0
    if (rankA === undefined) return 1
    if (rankB === undefined) return -1
    return rankA - rankB
  })
}

const BOTTOM_NAV_CORE = ['/dashboard', '/cra', '/conges'] as const
const BOTTOM_NAV_REQUEST_CHANNELS = ['/tma', '/support', '/maintenance'] as const
const BOTTOM_NAV_NEW_REQUEST = '/demandes/nouveau'

/**
 * Groups filtered nav items into labeled sections.
 * Empty sections are omitted; a divider appears before the first secondary section.
 */
export function buildNavSections<T extends { section: NavSectionId }>(
  items: T[],
  t: (key: string) => string
): NavSectionView<T>[] {
  const grouped = new Map<NavSectionId, T[]>()
  for (const def of NAV_SECTIONS) {
    grouped.set(def.id, [])
  }
  for (const item of items) {
    grouped.get(item.section)?.push(item)
  }

  const sections: NavSectionView<T>[] = []
  let secondaryStarted = false
  for (const def of NAV_SECTIONS) {
    const sectionItems = grouped.get(def.id) ?? []
    if (sectionItems.length === 0) continue
    const showDivider = def.secondary && !secondaryStarted
    if (def.secondary) secondaryStarted = true
    sections.push({
      id: def.id,
      label: def.labelKey ? t(def.labelKey) : null,
      items: sectionItems,
      showDivider
    })
  }
  return sections
}

/**
 * Mobile bottom nav: Dashboard → CRA → Congés → Nouvelle demande (multi-canal)
 * or first request channel when Nouvelle demande is hidden.
 */
export function buildBottomNavItems<T extends { to: string }>(items: T[]): T[] {
  const byTo = new Map(items.map((item) => [item.to, item]))
  const result: T[] = []
  for (const to of BOTTOM_NAV_CORE) {
    const item = byTo.get(to)
    if (item) result.push(item)
  }
  const newRequest = byTo.get(BOTTOM_NAV_NEW_REQUEST)
  if (newRequest) {
    result.push(newRequest)
    return result
  }
  for (const to of BOTTOM_NAV_REQUEST_CHANNELS) {
    const item = byTo.get(to)
    if (item) {
      result.push(item)
      break
    }
  }
  return result
}

export type NavVisibilityItem = {
  adminOnly?: boolean
  platformOnly?: boolean
  module?: string
  rbacModule?: RbacModule
  rbacAnyOf?: RbacModule[]
  requestChannel?: string
  orgInvoicing?: boolean
  multiChannelOnly?: boolean
}

export type NavVisibilityContext = {
  isAdmin: boolean
  isPlatformAdmin: boolean
  activeChannelCount: number
  hasModule: (code: string) => boolean
  can: (module: RbacModule, action: RbacAction) => boolean
  isChannelEnabled: (channel: string) => boolean
  isInvoicingEnabled: boolean
}

/** Shared sidebar / drawer visibility rules (entitlements, RBAC, channels). */
export function isNavItemVisible(item: NavVisibilityItem, ctx: NavVisibilityContext): boolean {
  if (item.platformOnly && !ctx.isPlatformAdmin) return false
  if (item.adminOnly && !ctx.isAdmin) return false
  if (item.multiChannelOnly && ctx.activeChannelCount < 2) return false
  if (item.module && !ctx.hasModule(item.module)) return false
  if (item.rbacModule && !ctx.can(item.rbacModule, 'L')) return false
  if (item.rbacAnyOf && !item.rbacAnyOf.some((mod) => ctx.can(mod, 'L'))) return false
  if (item.requestChannel && !ctx.isChannelEnabled(item.requestChannel)) return false
  if (item.orgInvoicing && !ctx.isInvoicingEnabled) return false
  return true
}
