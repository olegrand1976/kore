import type { MethodologyProfile } from '~/composables/useMethodologyTerms'

export type ProjectEpic = {
  id?: string
  ID?: string
  applicationId?: string
  ApplicationID?: string
  title?: string
  Title?: string
  description?: string
  Description?: string
  status?: string
  Status?: string
  priority?: string
  Priority?: string
}

export type ProjectSprint = {
  id?: string
  ID?: string
  applicationId?: string
  ApplicationID?: string
  name?: string
  Name?: string
  goal?: string
  Goal?: string
  startDate?: string
  StartDate?: string
  endDate?: string
  EndDate?: string
  status?: string
  Status?: string
  capacityPoints?: number | null
  CapacityPoints?: number | null
}

export type BacklogItem = {
  demandId?: string
  DemandID?: string
  subject?: string
  Subject?: string
  status?: string
  Status?: string
  storyPoints?: number | null
  StoryPoints?: number | null
  epicId?: string | null
  EpicID?: string | null
  sprintId?: string | null
  SprintID?: string | null
  backlogRank?: number | null
  BacklogRank?: number | null
}

export type KanbanColumnConfig = {
  stateCode: string
  label?: string
  wipLimit?: number | null
}

export type KanbanConfig = {
  columns?: KanbanColumnConfig[]
  Columns?: KanbanColumnConfig[]
}

export type AgileApplication = {
  id?: string
  ID?: string
  libelle?: string
  Libelle?: string
  methodologyProfile?: MethodologyProfile
  MethodologyProfile?: MethodologyProfile
}

export function pickEpicId(epic: ProjectEpic) {
  return epic.id ?? epic.ID ?? ''
}

export function pickEpicTitle(epic: ProjectEpic) {
  return epic.title ?? epic.Title ?? ''
}

export function pickSprintId(sprint: ProjectSprint) {
  return sprint.id ?? sprint.ID ?? ''
}

export function pickSprintName(sprint: ProjectSprint) {
  return sprint.name ?? sprint.Name ?? ''
}

export function pickDemandId(item: BacklogItem) {
  return item.demandId ?? item.DemandID ?? ''
}

export function useProject() {
  const { apiFetch } = useApiFetch()

  const listAgileApplications = async () => {
    const res = await apiFetch<{ data?: AgileApplication[] }>('/api/project/applications')
    return res?.data ?? []
  }

  const listEpics = async (appId: string) => {
    const res = await apiFetch<{ data?: ProjectEpic[] }>(`/api/project/applications/${appId}/epics`)
    return res?.data ?? []
  }

  const createEpic = async (appId: string, body: { title: string; description?: string; priority?: string }) => {
    return apiFetch(`/api/project/applications/${appId}/epics`, { method: 'POST', body })
  }

  const listSprints = async (appId: string) => {
    const res = await apiFetch<{ data?: ProjectSprint[] }>(`/api/project/applications/${appId}/sprints`)
    return res?.data ?? []
  }

  const createSprint = async (
    appId: string,
    body: { name: string; goal?: string; startDate: string; endDate: string; capacityPoints?: number | null }
  ) => {
    return apiFetch(`/api/project/applications/${appId}/sprints`, { method: 'POST', body })
  }

  const startSprint = async (appId: string, sprintId: string) => {
    return apiFetch(`/api/project/applications/${appId}/sprints/${sprintId}/start`, { method: 'POST' })
  }

  const closeSprint = async (appId: string, sprintId: string) => {
    return apiFetch(`/api/project/applications/${appId}/sprints/${sprintId}/close`, { method: 'POST' })
  }

  const planSprint = async (appId: string, sprintId: string, demandIds: string[]) => {
    return apiFetch(`/api/project/applications/${appId}/sprints/${sprintId}/plan`, {
      method: 'POST',
      body: { demandIds }
    })
  }

  const listBacklog = async (appId: string, backlogOnly = true) => {
    const qs = backlogOnly ? '?backlogOnly=true' : ''
    const res = await apiFetch<{ data?: BacklogItem[] }>(`/api/project/applications/${appId}/backlog${qs}`)
    return res?.data ?? []
  }

  const reorderBacklog = async (appId: string, demandIds: string[]) => {
    return apiFetch(`/api/project/applications/${appId}/backlog/reorder`, {
      method: 'PATCH',
      body: { demandIds }
    })
  }

  const getBurndown = async (appId: string, sprintId: string) => {
    const res = await apiFetch<{ data?: { plannedPoints: number; points: { date: string; remainingPoints: number; idealPoints: number }[] } }>(
      `/api/project/applications/${appId}/sprints/${sprintId}/burndown`
    )
    return res?.data
  }

  const getVelocity = async (appId: string) => {
    const res = await apiFetch<{ data?: { averageVelocity: number; sprints: { sprintName: string; closedPoints: number }[] } }>(
      `/api/project/applications/${appId}/velocity`
    )
    return res?.data
  }

  const updateEpic = async (appId: string, epicId: string, body: { status?: string; title?: string }) => {
    return apiFetch(`/api/project/applications/${appId}/epics/${epicId}`, { method: 'PATCH', body })
  }

  const getKanbanConfig = async (appId: string) => {
    const res = await apiFetch<{ data?: KanbanConfig }>(`/api/project/applications/${appId}/kanban-config`)
    return res?.data
  }

  const saveKanbanConfig = async (appId: string, columns: KanbanColumnConfig[]) => {
    const res = await apiFetch<{ data?: KanbanConfig }>(`/api/project/applications/${appId}/kanban-config`, {
      method: 'PUT',
      body: { columns }
    })
    return res?.data
  }

  const pickKanbanColumns = (cfg: KanbanConfig | undefined | null) => cfg?.columns ?? cfg?.Columns ?? []

  return {
    listAgileApplications,
    listEpics,
    createEpic,
    updateEpic,
    listSprints,
    createSprint,
    startSprint,
    closeSprint,
    planSprint,
    listBacklog,
    reorderBacklog,
    getBurndown,
    getVelocity,
    getKanbanConfig,
    saveKanbanConfig,
    pickKanbanColumns,
    pickEpicId,
    pickEpicTitle,
    pickSprintId,
    pickSprintName,
    pickDemandId
  }
}
