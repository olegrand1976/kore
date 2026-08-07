import type { Ref } from 'vue'

export type BillingCountryCode = 'FR' | 'BE' | 'MG' | 'MA' | 'TN' | 'CA' | ''

export type ClientBillingFields = {
  raisonSociale: string
  tva: string
  pays: BillingCountryCode
  adresse: string
  adresseNumero: string
  adresseBoite: string
  codePostal: string
  ville: string
  siret: string
}

export function emptyClientBillingFields(): ClientBillingFields {
  return {
    raisonSociale: '',
    tva: '',
    pays: '',
    adresse: '',
    adresseNumero: '',
    adresseBoite: '',
    codePostal: '',
    ville: '',
    siret: ''
  }
}

export function normalizeBillingCountry(value: unknown): BillingCountryCode {
  const code = typeof value === 'string' ? value.trim().toUpperCase() : ''
  switch (code) {
    case 'BE':
      return 'BE'
    case 'FR':
      return 'FR'
    case 'MG':
    case 'MD':
      return 'MG'
    case 'MA':
      return 'MA'
    case 'TN':
      return 'TN'
    case 'CA':
      return 'CA'
    case '':
      return ''
    default:
      return ''
  }
}

export function clientBillingPayload(form: ClientBillingFields) {
  // Full-replace body: always send every field so omitted JSON does not silently
  // clear billing data on PUT (empty string = clear intentional).
  return {
    raisonSociale: form.raisonSociale.trim(),
    tva: form.tva.trim(),
    pays: form.pays,
    adresse: form.adresse.trim(),
    adresseNumero: form.adresseNumero.trim(),
    adresseBoite: form.adresseBoite.trim(),
    codePostal: form.codePostal.trim(),
    ville: form.ville.trim(),
    siret: form.siret.trim()
  }
}

/**
 * Labels / placeholders registre (SIRET, BCE, …) selon le pays de facturation.
 * Réutilise les clés i18n `org.*` déjà en place pour la société tenant.
 */
export function useBillingCountryLabels(pays: Ref<BillingCountryCode> | (() => BillingCountryCode)) {
  const { t } = useI18n()
  const current = computed(() => (typeof pays === 'function' ? pays() : pays.value))

  const registryLabel = computed(() => {
    switch (current.value) {
      case 'BE':
        return t('org.bce')
      case 'FR':
        return t('org.siret')
      case 'MG':
        return t('org.nif_stat')
      case 'MA':
        return t('org.ice')
      case 'TN':
        return t('org.matricule_fiscal')
      case 'CA':
        return t('org.ne')
      case '':
        return t('clients.registry')
      default: {
        const _exhaustive: never = current.value
        return _exhaustive
      }
    }
  })

  const registryPlaceholder = computed(() => {
    switch (current.value) {
      case 'BE':
        return '0123456789'
      case 'FR':
        return '12345678901234'
      case '':
        return ''
      case 'MG':
        return '1234567890'
      case 'MA':
        return '123456789012345'
      case 'TN':
        return '1234567/A/A/A/000'
      case 'CA':
        return '123456789'
      default: {
        const _exhaustive: never = current.value
        return _exhaustive
      }
    }
  })

  const registryHint = computed(() => {
    switch (current.value) {
      case 'BE':
        return t('org.registry_hint_be')
      case 'FR':
        return t('org.registry_hint_fr')
      case '':
        return t('clients.registry_hint')
      case 'MG':
        return t('org.registry_hint_mg')
      case 'MA':
        return t('org.registry_hint_ma')
      case 'TN':
        return t('org.registry_hint_tn')
      case 'CA':
        return t('org.registry_hint_ca')
      default: {
        const _exhaustive: never = current.value
        return _exhaustive
      }
    }
  })

  const addressBoxLabel = computed(() => {
    switch (current.value) {
      case 'BE':
      case 'FR':
      case '':
        return t('org.address_box')
      case 'CA':
        return t('org.address_box_ca')
      case 'MG':
      case 'MA':
      case 'TN':
        return t('org.address_box_other')
      default: {
        const _exhaustive: never = current.value
        return _exhaustive
      }
    }
  })

  return { registryLabel, registryPlaceholder, registryHint, addressBoxLabel }
}
