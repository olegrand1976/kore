<template>
  <section
    class="day-block"
    :class="{
      'day-block--open': open,
      'day-block--over': overCapacity,
      'day-block--absence': isAbsenceDay,
      [absenceClass]: isAbsenceDay
    }"
  >
    <button type="button" class="day-block__toggle" :aria-expanded="showBody" @click="toggle">
      <span class="day-block__date">
        <AppIcon v-if="isAbsenceDay" :name="headerIcon" class="day-block__date-icon" />
        {{ dayLabel }}
      </span>
      <span v-if="isAbsenceDay" class="day-block__absence-label">{{ headerAbsenceLabel }}</span>
      <AppBadge v-if="isAbsenceDay" variant="info">{{ $t('cra.day_non_worked') }}</AppBadge>
      <span class="day-block__total" :class="{ 'day-block__total--muted': isAbsenceDay }">{{ totalLabel }}</span>
      <AppIcon :name="open ? 'expand_less' : 'expand_more'" />
    </button>
    <div v-show="showBody" class="day-block__body">
      <template v-for="(row, idx) in localRows" :key="row.key">
        <ActivityLineRow
          v-show="isRowVisible(row)"
          :input-id="`line-${day}-${idx}`"
          :label="labelFor(row.sourceType, row.sourceId)"
          :icon="iconFor(row.sourceType)"
          :source-type="row.sourceType"
          :hours="row.hours"
          :comment="row.comment"
          :origin="row.origin"
          :billable="row.billable"
          :absence="isAbsenceSourceType(row.sourceType)"
          :disabled="disabled"
          :can-remove="localRows.length > 1"
          @update:hours="(v) => updateRow(idx, 'hours', v)"
          @update:comment="(v) => updateRow(idx, 'comment', v)"
          @update:billable="(v) => updateRowBillable(idx, v)"
          @remove="removeRow(idx)"
        />
      </template>
      <AppButton variant="ghost" size="sm" :disabled="disabled" @click="$emit('add-activity', day)">
        <AppIcon name="add" /> {{ $t('cra.add_activity') }}
      </AppButton>
    </div>
  </section>
</template>

<script setup lang="ts">
import type { ActivityRow } from '~/composables/useWeekRows'
import { hoursToMinutes, minutesToHoursLabel } from '~/composables/useWeekCalendar'
import { absenceDayClass, isAbsenceSourceType } from '~/utils/craAbsence'

import type { OriginFilter } from '~/components/cra/CraWeekSummary.vue'

const props = defineProps<{
  day: string
  rows: ActivityRow[]
  capacityMinutes: number
  disabled?: boolean
  defaultOpen?: boolean
  originFilter?: OriginFilter
  labelFor: (sourceType: string, sourceId: string) => string
  iconFor: (sourceType: string) => string
}>()

const emit = defineEmits<{
  'update:rows': [rows: ActivityRow[]]
  'add-activity': [day: string]
}>()

const { t, locale } = useI18n()
const open = ref(props.defaultOpen ?? false)
const isMobile = ref(false)
const localRows = ref<ActivityRow[]>([])

const showBody = computed(() => !isMobile.value || open.value)

const toggle = () => {
  if (isMobile.value) open.value = !open.value
}

onMounted(() => {
  const mq = window.matchMedia('(max-width: 768px)')
  const update = () => { isMobile.value = mq.matches }
  update()
  mq.addEventListener('change', update)
  onUnmounted(() => mq.removeEventListener('change', update))
})

watch(
  () => props.rows,
  (rows) => {
    localRows.value = rows.map((r) => ({ ...r }))
  },
  { immediate: true, deep: true }
)

watch(localRows, (rows) => emit('update:rows', rows.map((r) => ({ ...r }))), { deep: true })

const dayLabel = computed(() =>
  new Date(`${props.day}T12:00:00`).toLocaleDateString(locale.value === 'en' ? 'en-US' : 'fr-FR', {
    weekday: 'short',
    day: 'numeric',
    month: 'short'
  })
)

