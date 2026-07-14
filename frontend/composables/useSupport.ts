export type SupportTicket = {
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
  channel?: string
  Channel?: string
  assigneeId?: string
  AssigneeID?: string
  createdAt?: string
  CreatedAt?: string
  resolvedAt?: string
  ResolvedAt?: string
}

export type TicketReply = {
  id?: string
  ID?: string
  content?: string
  Content?: string
  authorId?: string
  AuthorID?: string
  createdAt?: string
  CreatedAt?: string
}

export type CreateTicketPayload = {
  applicationId: string
  subject: string
  description: string
  priority?: string
  dueAt?: string
}

function toDueAtISO(raw?: string) {
  if (!raw) return undefined
  const parsed = new Date(raw)
  if (Number.isNaN(parsed.getTime())) return undefined
  return parsed.toISOString()
}

export function useSupport() {
  const pickId = (t: SupportTicket) => t.id ?? t.ID ?? ''
  const pickSubject = (t: SupportTicket) => t.subject ?? t.Subject ?? ''
  const pickState = (t: SupportTicket) => t.state ?? t.State ?? ''
  const pickDescription = (t: SupportTicket) => t.description ?? t.Description ?? ''
  const pickPriority = (t: SupportTicket) => t.priority ?? t.Priority ?? 'normal'
  const pickDueAt = (t: SupportTicket) => t.dueAt ?? t.DueAt ?? ''
  const pickApplicationId = (t: SupportTicket) => t.applicationId ?? t.ApplicationID ?? ''
  const pickAssigneeId = (t: SupportTicket) => t.assigneeId ?? t.AssigneeID ?? ''

  const list = async () => {
    const res = await $fetch<{ data?: SupportTicket[] }>('/api/tickets')
    return res?.data ?? []
  }

  const get = async (id: string) => {
    const res = await $fetch<{ data?: SupportTicket }>(`/api/tickets/${id}`)
    return (res?.data ?? res) as SupportTicket
  }

  const create = async (payload: CreateTicketPayload) => {
    const res = await $fetch<{ data?: SupportTicket }>('/api/tickets', {
      method: 'POST',
      body: {
        ...payload,
        dueAt: toDueAtISO(payload.dueAt)
      }
    })
    return (res?.data ?? res) as SupportTicket
  }

  const assign = async (id: string, assigneeId: string) => {
    const res = await $fetch<{ data?: SupportTicket }>(`/api/tickets/${id}/assign`, {
      method: 'POST',
      body: { assigneeId }
    })
    return (res?.data ?? res) as SupportTicket
  }

  const takeOver = async (id: string) => {
    const res = await $fetch<{ data?: SupportTicket }>(`/api/tickets/${id}/take-over`, { method: 'POST' })
    return (res?.data ?? res) as SupportTicket
  }

  const resolve = async (id: string) => {
    const res = await $fetch<{ data?: SupportTicket }>(`/api/tickets/${id}/resolve`, { method: 'POST' })
    return (res?.data ?? res) as SupportTicket
  }

  const addReply = async (id: string, content: string) => {
    const res = await $fetch<{ data?: TicketReply }>(`/api/tickets/${id}/replies`, {
      method: 'POST',
      body: { content }
    })
    return (res?.data ?? res) as TicketReply
  }

  return {
    list,
    get,
    create,
    assign,
    takeOver,
    resolve,
    addReply,
    pickId,
    pickSubject,
    pickState,
    pickDescription,
    pickPriority,
    pickDueAt,
    pickApplicationId,
    pickAssigneeId
  }
}
