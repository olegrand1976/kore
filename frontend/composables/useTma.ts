export type TmaDemand = {
  id?: string
  ID?: string
  applicationId?: string
  ApplicationID?: string
  subject?: string
  Subject?: string
  status?: string
  Status?: string
  visible?: boolean
  Visible?: boolean
  workflowInstanceId?: string
  WorkflowInstanceID?: string
  assigneeId?: string
  AssigneeID?: string
  requiresChefGate?: boolean
  RequiresChefGate?: boolean
  createdAt?: string
  CreatedAt?: string
}

export type TmaGanttItem = {
  id: string
  label: string
  status: string
  start: Date
  end: Date
}

export type TmaAnalysis = {
  functional?: string
  Functional?: string
  technical?: string
  Technical?: string
  risks?: string
  Risks?: string
  testScenario?: string
  TestScenario?: string
}

export function useTma() {
  const pickId = (d: TmaDemand) => d.id ?? d.ID ?? ''
  const pickSubject = (d: TmaDemand) => d.subject ?? d.Subject ?? ''
  const pickStatus = (d: TmaDemand) => d.status ?? d.Status ?? ''
  const pickWorkflowId = (d: TmaDemand) => d.workflowInstanceId ?? d.WorkflowInstanceID ?? ''
  const pickCreatedAt = (d: TmaDemand) => d.createdAt ?? d.CreatedAt ?? ''

  const toGanttItem = (d: TmaDemand): TmaGanttItem | null => {
    const id = pickId(d)
    if (!id) return null
    const status = pickStatus(d)
    const rawStart = pickCreatedAt(d)
    const now = new Date()
    const start = rawStart ? new Date(rawStart) : new Date(now.getTime() - 7 * 86400000)
    if (Number.isNaN(start.getTime())) return null
    const end = status === 'resolue'
      ? new Date(Math.max(now.getTime(), start.getTime() + 86400000))
      : new Date(Math.max(now.getTime(), start.getTime() + 14 * 86400000))
    return {
      id,
      label: pickSubject(d) || id,
      status,
      start,
      end
    }
  }

  const toGanttItems = (demands: TmaDemand[]) =>
    demands.map(toGanttItem).filter((item): item is TmaGanttItem => item !== null)

  const list = async () => {
    const res = await $fetch<{ data?: TmaDemand[] }>('/api/tma/demands')
    return res?.data ?? []
  }

  const get = async (id: string) => {
    const res = await $fetch<{ data?: TmaDemand }>(`/api/tma/demands/${id}`)
    return (res?.data ?? res) as TmaDemand
  }

  const getAnalysis = async (id: string) => {
    const res = await $fetch<{ data?: TmaAnalysis }>(`/api/tma/demands/${id}/analysis`)
    return (res?.data ?? res) as TmaAnalysis
  }

  const create = async (payload: { applicationId: string; subject: string; requiresChefGate?: boolean }) => {
    return $fetch('/api/tma/demands', { method: 'POST', body: payload })
  }

  const validateCreation = (id: string) =>
    $fetch(`/api/tma/demands/${id}/validate-creation`, { method: 'POST' })

  const assign = (id: string, assigneeId: string) =>
    $fetch(`/api/tma/demands/${id}/assign`, { method: 'POST', body: { assigneeId } })

  const takeOver = (id: string) =>
    $fetch(`/api/tma/demands/${id}/take-over`, { method: 'POST' })

  const saveAnalysis = (id: string, analysis: TmaAnalysis) =>
    $fetch(`/api/tma/demands/${id}/analysis`, {
      method: 'POST',
      body: {
        functional: analysis.functional ?? analysis.Functional ?? '',
        technical: analysis.technical ?? analysis.Technical ?? '',
        risks: analysis.risks ?? analysis.Risks ?? '',
        testScenario: analysis.testScenario ?? analysis.TestScenario ?? ''
      }
    })

  const resolve = (id: string) =>
    $fetch(`/api/tma/demands/${id}/resolve`, { method: 'POST' })

  const reopen = (id: string, reason: string) =>
    $fetch(`/api/tma/demands/${id}/reopen`, { method: 'POST', body: { reason } })

  const exportXml = () => window.open('/api/tma/demands/export.xml', '_blank')

  return {
    list,
    get,
    getAnalysis,
    create,
    validateCreation,
    assign,
    takeOver,
    saveAnalysis,
    resolve,
    reopen,
    exportXml,
    pickId,
    pickSubject,
    pickStatus,
    pickWorkflowId,
    pickCreatedAt,
    toGanttItem,
    toGanttItems
  }
}
