export type SupportTicket = {
  id?: string
  ID?: string
  applicationId?: string
  ApplicationID?: string
  subject?: string
  Subject?: string
  description?: string
  Description?: string
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

export function useSupport() {
  const pickId = (t: SupportTicket) => t.id ?? t.ID ?? ''
  const pickSubject = (t: SupportTicket) => t.subject ?? t.Subject ?? ''
  const pickState = (t: SupportTicket) => t.state ?? t.State ?? ''
  const pickDescription = (t: SupportTicket) => t.description ?? t.Description ?? ''

  const list = async () => {
    const res = await $fetch<{ data?: SupportTicket[] }>('/api/tickets')
    return res?.data ?? []
  }

  const get = async (id: string) => {
    const res = await $fetch<{ data?: SupportTicket }>(`/api/tickets/${id}`)
    return (res?.data ?? res) as SupportTicket
  }

  const create = async (payload: { applicationId: string; subject: string; description: string }) => {
    const res = await $fetch<{ data?: SupportTicket }>('/api/tickets', { method: 'POST', body: payload })
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

  return { list, get, create, takeOver, resolve, addReply, pickId, pickSubject, pickState, pickDescription }
}
