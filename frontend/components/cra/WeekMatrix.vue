<template>
  <div class="week-matrix">
    <CraWeekSummary
      :title="summaryTitle"
      :total-minutes="weekTotalMinutes"
      :capacity-minutes="weekCapacityMinutes"
      :submitted-at="week?.submittedAt"
      :origin-filter="originFilter"
      @update:origin-filter="originFilter = $event"
    />

    <CraStickyTotalsPill
      :total-minutes="weekTotalMinutes"
      :capacity-minutes="weekCapacityMinutes"
      :visible="showStickyPill"
    />

    <div class="week-matrix__days">
      <DayActivityBlock
        v-for="(day, idx) in weekDays"
        :key="day"
        :day="day"
        :rows="editableRows.get(day) ?? []"
        :capacity-minutes="dayCapacityMinutes"
        :origin-filter="originFilter"
        :disabled="disabled"
        :default-open="isMobile ? idx === 0 : true"
        :label-for="labelFor"
        :icon-for="iconFor"
        @update:rows="(rows) => setDayRows(day, rows)"
        @add-activity="openAddModal"
      />
    </div>

    <div class="week-matrix__actions">
      <AppButton variant="primary" size="sm" :disabled="disabled || saving" @click="emitSave">
        {{ $t('cra.save_week') }}
      </AppButton>
      <AppButton variant="secondary" size="sm" :disabled="disabled || saving" @click="emitSubmit">
        {{ $t('cra.submit_week') }}
      </AppButton>
    </div>

    <CraAddActivityModal
      v-model:open="addModalOpen"
      :missions="missions"
      @add="onAddActivity"
    />
  </div>
</template>

<script setup lang="ts">
import type { CraLine, CraWeek } from '~/stores/cra'
import type { MissionSummary } from '~/composables/useCraSourceLabels'
import type { ActivityRow } from '~/composables/useWeekRows'
import { hoursToMinutes } from '~/composables/useWeekCalendar'
import { useCraSourceLabels } from '~/composables/useCraSourceLabels'
import { useWeekRows } from '~/composables/useWeekRows'

const props = defineProps<{
  weekNumber: number
  week?: CraWeek
  month: string
  weekStartDay: number
  dayCapacityMinutes?: number
  weekSubmitPolicy?: 'block' | 'warn' | 'none'
  weekLabel?: string
  disabled?: boolean
  saving?: boolean
  missions?: MissionSummary[]
}>()

const emit = defineEmits<{
  save: [lines: CraLine[]]
  submit: []
}>()

const { t } = useI18n()
const weekRef = toRef(props, 'week')
const weekNumberRef = toRef(props, 'weekNumber')
const monthRef = toRef(props, 'month')
const weekStartDayRef = toRef(props, 'weekStartDay')
const missionsRef = computed(() => props.missions ?? [])

const { labelFor, iconFor } = useCraSourceLabels(missionsRef)
const { weekDays, rowsByDay, toSaveLines, buildKey } = useWeekRows(
  weekRef,
  weekNumberRef,
  monthRef,
  weekStartDayRef
)

const editableRows = ref(new Map<string, ActivityRow[]>())
const addModalOpen = ref(false)
const addTargetDay = ref('')
const isMobile = ref(false)
const showStickyPill = ref(false)
const originFilter = ref<'all' | 'prefill' | 'manual'>('all')

const dayCapacityMinutes = computed(() => props.dayCapacityMinutes ?? 8 * 60)
const weekCapacityMinutes = computed(() => weekDays.value.length * dayCapacityMinutes.value)

const matchesOriginFilter = (row: ActivityRow) => {
  if (originFilter.value === 'all') return true
  if (originFilter.value === 'prefill') return row.origin === 'prefill'
  return row.origin !== 'prefill'
}

const summaryTitle = computed(() => props.weekLabel ?? t('cra.week_n', { n: props.weekNumber }))

const weekTotalMinutes = computed(() => {
  let total = 0
  for (const rows of editableRows.value.values()) {
    for (const row of rows) {
      if (!matchesOriginFilter(row)) continue
      total += hoursToMinutes(row.hours)
    }
  }
  return total
})

watch(rowsByDay, (map) => {
  const next = new Map<string, ActivityRow[]>()
  for (const [day, rows] of map) {
    next.set(day, rows.map((r) => ({ ...r })))
  }
  editableRows.value = next
}, { immediate: true, deep: true })

onMounted(() => {
  const mq = window.matchMedia('(max-width: 768px)')
  const update = () => { isMobile.value = mq.matches }
  update()
  mq.addEventListener('change', update)
  onUnmounted(() => mq.removeEventListener('change', update))

  const onScroll = () => {
    showStickyPill.value = window.scrollY > 120 && isMobile.value
  }
  window.addEventListener('scroll', onScroll, { passive: true })
  onUnmounted(() => window.removeEventListener('scroll', onScroll))
})

const setDayRows = (day: string, rows: ActivityRow[]) => {
  const next = new Map(editableRows.value)
  next.set(day, rows)
  editableRows.value = next
}

const openAddModal = (day: string) => {
  addTargetDay.value = day
  addModalOpen.value = true
}

const onAddActivity = ({ sourceType, sourceId }: { sourceType: string; sourceId: string }) => {
  const day = addTargetDay.value
  if (!day) return
  const rows = [...(editableRows.value.get(day) ?? [])]
  const key = buildKey(sourceType, sourceId, day)
  if (rows.some((r) => r.key === key)) return
  rows.push({
    key,
    sourceType,
    sourceId,
    day,
    hours: '',
    comment: '',
    origin: 'manual'
  })
  setDayRows(day, rows)
}

const emitSave = () => {
  const allRows: ActivityRow[] = []
  for (const rows of editableRows.value.values()) {
    allRows.push(...rows)
  }
  emit('save', toSaveLines(allRows))
}

const incompleteDays = computed(() => {
  const missing: string[] = []
  for (const day of weekDays.value) {
    const rows = editableRows.value.get(day) ?? []
    const total = rows.reduce((sum, row) => sum + hoursToMinutes(row.hours), 0)
    if (total <= 0) missing.push(day)
  }
  return missing
})

const emitSubmit = () => {
  const missing = incompleteDays.value
  const policy = props.weekSubmitPolicy ?? 'warn'
  if (missing.length > 0) {
    if (policy === 'block') {
      window.alert(t('cra.submit_week_blocked', { n: missing.length }))
      return
    }
    if (policy === 'warn') {
      const ok = window.confirm(t('cra.submit_week_incomplete', { n: missing.length }))
      if (!ok) return
    }
  }
  emit('submit')
}
</script>

<style scoped>
.week-matrix {
  display: grid;
  gap: var(--kore-space-lg);
}

.week-matrix__days {
  display: grid;
  gap: var(--kore-space-md);
}

.week-matrix__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
  position: sticky;
  bottom: calc(4.5rem + env(safe-area-inset-bottom, 0px));
  z-index: 2;
  padding: var(--kore-space-sm) 0;
  background: linear-gradient(to top, var(--kore-bg) 70%, transparent);
}

@media (max-width: 768px) {
  .week-matrix__actions :deep(.app-btn) {
    flex: 1 1 100%;
  }
}

@media (min-width: 900px) {
  .week-matrix__actions {
    position: static;
    background: none;
    padding: 0;
  }
}
</style>
