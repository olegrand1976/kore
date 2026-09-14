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
        :work-ref-options="workRefOptions"
        :work-ref-label-for="workRefLabelFor"
        :dirty-row-keys="dirtyKeys"
        :saving-row-key="savingRowKey"
        :saving="saving"
        @update:rows="(rows) => setDayRows(day, rows)"
        @add-activity="openAddModal"
        @save-line="saveLine"
      />
    </div>

    <div class="week-matrix__actions">
      <AppButton variant="primary" size="sm" :disabled="disabled || saving" @click="emitSave">
        {{ $t('cra.save_week') }}
      </AppButton>
      <AppButton variant="secondary" size="sm" :disabled="disabled || saving" @click="startSubmit">
        {{ $t('cra.submit_week') }}
      </AppButton>
    </div>

    <CraAddActivityModal
      v-model:open="addModalOpen"
      :missions="missions"
      :task-types="taskTypes"
      @add="onAddActivity"
    />

    <AppModal
      v-model:open="submitDialogOpen"
      width="sm"
      :title-id="submitDialogTitleId"
      :aria-label="submitDialogTitle"
      :close-label="$t('common.close')"
    >
      <div class="week-matrix__dialog">
        <h2 :id="submitDialogTitleId" class="week-matrix__dialog-title">{{ submitDialogTitle }}</h2>
        <p class="week-matrix__dialog-body">{{ submitDialogBody }}</p>
        <div class="week-matrix__dialog-actions">
          <AppButton
            v-if="submitDialogCanConfirm"
            variant="ghost"
            size="sm"
            type="button"
            @click="closeSubmitDialog"
          >
            {{ $t('common.cancel') }}
          </AppButton>
          <AppButton
            :variant="submitDialogCanConfirm ? 'primary' : 'secondary'"
            size="sm"
            type="button"
            @click="onSubmitDialogPrimary"
          >
            {{ submitDialogPrimaryLabel }}
          </AppButton>
        </div>
      </div>
    </AppModal>
  </div>
</template>

<script setup lang="ts">
import type { CraLine, CraWeek } from '~/stores/cra'
import type { MissionSummary } from '~/composables/useCraSourceLabels'
import type { ActivityRow } from '~/composables/useWeekRows'
import { hoursToMinutes } from '~/composables/useWeekCalendar'
import { useCraSourceLabels } from '~/composables/useCraSourceLabels'
import { newRowKey, useWeekRows } from '~/composables/useWeekRows'
import {
  collectIncompleteWorkingDays,
  countWorkingDays
} from '~/utils/craCalendar'
import { dirtyRowKeys, unlockHolidayPrefillRows } from '~/utils/craDayState'
import { resolveWeekCapacityMinutes } from '~/utils/craWeekCapacity'

import type { CraWorkRefOption } from '~/composables/useCraWorkRefs'

type SubmitDialog =
  | { kind: 'incomplete-warn'; n: number }
  | { kind: 'incomplete-block'; n: number }
  | { kind: 'below-planned-warn'; actual: string; target: string }
  | { kind: 'below-planned-block'; actual: string; target: string }

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
  /** Optional mission planned week target (minutes); overrides org day×days capacity. */
  plannedWeekMinutes?: number | null
  missions?: MissionSummary[]
  taskTypes?: string[]
  workRefOptions?: CraWorkRefOption[]
  workRefLabelFor?: (type: string, id: string) => string
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
const taskTypes = computed(() => props.taskTypes ?? ['manual', 'interne', 'formation', 'mission'])

