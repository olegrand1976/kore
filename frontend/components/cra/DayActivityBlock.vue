<template>
  <section class="day-block" :class="{ 'day-block--open': open, 'day-block--over': overCapacity }">
    <button type="button" class="day-block__toggle" :aria-expanded="showBody" @click="toggle">
      <span class="day-block__date">{{ dayLabel }}</span>
      <span class="day-block__total">{{ totalLabel }}</span>
      <AppIcon :name="open ? 'expand_less' : 'expand_more'" />
    </button>
    <div v-show="showBody" class="day-block__body">
      <template v-for="(row, idx) in localRows" :key="row.key">
        <ActivityLineRow
          v-show="isRowVisible(row)"
          :input-id="`line-${day}-${idx}`"
        :label="labelFor(row.sourceType, row.sourceId)"
        :icon="iconFor(row.sourceType)"
        :hours="row.hours"
        :comment="row.comment"
        :origin="row.origin"
        :disabled="disabled"
        :can-remove="localRows.length > 1"
        @update:hours="(v) => updateRow(idx, 'hours', v)"
        @update:comment="(v) => updateRow(idx, 'comment', v)"
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

const { locale } = useI18n()
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

const totalMinutes = computed(() =>
  localRows.value
    .filter((row) => isRowVisible(row))
    .reduce((sum, row) => sum + hoursToMinutes(row.hours), 0)
)
const totalLabel = computed(() => `${minutesToHoursLabel(totalMinutes.value)}h / ${minutesToHoursLabel(props.capacityMinutes)}h`)
const overCapacity = computed(() => totalMinutes.value > props.capacityMinutes)

const updateRow = (idx: number, field: 'hours' | 'comment', value: string) => {
  const row = localRows.value[idx]
  if (!row) return
  localRows.value[idx] = { ...row, [field]: value, origin: 'manual' }
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
}

.day-block__date {
  font-weight: 600;
  flex: 1;
}

.day-block__total {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
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
