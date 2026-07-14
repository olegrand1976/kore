import { storeToRefs } from 'pinia'
import { useCraStore, type CraLine } from '~/stores/cra'

export function useCra(timesheetId?: Ref<string> | string) {
  const store = useCraStore()
  const { timesheet, loading, saving, error, canEdit, selectedWeeks } = storeToRefs(store)

  const idRef = computed(() => {
    if (typeof timesheetId === 'string') return timesheetId
    if (timesheetId) return timesheetId.value
    return timesheet.value?.id ?? ''
  })

  const load = async (id?: string) => {
    const target = id ?? idRef.value
    if (!target) return
    await store.load(target)
  }

  const saveWeek = (weekNumber: number, lines: CraLine[]) => store.saveWeek(weekNumber, lines)
  const submitWeek = (weekNumber: number) => store.submitWeek(weekNumber)
  const validateFinal = () => store.validateFinal()
  const rejectTimesheet = (reason: string) => store.rejectTimesheet(reason)

  return {
    timesheet,
    loading,
    saving,
    error,
    canEdit,
    selectedWeeks,
    load,
    saveWeek,
    submitWeek,
    validateFinal,
    rejectTimesheet
  }
}
