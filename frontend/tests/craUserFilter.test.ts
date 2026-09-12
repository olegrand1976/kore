import { describe, expect, it } from 'vitest'
import { buildCraUserFilterOptions } from '../utils/craUserFilter'

describe('buildCraUserFilterOptions', () => {
  it('lists active CRA-required org users, not only those present in CRA rows', () => {
    const opts = buildCraUserFilterOptions(
      [
        { id: 'u1', prenom: 'Alice', nom: 'Martin', login: 'alice', active: true, craRequis: true },
        { id: 'u2', prenom: 'Bob', nom: 'Dupont', login: 'bob', active: true, craRequis: true },
        { id: 'u3', prenom: 'Old', nom: 'User', login: 'old', active: false, craRequis: true },
        { id: 'u4', prenom: 'No', nom: 'Cra', login: 'nocra', active: true, craRequis: false }
      ],
      [{ id: 'u1', label: 'Alice from CRA' }],
      'me',
      'Moi'
    )
    expect(opts.map((o) => o.value)).toEqual(['u1', 'u2', 'me'])
    expect(opts.find((o) => o.value === 'u1')?.label).toBe('Alice Martin')
    expect(opts.find((o) => o.value === 'me')?.label).toBe('Moi')
  })

  it('keeps inactive/non-CRA users when they appear in extras', () => {
    const opts = buildCraUserFilterOptions(
      [{ id: 'u4', prenom: 'No', nom: 'Cra', login: 'nocra', active: true, craRequis: false }],
      [{ id: 'u4', label: 'No Cra (from timesheet)' }],
      '',
      ''
    )
    expect(opts).toEqual([{ value: 'u4', label: 'No Cra (from timesheet)' }])
  })
})
