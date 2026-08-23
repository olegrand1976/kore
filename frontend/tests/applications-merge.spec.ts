import { describe, expect, it } from 'vitest'
import {
  canMergeApplications,
  defaultMergeReferenceId,
  hasMergeMethodologyMismatch,
  isMergeReferenceLocked,
  mergeAbsorbedId,
  type MergeApplicationRow
} from '~/utils/applicationMerge'

const row = (
  id: string,
  libelle: string,
  methodologyProfile: MergeApplicationRow['methodologyProfile'] = 'psa',
  taigaLinked = false
): MergeApplicationRow => ({
  id,
  libelle,
  methodologyProfile,
  taigaLinked
})

describe('application merge rules', () => {
  it('locks reference when one app is Taiga-linked', () => {
    const selected = [row('taiga', 'Taiga app', 'psa', true), row('manual', 'Manual')]
    expect(isMergeReferenceLocked(selected)).toBe(true)
    expect(defaultMergeReferenceId(selected)).toBe('taiga')
  })

  it('defaults reference to first app when none are Taiga-linked', () => {
    const selected = [row('a', 'Alpha'), row('b', 'Beta')]
    expect(isMergeReferenceLocked(selected)).toBe(false)
    expect(defaultMergeReferenceId(selected)).toBe('a')
  })

  it('detects methodology mismatch against reference', () => {
    const selected = [row('a', 'Alpha', 'psa'), row('b', 'Beta', 'agile_scrum')]
    expect(hasMergeMethodologyMismatch(selected, 'a')).toBe(true)
    expect(hasMergeMethodologyMismatch(selected, 'b')).toBe(true)
  })

  it('returns absorbed id as the non-reference app', () => {
    const selected = [row('a', 'Alpha'), row('b', 'Beta')]
    expect(mergeAbsorbedId(selected, 'a')).toBe('b')
    expect(mergeAbsorbedId(selected, 'b')).toBe('a')
  })

  it('blocks merge when both apps are Taiga-linked', () => {
    const selected = [row('a', 'Taiga A', 'psa', true), row('b', 'Taiga B', 'psa', true)]
    expect(canMergeApplications(selected)).toBe(false)
  })
})
