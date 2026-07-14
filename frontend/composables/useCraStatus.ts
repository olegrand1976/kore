export type CraStatus = 'Brouillon' | 'ValidéSemaine' | 'Définitif'

export function useCraStatus() {
  const { t } = useI18n()

  const statusLabel = (status: string) => {
    switch (status as CraStatus) {
      case 'Brouillon':
        return t('cra.status_draft')
      case 'ValidéSemaine':
        return t('cra.status_submitted')
      case 'Définitif':
        return t('cra.status_validated')
      default:
        return status
    }
  }

  const statusVariant = (status: string): 'default' | 'success' | 'warning' => {
    switch (status as CraStatus) {
      case 'Définitif':
        return 'success'
      case 'ValidéSemaine':
        return 'warning'
      default:
        return 'default'
    }
  }

  return { statusLabel, statusVariant, currentMonthKey }
}

export function currentMonthKey(): string {
  const now = new Date()
  const y = now.getFullYear()
  const m = String(now.getMonth() + 1).padStart(2, '0')
  return `${y}-${m}`
}
