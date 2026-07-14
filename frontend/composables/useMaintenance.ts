export type WorkRequest = {
  id?: string
  ID?: string
  applicationId?: string
  ApplicationID?: string
  subject?: string
  Subject?: string
  description?: string
  Description?: string
  priority?: string
  Priority?: string
  dueAt?: string
  DueAt?: string
  state?: string
  State?: string
  assigneeId?: string
  AssigneeID?: string
  consumptionDays?: number
  ConsumptionDays?: number
  createdAt?: string
  CreatedAt?: string
  completedAt?: string
  CompletedAt?: string
}

export type CreateWorkRequestPayload = {
  applicationId: string
  subject: string
  description?: string
  priority?: string
  dueAt?: string
}

function toDueAtISO(raw?: string) {
  if (!raw) return undefined
  const parsed = new Date(raw)
  if (Number.isNaN(parsed.getTime())) return undefined
  return parsed.toISOString()
}

export function useMaintenance() {
  const pickId = (w: WorkRequest) => w.id ?? w.ID ?? ''
  const pickSubject = (w: WorkRequest) => w.subject ?? w.Subject ?? ''
  const pickState = (w: WorkRequest) => w.state ?? w.State ?? ''
  const pickConsumption = (w: WorkRequest) => w.consumptionDays ?? w.ConsumptionDays ?? 0
  const pickPriority = (w: WorkRequest) => w.priority ?? w.Priority ?? 'normal'
  const pickDueAt = (w: WorkRequest) => w.dueAt ?? w.DueAt ?? ''
  const pickApplicationId = (w: WorkRequest) => w.applicationId ?? w.ApplicationID ?? ''
  const pickAssigneeId = (w: WorkRequest) => w.assigneeId ?? w.AssigneeID ?? ''

  const list = async () => {
    const res = await $fetch<{ data?: WorkRequest[] }>('/api/work-requests')
    return res?.data ?? []
  }

  const get = async (id: string) => {
    const res = await $fetch<{ data?: WorkRequest }>(`/api/work-requests/${id}`)
    return (res?.data ?? res) as WorkRequest
  }

  const create = async (payload: CreateWorkRequestPayload) => {
    const res = await $fetch<{ data?: WorkRequest }>('/api/work-requests', {
      method: 'POST',
      body: {
        ...payload,
        dueAt: toDueAtISO(payload.dueAt)
      }
    })
    return (res?.data ?? res) as WorkRequest
  }

  const assign = async (id: string, assigneeId: string) => {
    const res = await $fetch<{ data?: WorkRequest }>(`/api/work-requests/${id}/assign`, {
      method: 'POST',
      body: { assigneeId }
    })
    return (res?.data ?? res) as WorkRequest
  }

  const progress = async (id: string, consumptionDays: number) => {
    const res = await $fetch<{ data?: WorkRequest }>(`/api/work-requests/${id}/progress`, {
      method: 'POST',
      body: { consumptionDays }
    })
    return (res?.data ?? res) as WorkRequest
  }

  const complete = async (id: string) => {
    const res = await $fetch<{ data?: WorkRequest }>(`/api/work-requests/${id}/complete`, { method: 'POST' })
    return (res?.data ?? res) as WorkRequest
  }

  return {
    list,
    get,
    create,
    assign,
    progress,
    complete,
    pickId,
    pickSubject,
    pickState,
    pickConsumption,
    pickPriority,
    pickDueAt,
    pickApplicationId,
    pickAssigneeId
  }
}
