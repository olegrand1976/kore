export type WorkflowInstance = {
  id?: string
  ID?: string
  currentState?: string
  CurrentState?: string
  definitionCode?: string
  DefinitionCode?: string
}

export function useWorkflow() {
  const { apiFetch } = useApiFetch()
  const getInstance = async (id: string) => {
    const res = await apiFetch<{ data?: WorkflowInstance }>(`/api/workflow/instances/${id}`)
    return (res?.data ?? res) as WorkflowInstance
  }

  const availableActions = async (id: string) => {
    const res = await apiFetch<{ data?: string[] }>(`/api/workflow/instances/${id}/actions`)
    return res?.data ?? []
  }

  const fire = async (id: string, action: string) => {
    return apiFetch(`/api/workflow/instances/${id}/fire`, { method: 'POST', body: { action } })
  }

  const pickState = (inst: WorkflowInstance) => inst.currentState ?? inst.CurrentState ?? ''

  return { getInstance, availableActions, fire, pickState }
}
