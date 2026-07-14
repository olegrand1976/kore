export type CraWorkRefType = 'tma' | 'ticket' | 'work_request'

export type CraWorkRefOption = {
  type: CraWorkRefType
  id: string
  label: string
}

const TERMINAL_TMA = new Set(['resolue'])
const TERMINAL_TICKET = new Set(['resolved'])
const TERMINAL_WORK_REQUEST = new Set(['completed'])

const pickAssigneeId = (row: Record<string, unknown>) =>
  String(row.assigneeId ?? row.AssigneeID ?? '')

const pickId = (row: Record<string, unknown>) => String(row.id ?? row.ID ?? '')

const pickSubject = (row: Record<string, unknown>) =>
  String(row.subject ?? row.Subject ?? '').trim()

const pickStatus = (row: Record<string, unknown>) =>
  String(row.status ?? row.Status ?? row.state ?? row.State ?? '')

export const encodeWorkRef = (type: string, id: string) => (type && id ? `${type}:${id}` : '')

export const decodeWorkRef = (value: string): { type: string; id: string } => {
  if (!value) return { type: '', id: '' }
  const idx = value.indexOf(':')
  if (idx <= 0) return { type: '', id: '' }
  return { type: value.slice(0, idx), id: value.slice(idx + 1) }
}

export function useCraWorkRefs() {
  const { t } = useI18n()
  const options = ref<CraWorkRefOption[]>([])

  const typeLabel = (type: CraWorkRefType) => {
    switch (type) {
      case 'tma':
        return t('cra.source_tma')
      case 'ticket':
        return t('cra.source_ticket')
      case 'work_request':
        return t('cra.source_work_request')
      default: {
        const _exhaustive: never = type
        return _exhaustive
      }
    }
  }

  const labelFor = (type: string, id: string) => {
    if (!type || !id) return ''
    const found = options.value.find((opt) => opt.type === type && opt.id === id)
    if (found) return found.label
    const prefix = typeLabel(type as CraWorkRefType)
    return `${prefix} #${id.slice(0, 8)}`
  }

  const load = async (userId: string) => {
    if (!userId) {
      options.value = []
      return
    }

    const [tmaRes, ticketRes, workRes] = await Promise.allSettled([
      $fetch<{ data?: Record<string, unknown>[] }>('/api/tma/demands'),
      $fetch<{ data?: Record<string, unknown>[] }>('/api/tickets'),
      $fetch<{ data?: Record<string, unknown>[] }>('/api/work-requests')
    ])

    const next: CraWorkRefOption[] = []

    if (tmaRes.status === 'fulfilled') {
      for (const row of tmaRes.value.data ?? []) {
        const id = pickId(row)
        if (!id || pickAssigneeId(row) !== userId) continue
        const status = pickStatus(row)
        if (TERMINAL_TMA.has(status)) continue
        const subject = pickSubject(row) || id.slice(0, 8)
        next.push({ type: 'tma', id, label: `${t('cra.source_tma')} — ${subject}` })
      }
    }

    if (ticketRes.status === 'fulfilled') {
      for (const row of ticketRes.value.data ?? []) {
        const id = pickId(row)
        if (!id || pickAssigneeId(row) !== userId) continue
        const status = pickStatus(row)
        if (TERMINAL_TICKET.has(status)) continue
        const subject = pickSubject(row) || id.slice(0, 8)
        next.push({ type: 'ticket', id, label: `${t('cra.source_ticket')} — ${subject}` })
      }
    }

    if (workRes.status === 'fulfilled') {
      for (const row of workRes.value.data ?? []) {
        const id = pickId(row)
        if (!id || pickAssigneeId(row) !== userId) continue
        const status = pickStatus(row)
        if (TERMINAL_WORK_REQUEST.has(status)) continue
        const subject = pickSubject(row) || id.slice(0, 8)
        next.push({ type: 'work_request', id, label: `${t('cra.source_work_request')} — ${subject}` })
      }
    }

    next.sort((a, b) => a.label.localeCompare(b.label, undefined, { sensitivity: 'base' }))
    options.value = next
  }

  const groupedOptions = computed(() => {
    const groups: Record<CraWorkRefType, CraWorkRefOption[]> = {
      tma: [],
      ticket: [],
      work_request: []
    }
    for (const opt of options.value) {
      groups[opt.type].push(opt)
    }
    return groups
  })

  return { options, groupedOptions, load, labelFor, encodeWorkRef, decodeWorkRef }
}
