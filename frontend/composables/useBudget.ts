export type BudgetTriple = {
  days?: number
  Days?: number
  uo?: number
  UO?: number
  amount?: number
  Amount?: number
}

export type BudgetItem = {
  id?: string
  ID?: string
  type?: string
  Type?: string
  applicationId?: string
  ApplicationID?: string
  planned?: BudgetTriple
  Planned?: BudgetTriple
  consumed?: BudgetTriple
  Consumed?: BudgetTriple
  remaining?: BudgetTriple
  Remaining?: BudgetTriple
  currency?: string
  Currency?: string
}

function tripleValue(triple: BudgetTriple | undefined, key: 'days' | 'uo' | 'amount') {
  if (!triple) return 0
  if (key === 'days') return triple.days ?? triple.Days ?? 0
  if (key === 'uo') return triple.uo ?? triple.UO ?? 0
  return triple.amount ?? triple.Amount ?? 0
}

export function useBudget() {
  const list = async () => {
    const res = await $fetch<{ data?: BudgetItem[] }>('/api/budget/budgets')
    return res?.data ?? []
  }

  const get = async (id: string) => {
    const res = await $fetch<{ data?: BudgetItem }>(`/api/budget/budgets/${id}`)
    return res?.data ?? res
  }

  const addEstimate = async (budgetId: string, payload: { demandId: string; effortDays: number; effortUO: number }) => {
    return $fetch(`/api/budget/budgets/${budgetId}/estimates`, { method: 'POST', body: payload })
  }

  const addQuote = async (
    budgetId: string,
    payload: { demandId: string; amount: number; effortDays: number; effortUO: number; supersedesEstimateId?: string }
  ) => {
    return $fetch(`/api/budget/budgets/${budgetId}/quotes`, { method: 'POST', body: payload })
  }

  const recompute = async (budgetId: string, period: { start: string; end: string }) => {
    return $fetch(`/api/budget/budgets/${budgetId}/recompute`, { method: 'POST', body: period })
  }

  const pickId = (b: BudgetItem) => b.id ?? b.ID ?? ''

  return { list, get, addEstimate, addQuote, recompute, tripleValue, pickId }
}
