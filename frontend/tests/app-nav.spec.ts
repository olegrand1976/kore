import { describe, expect, it } from 'vitest'
import {
  buildBottomNavItems,
  buildNavSections,
  isNavItemVisible,
  type NavSectionId,
  type NavVisibilityContext
} from '../utils/appNav'
import type { RbacModule } from '../utils/rbac'

type Item = { to: string; section: NavSectionId; label: string }

const t = (key: string) => key

function item(to: string, section: NavSectionId): Item {
  return { to, section, label: to }
}

function ctx(overrides: Partial<NavVisibilityContext> = {}): NavVisibilityContext {
  return {
    isAdmin: false,
    isPlatformAdmin: false,
    activeChannelCount: 1,
    hasModule: () => true,
    can: () => false,
    isChannelEnabled: () => true,
    isInvoicingEnabled: false,
    ...overrides
  }
}

describe('buildNavSections', () => {
  it('omits empty sections and orders by NAV_SECTIONS', () => {
    const sections = buildNavSections(
      [
        item('/dashboard', 'home'),
        item('/cra', 'time'),
        item('/budget', 'ops'),
        item('/compte', 'account')
      ],
      t
    )
    expect(sections.map((s) => s.id)).toEqual(['home', 'time', 'ops', 'account'])
    expect(sections.find((s) => s.id === 'requests')).toBeUndefined()
  })

  it('shows a single divider before the first secondary section', () => {
    const sections = buildNavSections(
      [
        item('/dashboard', 'home'),
        item('/compte', 'account'),
        item('/admin/organisation', 'organisation'),
        item('/admin/workflows', 'automation')
      ],
      t
    )
    const dividers = sections.filter((s) => s.showDivider)
    expect(dividers).toHaveLength(1)
    expect(dividers[0]?.id).toBe('account')
    expect(sections.find((s) => s.id === 'organisation')?.showDivider).toBe(false)
  })

  it('keeps integrations only in automation once', () => {
    const sections = buildNavSections(
      [
        item('/dashboard', 'home'),
        item('/admin/integrations', 'automation')
      ],
      t
    )
    const routes = sections.flatMap((s) => s.items.map((i) => i.to))
    expect(routes.filter((to) => to === '/admin/integrations')).toEqual(['/admin/integrations'])
    expect(sections.find((s) => s.id === 'automation')?.label).toBe('nav.section_automation')
  })

  it('resolves section labels via i18n keys', () => {
    const sections = buildNavSections([item('/cra', 'time')], t)
    expect(sections[0]?.label).toBe('nav.section_time')
    expect(buildNavSections([item('/dashboard', 'home')], t)[0]?.label).toBeNull()
  })
  it('groups Clients and Missions in ops and preserves input order', () => {
    const sections = buildNavSections(
      [
        item('/clients', 'ops'),
        item('/missions', 'ops'),
        item('/budget', 'ops'),
        item('/facturation', 'ops')
      ],
      t
    )
    expect(sections).toHaveLength(1)
    expect(sections[0]?.id).toBe('ops')
    expect(sections[0]?.items.map((i) => i.to)).toEqual([
      '/clients',
      '/missions',
      '/budget',
      '/facturation'
    ])
  })
})

describe('buildBottomNavItems', () => {
  it('orders dashboard → cra → conges → first request channel', () => {
    const items = buildBottomNavItems([
      item('/budget', 'ops'),
      item('/tma', 'requests'),
      item('/support', 'requests'),
      item('/conges', 'time'),
      item('/cra', 'time'),
      item('/dashboard', 'home'),
      item('/prestations', 'time')
    ])
    expect(items.map((i) => i.to)).toEqual(['/dashboard', '/cra', '/conges', '/tma'])
  })

  it('prefers Nouvelle demande over request channels when multi-canal', () => {
    const items = buildBottomNavItems([
      item('/dashboard', 'home'),
      item('/cra', 'time'),
      item('/conges', 'time'),
      item('/demandes/nouveau', 'home'),
      item('/tma', 'requests'),
      item('/support', 'requests')
    ])
    expect(items.map((i) => i.to)).toEqual([
      '/dashboard',
      '/cra',
      '/conges',
      '/demandes/nouveau'
    ])
  })

  it('falls back to support then maintenance when TMA is absent', () => {
    expect(
      buildBottomNavItems([
        item('/dashboard', 'home'),
        item('/support', 'requests'),
        item('/maintenance', 'requests')
      ]).map((i) => i.to)
    ).toEqual(['/dashboard', '/support'])
  })

  it('does not include budget, prestations, clients or missions', () => {
    const items = buildBottomNavItems([
      item('/dashboard', 'home'),
      item('/cra', 'time'),
      item('/conges', 'time'),
      item('/budget', 'ops'),
      item('/clients', 'ops'),
      item('/missions', 'ops'),
      item('/prestations', 'time')
    ])
    expect(items.map((i) => i.to)).toEqual(['/dashboard', '/cra', '/conges'])
  })
})

describe('isNavItemVisible', () => {
  it('shows Clients when collaborator has CRA read only', () => {
    const visible = isNavItemVisible(
      { rbacAnyOf: ['org', 'ssii', 'cra'] },
      ctx({
        can: (module: RbacModule) => module === 'cra'
      })
    )
    expect(visible).toBe(true)
  })

  it('shows Missions when user has ssii or cra read', () => {
    expect(
      isNavItemVisible({ rbacAnyOf: ['ssii', 'cra'] }, ctx({ can: (m) => m === 'ssii' }))
    ).toBe(true)
    expect(
      isNavItemVisible({ rbacAnyOf: ['ssii', 'cra'] }, ctx({ can: (m) => m === 'budget' }))
    ).toBe(false)
  })

  it('hides Clients when no org/ssii/cra read', () => {
    expect(
      isNavItemVisible(
        { rbacAnyOf: ['org', 'ssii', 'cra'] },
        ctx({ can: () => false })
      )
    ).toBe(false)
  })

  it('still enforces adminOnly and entitlements', () => {
    expect(
      isNavItemVisible({ adminOnly: true }, ctx({ isAdmin: false }))
    ).toBe(false)
    expect(
      isNavItemVisible({ module: 'budget' }, ctx({ hasModule: () => false }))
    ).toBe(false)
  })
})
