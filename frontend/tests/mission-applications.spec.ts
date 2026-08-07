import { describe, expect, it } from 'vitest'

/** Mirrors MissionApplicationMultiSelect toggle rules. */
function canToggleMissionApplication(active: boolean | undefined, currentlyChecked: boolean): boolean {
  // Inactive apps can only be removed (uncheck), not newly selected.
  if (active === false && !currentlyChecked) return false
  return true
}

type AppOption = { id: string; libelle: string; active?: boolean }

/** Active catalog ∪ currently linked (incl. inactive) for edit form. */
function mergeMissionApplicationCatalog(
  activeCatalog: AppOption[],
  linked: AppOption[]
): AppOption[] {
  const byId = new Map<string, AppOption>()
  for (const a of activeCatalog) {
    if (!a.id || a.active === false) continue
    byId.set(a.id, { ...a, active: true })
  }
  for (const linkedApp of linked) {
    if (!linkedApp.id || byId.has(linkedApp.id)) continue
    byId.set(linkedApp.id, {
      id: linkedApp.id,
      libelle: linkedApp.libelle || linkedApp.id,
      active: linkedApp.active !== false
    })
  }
  return [...byId.values()].sort((a, b) => a.libelle.localeCompare(b.libelle, 'fr'))
}

describe('mission applications selection rules', () => {
  it('blocks newly selecting an inactive app', () => {
    expect(canToggleMissionApplication(false, false)).toBe(false)
  })

  it('allows unchecking an inactive linked app', () => {
    expect(canToggleMissionApplication(false, true)).toBe(true)
  })

  it('allows toggling active apps', () => {
    expect(canToggleMissionApplication(true, false)).toBe(true)
    expect(canToggleMissionApplication(true, true)).toBe(true)
    expect(canToggleMissionApplication(undefined, false)).toBe(true)
  })

  it('merges active catalog with linked inactive apps', () => {
    const merged = mergeMissionApplicationCatalog(
      [
        { id: 'a1', libelle: 'Alpha', active: true },
        { id: 'a2', libelle: 'Zed', active: false }
      ],
      [
        { id: 'a2', libelle: 'Zed', active: false },
        { id: 'a3', libelle: 'Legacy', active: false }
      ]
    )
    expect(merged.map((a) => a.id)).toEqual(['a1', 'a3', 'a2'])
    expect(merged.find((a) => a.id === 'a3')?.active).toBe(false)
    expect(merged.find((a) => a.id === 'a1')?.active).toBe(true)
  })
})
