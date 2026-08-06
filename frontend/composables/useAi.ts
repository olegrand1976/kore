export type AnalysisDraft = {
  functional: string
  technical: string
  risks: string
  testScenario: string
}

export type AnalysisDraftResponse = {
  draft: AnalysisDraft
  requestId: string
}

export type BriefingResponse = {
  text: string
  requestId: string
}

export type ManagerContextResponse = {
  context: string
  requestId: string
}

export type BudgetEstimateResponse = {
  effortDays: number
  effortUO: number
  rationale: string
  requestId: string
}

export type DemandSuggestion = {
  demandId: string
  subject: string
  status: string
}

export type ChatResponse = {
  reply: string
  requestId: string
}

export function useAi() {
  const { apiFetch } = useApiFetch()
  const { extractFetchError } = useApiError()

  const generateAnalysisDraft = async (payload: {
    demandId: string
    subject?: string
    applicationId?: string
  }): Promise<AnalysisDraftResponse> => {
    return apiFetch<AnalysisDraftResponse>('/api/ai/tma/analysis-draft', {
      method: 'POST',
      body: payload
    })
  }

  const classifyDemand = async (subject: string) => {
    return apiFetch<{ category: string; confidence: number; requestId: string }>(
      '/api/ai/tma/classify',
      { method: 'POST', body: { subject } }
    )
  }

  const fetchBriefing = async (params: Record<string, string | number>) => {
    return apiFetch<BriefingResponse>('/api/ai/dashboard/briefing', { query: params })
  }

  const fetchManagerContext = async (leaveRequestId: string) => {
    return apiFetch<ManagerContextResponse>('/api/ai/conges/manager-context', {
      method: 'POST',
      body: { leaveRequestId }
    })
  }

  const estimateBudgetEffort = async (demandId: string, budgetId: string) => {
    return apiFetch<BudgetEstimateResponse>('/api/ai/budget/estimate-effort', {
      method: 'POST',
      body: { demandId, budgetId }
    })
  }

  const suggestBudgetDemands = async (budgetId: string, q = '') => {
    return apiFetch<DemandSuggestion[]>('/api/ai/budget/demand-suggest', {
      query: { budgetId, q }
    })
  }

  const suggestCraPrefill = async (timesheetId: string, weekNumber?: number) => {
    return apiFetch<{ lines: Array<{ day: string; duration: number; comment: string }>; requestId: string }>(
      '/api/ai/cra/prefill-suggest',
      { method: 'POST', body: { timesheetId, weekNumber } }
    )
  }

  const publicChat = async (message: string, sessionId?: string) => {
    return $fetch<ChatResponse>('/api/ai/public/chat', {
      method: 'POST',
      body: { message, sessionId }
    })
  }

  const fetchSettings = async () => {
    return apiFetch<{ data?: Record<string, unknown> }>('/api/ai/settings')
  }

  const enableAI = async (payload: { noticeAccepted: boolean; workersInformed: boolean }) => {
    return apiFetch('/api/ai/settings/enable', { method: 'POST', body: payload })
  }

  const fetchSimilarDemands = async (query: { subject: string; applicationId?: string; limit?: number }) => {
    return apiFetch('/api/ai/tma/similar', { query })
  }

  const fetchCraAnomalies = async (timesheetId: string) => {
    return apiFetch('/api/ai/cra/anomalies', { query: { timesheetId } })
  }

  const explainWorkflow = async (instanceId: string) => {
    return apiFetch('/api/ai/workflow/explain', { query: { instanceId } })
  }

  const explainRequest = async (requestId: string) => {
    return apiFetch(`/api/ai/explain/${requestId}`)
  }

  return {
    extractFetchError,
    generateAnalysisDraft,
    classifyDemand,
    fetchBriefing,
    fetchManagerContext,
    estimateBudgetEffort,
    suggestBudgetDemands,
    suggestCraPrefill,
    publicChat,
    fetchSettings,
    enableAI,
    fetchSimilarDemands,
    fetchCraAnomalies,
    explainWorkflow,
    explainRequest
  }
}