const isRowVisible = (row: ActivityRow) => {
  const filter = props.originFilter ?? 'all'
  if (filter === 'all') return true
  if (filter === 'prefill') return row.origin === 'prefill'
  return row.origin !== 'prefill'
}

const visibleRows = computed(() => localRows.value.filter((row) => isRowVisible(row)))

const absenceRows = computed(() => visibleRows.value.filter((row) => isAbsenceSourceType(row.sourceType)))

const hasWorkedHours = computed(() =>
  visibleRows.value.some((row) => !isAbsenceSourceType(row.sourceType) && hoursToMinutes(row.hours) > 0)
)

const isAbsenceDay = computed(() => absenceRows.value.length > 0 && !hasWorkedHours.value)

const primaryAbsenceRow = computed(() => absenceRows.value[0])

const absenceClass = computed(() =>
  primaryAbsenceRow.value ? absenceDayClass(primaryAbsenceRow.value.sourceType) : ''
)

const headerIcon = computed(() =>
  primaryAbsenceRow.value ? props.iconFor(primaryAbsenceRow.value.sourceType) : 'event_busy'
)

const headerAbsenceLabel = computed(() => {
  const row = primaryAbsenceRow.value
  if (!row) return ''
  return props.labelFor(row.sourceType, row.sourceId)
})

const totalMinutes = computed(() =>
  visibleRows.value.reduce((sum, row) => sum + hoursToMinutes(row.hours), 0)
)

const totalLabel = computed(() => {
  if (isAbsenceDay.value) {
    return t('cra.full_day_absence')
  }
  return `${minutesToHoursLabel(totalMinutes.value)}h / ${minutesToHoursLabel(props.capacityMinutes)}h`
})

const overCapacity = computed(() => !isAbsenceDay.value && totalMinutes.value > props.capacityMinutes)

const updateRow = (idx: number, field: 'hours' | 'comment', value: string) => {
  const row = localRows.value[idx]
  if (!row) return
  localRows.value[idx] = { ...row, [field]: value, origin: 'manual' }
}

const updateRowBillable = (idx: number, value: boolean) => {
  const row = localRows.value[idx]
  if (!row) return
  localRows.value[idx] = { ...row, billable: value, origin: 'manual' }
}

const removeRow = (idx: number) => {
  localRows.value.splice(idx, 1)
}
</script>

<style scoped>
.day-block {
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg);
}

.day-block--absence {
  background: var(--kore-bg-subtle);
  border-color: var(--kore-border);
  border-left-width: 3px;
  border-left-style: solid;
}

.day-block--absence-holiday {
  border-left-color: var(--kore-brand-gold);
}

.day-block--absence-leave {
  border-left-color: var(--kore-brand-blue);
}

.day-block--absence-other {
  border-left-color: var(--kore-text-muted);
}

.day-block--over {
  border-color: var(--kore-error);
}

.day-block__toggle {
  width: 100%;
  display: flex;
  align-items: center;
  gap: var(--kore-space-sm);
  padding: var(--kore-space-md);
  border: none;
  background: transparent;
  cursor: pointer;
  text-align: left;
  flex-wrap: wrap;
}

.day-block__date {
  display: inline-flex;
  align-items: center;
  gap: var(--kore-space-xs);
  font-weight: 600;
  flex: 1;
  min-width: 6rem;
}

.day-block__date-icon {
  color: var(--kore-brand-gold);
  flex-shrink: 0;
}

.day-block--absence-leave .day-block__date-icon {
  color: var(--kore-brand-blue);
}

.day-block__absence-label {
  font-size: var(--kore-text-small);
  font-weight: 600;
  color: var(--kore-text);
}

.day-block__total {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.day-block__total--muted {
  font-style: italic;
}

.day-block--open .day-block__toggle,
.day-block--over .day-block__total {
  color: inherit;
}

.day-block--over .day-block__total {
  color: var(--kore-error);
}

.day-block__body {
  padding: 0 var(--kore-space-md) var(--kore-space-md);
  display: grid;
  gap: var(--kore-space-sm);
}

@media (min-width: 900px) {
  .day-block__toggle {
    cursor: default;
    pointer-events: none;
  }

  .day-block__body {
    display: block;
  }
}
</style>
