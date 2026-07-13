export function useFicheFormat() {
  const { locale, t } = useI18n()

  const formatDate = (raw?: string | null) => {
    if (!raw) return '—'
    const date = new Date(raw)
    if (Number.isNaN(date.getTime())) return raw
    return date.toLocaleDateString(locale.value === 'en' ? 'en-US' : 'fr-FR', {
      day: 'numeric',
      month: 'long',
      year: 'numeric'
    })
  }

  const formatMoney = (amountCents: number, currency = 'EUR') => {
    const value = amountCents / 100
    return new Intl.NumberFormat(locale.value === 'en' ? 'en-US' : 'fr-FR', {
      style: 'currency',
      currency,
      maximumFractionDigits: 0
    }).format(value)
  }

  const missionStatusLabel = (status: string) => {
    switch (status) {
      case 'active':
        return t('fiche.mission_status_active')
      case 'arretee':
        return t('fiche.mission_status_stopped')
      case 'terminee':
        return t('fiche.mission_status_done')
      default:
        return status
    }
  }

  const missionStatusVariant = (status: string): 'success' | 'warn' | 'neutral' | 'default' => {
    switch (status) {
      case 'active':
        return 'success'
      case 'arretee':
        return 'warn'
      case 'terminee':
        return 'neutral'
      default:
        return 'default'
    }
  }

  return { formatDate, formatMoney, missionStatusLabel, missionStatusVariant }
}
