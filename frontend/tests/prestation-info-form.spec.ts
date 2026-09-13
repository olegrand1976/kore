import { describe, expect, it } from 'vitest'
import {
  canLinkCraMission,
  isKnownMissionLink,
  isManualPrestationEntry,
  missionPrestationPatch,
  prestationInfoComplete
} from '../utils/craPrestation'

describe('prestationInfoComplete', () => {
  it('requires client and mission labels', () => {
    expect(prestationInfoComplete('', '')).toBe(false)
    expect(prestationInfoComplete('ACME', '')).toBe(false)
    expect(prestationInfoComplete('ACME', 'Support')).toBe(true)
    expect(prestationInfoComplete('  ACME  ', '  Support  ')).toBe(true)
  })
})

describe('canLinkCraMission', () => {
  it('allows write while timesheet is editable', () => {
    expect(canLinkCraMission({ canWrite: true, canEditTimesheet: true, prestationComplete: true })).toBe(true)
    expect(canLinkCraMission({ canWrite: true, canEditTimesheet: true, prestationComplete: false })).toBe(true)
  })

  it('allows fill-once after final when prestation incomplete', () => {
    expect(canLinkCraMission({ canWrite: true, canEditTimesheet: false, prestationComplete: false })).toBe(true)
  })

  it('locks after final when prestation complete', () => {
    expect(canLinkCraMission({ canWrite: true, canEditTimesheet: false, prestationComplete: true })).toBe(false)
  })

  it('requires write permission', () => {
    expect(canLinkCraMission({ canWrite: false, canEditTimesheet: true, prestationComplete: false })).toBe(false)
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

describe('isManualPrestationEntry', () => {
  it('treats empty mission id as manual entry', () => {
    expect(isManualPrestationEntry('')).toBe(true)
    expect(isManualPrestationEntry(undefined)).toBe(true)
    expect(isManualPrestationEntry('mission-uuid')).toBe(false)
  })
})

describe('missionPrestationPatch', () => {
  it('normalises mission payload for prestation context', () => {
    const patch = missionPrestationPatch({
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
