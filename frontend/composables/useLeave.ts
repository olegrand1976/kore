export type LeavePeriod = {
  from?: string
  From?: string
  to?: string
  To?: string
}

export type LeaveRequest = {
  id?: string
  ID?: string
  type?: string
  Type?: string
  from?: string
  From?: string
  to?: string
  To?: string
  period?: LeavePeriod
  Period?: LeavePeriod
  status?: string
  Status?: string
  motif?: string
  Motif?: string
  decidedAt?: string
  DecidedAt?: string
  userId?: string
  UserID?: string
}

export type LeaveBalance = {
  type?: string
  Type?: string
  acquired?: number
  Acquired?: number
  taken?: number
  Taken?: number
  remaining?: number
  Remaining?: number
  balance?: number
  Balance?: number
  unit?: string
  Unit?: string
}

export type LeaveTypeCode = string
export type LeaveStatusCode = 'en_attente' | 'valide' | 'refuse'

export type LeaveTypeConfig = {
  id?: string
  ID?: string
  code?: string
  Code?: string
  label?: string
  Label?: string
  tracksBalance?: boolean
  TracksBalance?: boolean
  active?: boolean
  Active?: boolean
  sortOrder?: number
  SortOrder?: number
}

export function pickLeaveTypeCode(cfg: LeaveTypeConfig) {
  return cfg.code ?? cfg.Code ?? ''
}

export function pickLeaveTypeLabel(cfg: LeaveTypeConfig) {
  return cfg.label ?? cfg.Label ?? pickLeaveTypeCode(cfg)
}

export function pickTracksBalance(cfg: LeaveTypeConfig) {
  return cfg.tracksBalance ?? cfg.TracksBalance ?? false
}

export function pickSortOrder(cfg: LeaveTypeConfig) {
  return cfg.sortOrder ?? cfg.SortOrder ?? 0
}

export function useLeaveTypeConfigs() {
  const types = useState<LeaveTypeConfig[]>('leave-type-configs', () => [])

  const fetchMine = async () => {
    const res = await $fetch<{ data?: LeaveTypeConfig[] }>('/api/conges/leave-type-configs/mine')
    types.value = (res?.data ?? []).slice().sort((a, b) => pickSortOrder(a) - pickSortOrder(b))
    return types.value
  }

  const fetchForSociete = async (societeId: string) => {
    const res = await $fetch<{ data?: LeaveTypeConfig[] }>('/api/conges/leave-type-configs', {
      query: { societeId }
    })
    return (res?.data ?? []).slice().sort((a, b) => pickSortOrder(a) - pickSortOrder(b))
  }

  const typeLabel = (code: string) => {
    const match = types.value.find((item) => pickLeaveTypeCode(item) === code)
    return match ? pickLeaveTypeLabel(match) : code || '—'
  }

  const activeTypes = computed(() =>
    types.value.filter((item) => item.active ?? item.Active ?? true)
  )

  const balanceTypes = computed(() =>
    activeTypes.value.filter((item) => pickTracksBalance(item))
  )

  return {
    types,
    fetchMine,
    fetchForSociete,
    typeLabel,
    activeTypes,
    balanceTypes,
    pickLeaveTypeCode,
    pickLeaveTypeLabel,
    pickTracksBalance,
    pickSortOrder
  }
}

function pickId(item: LeaveRequest) {
  return item.id ?? item.ID ?? ''
}

function pickStatus(item: LeaveRequest) {
  return item.status ?? item.Status ?? ''
}

function pickType(item: LeaveRequest) {
  return item.type ?? item.Type ?? ''
}

function pickMotif(item: LeaveRequest) {
  return item.motif ?? item.Motif ?? ''
}

function pickFrom(item: LeaveRequest) {
  const period = item.period ?? item.Period
  return String(item.from ?? item.From ?? period?.from ?? period?.From ?? '').slice(0, 10)
}

function pickTo(item: LeaveRequest) {
  const period = item.period ?? item.Period
  return String(item.to ?? item.To ?? period?.to ?? period?.To ?? '').slice(0, 10)
}

function pickDecidedAt(item: LeaveRequest) {
  return item.decidedAt ?? item.DecidedAt ?? ''
}

function pickUserId(item: LeaveRequest) {
  return String(item.userId ?? item.UserID ?? '')
}

export function formatLeaveUserLogin(login: string) {
  if (!login || login === '—') return '—'
  const idx = login.indexOf('_')
  return idx >= 0 ? login.slice(idx + 1) : login
}

export function leaveDayCount(from: string, to: string) {
  if (!from || !to) return 0
  const start = new Date(from)
  const end = new Date(to)
  if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime())) return 0
  const diff = Math.round((end.getTime() - start.getTime()) / 86400000)
  return diff >= 0 ? diff + 1 : 0
}

export { pickFrom, pickTo }

export function useLeaveLabels() {
  const { t } = useI18n()
  const { typeLabel: configTypeLabel, types } = useLeaveTypeConfigs()

  const typeLabel = (type: string) => {
    if (types.value.length > 0) return configTypeLabel(type)
    switch (type) {
      case 'conges_payes':
      case 'conges_annuels':
        return t('conges.type_cp')
      case 'rtt':
      case 'recuperation':
        return t('conges.type_rtt')
      case 'maladie':
        return t('conges.type_sick')
      default:
        return type || '—'
    }
  }

  const statusLabel = (status: string) => {
    switch (status as LeaveStatusCode) {
      case 'en_attente':
        return t('conges.status_pending')
      case 'valide':
        return t('conges.status_approved')
      case 'refuse':
        return t('conges.status_rejected')
      default:
        return status || '—'
    }
  }

  const statusVariant = (status: string): 'default' | 'success' | 'warning' | 'error' => {
    switch (status as LeaveStatusCode) {
      case 'valide':
        return 'success'
      case 'en_attente':
        return 'warning'
      case 'refuse':
        return 'error'
      default:
        return 'default'
    }
  }

  return { typeLabel, statusLabel, statusVariant }
}

export function useLeave() {
  const list = async () => {
    const res = await $fetch<{ data?: LeaveRequest[] }>('/api/conges/leave-requests')
    return res?.data ?? []
  }

  const create = async (payload: { type: string; from: string; to: string; motif: string }) => {
    return $fetch('/api/conges/leave-requests', { method: 'POST', body: payload })
  }

  const approve = async (id: string) => {
    return $fetch(`/api/conges/leave-requests/${id}/approve`, { method: 'POST' })
  }

  const reject = async (id: string) => {
    return $fetch(`/api/conges/leave-requests/${id}/reject`, { method: 'POST' })
  }

  const balances = async () => {
    const res = await $fetch<{ data?: LeaveBalance[] }>('/api/conges/leave-balances')
    return res?.data ?? []
  }

  const pending = (items: LeaveRequest[]) =>
    items.filter((item) => pickStatus(item) === 'en_attente')

  return {
    list,
    create,
    approve,
    reject,
    balances,
    pending,
    pickId,
    pickStatus,
    pickType,
    pickMotif,
    pickFrom,
    pickTo,
    pickDecidedAt,
    pickUserId,
    formatLeaveUserLogin,
    leaveDayCount
  }
}
