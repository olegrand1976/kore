import { describe, expect, it } from 'vitest'
import {
  isKnownMissionLink,
  isManualCommercialEntry,
  missionCommercialPatch,
  prestationInfoComplete
} from '../utils/craCommercial'

describe('prestationInfoComplete', () => {
  it('requires client and mission labels', () => {
    expect(prestationInfoComplete('', '')).toBe(false)
    expect(prestationInfoComplete('ACME', '')).toBe(false)
    expect(prestationInfoComplete('ACME', 'Support')).toBe(true)
    expect(prestationInfoComplete('  ACME  ', '  Support  ')).toBe(true)
  })
})

describe('isKnownMissionLink', () => {
  const missions = [{ id: 'mission-1' }, { id: 'mission-2' }]

  it('locks identity only when the selected mission exists in the list', () => {
    expect(isKnownMissionLink('mission-1', missions)).toBe(true)
    expect(isKnownMissionLink('mission-1', [])).toBe(false)
    expect(isKnownMissionLink('missing', missions)).toBe(false)
  })

  it('treats empty mission id as unlinked', () => {
    expect(isKnownMissionLink('', missions)).toBe(false)
    expect(isKnownMissionLink('   ', missions)).toBe(false)
    expect(isKnownMissionLink(undefined, missions)).toBe(false)
  })
})

describe('isManualCommercialEntry', () => {
  it('treats empty mission id as manual entry', () => {
    expect(isManualCommercialEntry('')).toBe(true)
    expect(isManualCommercialEntry(undefined)).toBe(true)
    expect(isManualCommercialEntry('mission-uuid')).toBe(false)
  })
})

describe('missionCommercialPatch', () => {
  it('normalises mission payload for prestation context', () => {
    const patch = missionCommercialPatch({
      clientName: 'ACME',
      clientId: 'client-1',
      technologies: ['Go', 'Vue'],
      clientContact: 'Jean Dupont'
    })
    expect(patch.client).toBe('ACME')
    expect(patch.clientId).toBe('client-1')
    expect(patch.technologies).toEqual(['Go', 'Vue'])
    expect(patch.responsableClient).toBe('Jean Dupont')
  })
})
