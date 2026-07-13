export type MissionSummary = {
  id: string
  clientName?: string
  clientId?: string
}

const SOURCE_LABELS: Record<string, string> = {
  manual: 'cra.source_manual',
  mission: 'cra.source_mission',
  leave: 'cra.source_leave',
  tma: 'cra.source_tma',
  ticket: 'cra.source_ticket',
  holiday: 'cra.source_holiday',
  interne: 'cra.source_internal',
  formation: 'cra.source_training'
}

export function useCraSourceLabels(missions: Ref<MissionSummary[]>) {
  const { t } = useI18n()

  const missionMap = computed(() => {
    const map = new Map<string, MissionSummary>()
    for (const m of missions.value) {
      map.set(m.id, m)
    }
    return map
  })

  const labelFor = (sourceType: string, sourceId: string) => {
    if (sourceType === 'mission') {
      const mission = missionMap.value.get(sourceId)
      if (mission?.clientName) return mission.clientName
    }
    const key = SOURCE_LABELS[sourceType] ?? SOURCE_LABELS.manual
    const base = t(key)
    if (sourceType === 'tma' || sourceType === 'ticket') {
      return `${base} #${sourceId.slice(0, 8)}`
    }
    if (sourceType === 'manual' && sourceId !== 'default') {
      return `${base} (${sourceId})`
    }
    return base
  }

  const iconFor = (sourceType: string): string => {
    switch (sourceType) {
      case 'mission':
        return 'business'
      case 'leave':
        return 'event_busy'
      case 'tma':
      case 'ticket':
        return 'support'
      case 'holiday':
        return 'celebration'
      case 'interne':
      case 'formation':
        return 'school'
      default:
        return 'schedule'
    }
  }

  return { labelFor, iconFor }
}
