import { describe, expect, it } from 'vitest'
import {
  clientBillingPayload,
  emptyClientBillingFields,
  normalizeBillingCountry
} from '../composables/useBillingCountry'

describe('useBillingCountry', () => {
  it('normalizes supported country codes', () => {
    expect(normalizeBillingCountry('be')).toBe('BE')
    expect(normalizeBillingCountry('MD')).toBe('MG')
    expect(normalizeBillingCountry('')).toBe('')
    expect(normalizeBillingCountry('DE')).toBe('')
  })

  it('builds full-replace billing payload', () => {
    const form = emptyClientBillingFields()
    form.raisonSociale = '  Acme  '
    form.tva = ' BE01 '
    form.pays = 'BE'
    form.adresse = 'Rue du Midi'
    form.siret = '0123456789'
    expect(clientBillingPayload(form)).toEqual({
      raisonSociale: 'Acme',
      tva: 'BE01',
      pays: 'BE',
      adresse: 'Rue du Midi',
      adresseNumero: '',
      adresseBoite: '',
      codePostal: '',
      ville: '',
      siret: '0123456789'
    })
  })
})