const { labelFor, iconFor } = useCraSourceLabels(missionsRef)
const { weekDays, rowsByDay, toSaveLines } = useWeekRows(
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
const submitDialog = ref<SubmitDialog | null>(null)
const submitDialogTitleId = 'cra-week-submit-dialog-title'

const dayCapacityMinutes = computed(() => props.dayCapacityMinutes ?? 8 * 60)
const weekCapacityMinutes = computed(() =>
  resolveWeekCapacityMinutes({
    weekDayCount: countWorkingDays(weekDays.value),
    dayCapacityMinutes: dayCapacityMinutes.value,
    plannedWeekMinutes: props.plannedWeekMinutes
  })
)

const matchesOriginFilter = (row: ActivityRow) => {
  if (originFilter.value === 'all') return true
  if (originFilter.value === 'prefill') return row.origin === 'prefill'
  return row.origin !== 'prefill'
}

const summaryTitle = computed(() => props.weekLabel ?? t('cra.week_n', { n: props.weekNumber }))

/** Display total — respects origin filter (summary / sticky pill). */
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

/** Full week total for submit checks — never filtered by origin. */
const submitTotalMinutes = computed(() => {
  let total = 0
  for (const rows of editableRows.value.values()) {
    for (const row of rows) {
      total += hoursToMinutes(row.hours)
    }
  }
  return total
})

// `rowsByDay` vient du store (dernière réponse serveur), `editableRows` porte le
// brouillon : leur écart donne les lignes à enregistrer.
const dirtyKeys = computed(() => dirtyRowKeys(editableRows.value, rowsByDay.value, hoursToMinutes))

const savingRowKey = ref('')

watch(rowsByDay, (map) => {
  const next = new Map<string, ActivityRow[]>()
  for (const [day, rows] of map) {
    next.set(day, rows.map((r) => ({ ...r })))
  }
  editableRows.value = next
  savingRowKey.value = ''
}, { immediate: true, deep: true })

// Un enregistrement en échec ne modifie pas le store, donc le watcher ci-dessus
// ne se déclenche pas : sans ce reset la ligne resterait bloquée sur le sablier,
// bouton désactivé, jusqu'au prochain succès ou rechargement.
watch(() => props.saving, (saving) => {
  if (!saving) savingRowKey.value = ''
})

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
  const rows = unlockHolidayPrefillRows(editableRows.value.get(day) ?? [])
  rows.push({
    key: newRowKey(),
    sourceType,
    sourceId,
    day,
    hours: '',
    comment: '',
    origin: 'manual',
    billable: true
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

/**
 * Enregistrer « une ligne » reste un PUT de la semaine entière — l'API remplace
 * intégralement `week.Lines`, envoyer la seule ligne éditée supprimerait les autres.
 * La granularité est donc UX : `savingRowKey` ne sert qu'à situer l'indicateur.
 */
const saveLine = (rowKey: string) => {
  savingRowKey.value = rowKey
  emitSave()
}

const incompleteDays = computed(() =>
  collectIncompleteWorkingDays(weekDays.value, editableRows.value, hoursToMinutes)
)

const submitDialogOpen = computed({
  get: () => submitDialog.value != null,
  set: (open: boolean) => {
    if (!open) submitDialog.value = null
  }
})

const submitDialogCanConfirm = computed(() => {
  const kind = submitDialog.value?.kind
  return kind === 'incomplete-warn' || kind === 'below-planned-warn'
})

const submitDialogTitle = computed(() => {
  const dialog = submitDialog.value
  if (!dialog) return ''
  switch (dialog.kind) {
    case 'incomplete-warn':
    case 'below-planned-warn':
      return t('cra.submit_confirm_title')
    case 'incomplete-block':
    case 'below-planned-block':
      return t('cra.submit_blocked_title')
    default: {
      const _exhaustive: never = dialog
      return _exhaustive
    }
  }
})

const submitDialogBody = computed(() => {
  const dialog = submitDialog.value
  if (!dialog) return ''
  switch (dialog.kind) {
    case 'incomplete-warn':
      return t('cra.submit_week_incomplete', { n: dialog.n })
    case 'incomplete-block':
      return t('cra.submit_week_blocked', { n: dialog.n })
    case 'below-planned-warn':
      return t('cra.submit_week_below_planned', {
        actual: dialog.actual,
        target: dialog.target
      })
    case 'below-planned-block':
      return t('cra.submit_week_below_planned_blocked', {
        actual: dialog.actual,
        target: dialog.target
      })
    default: {
      const _exhaustive: never = dialog
      return _exhaustive
    }
  }
})

const submitDialogPrimaryLabel = computed(() =>
  submitDialogCanConfirm.value ? t('cra.submit_anyway') : t('common.close')
)

const closeSubmitDialog = () => {
  submitDialog.value = null
}

const formatHours = (minutes: number) =>
  (minutes / 60).toLocaleString(undefined, { maximumFractionDigits: 2 })

const continueAfterIncomplete = () => {
  const planned = props.plannedWeekMinutes
  const policy = props.weekSubmitPolicy ?? 'warn'
  if (
    planned != null &&
    planned > 0 &&
    submitTotalMinutes.value < planned &&
    policy !== 'none'
  ) {
    const target = formatHours(planned)
    const actual = formatHours(submitTotalMinutes.value)
    if (policy === 'block') {
      submitDialog.value = { kind: 'below-planned-block', actual, target }
      return
    }
    submitDialog.value = { kind: 'below-planned-warn', actual, target }
    return
  }
  submitDialog.value = null
  emit('submit')
}

const startSubmit = () => {
  const missing = incompleteDays.value
  const policy = props.weekSubmitPolicy ?? 'warn'
  if (missing.length > 0) {
    if (policy === 'block') {
      submitDialog.value = { kind: 'incomplete-block', n: missing.length }
      return
    }
    if (policy === 'warn') {
      submitDialog.value = { kind: 'incomplete-warn', n: missing.length }
      return
    }
  }
  continueAfterIncomplete()
}

const onSubmitDialogPrimary = () => {
  const dialog = submitDialog.value
  if (!dialog) return
  switch (dialog.kind) {
    case 'incomplete-block':
    case 'below-planned-block':
      closeSubmitDialog()
      return
    case 'incomplete-warn':
      continueAfterIncomplete()
      return
    case 'below-planned-warn':
      closeSubmitDialog()
      emit('submit')
      return
    default: {
      const _exhaustive: never = dialog
      return _exhaustive
    }
  }
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

.week-matrix__dialog {
  display: grid;
  gap: var(--kore-space-md);
}

.week-matrix__dialog-title {
  margin: 0;
  font-size: var(--kore-text-h3);
  color: var(--kore-text);
}

.week-matrix__dialog-body {
  margin: 0;
  color: var(--kore-text);
}

.week-matrix__dialog-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
  justify-content: flex-end;
}

@media (max-width: 768px) {
  .week-matrix__actions :deep(.app-btn) {
    flex: 1 1 100%;
  }

  .week-matrix__dialog-actions {
    flex-direction: column-reverse;
  }

  .week-matrix__dialog-actions :deep(.app-btn) {
    width: 100%;
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
