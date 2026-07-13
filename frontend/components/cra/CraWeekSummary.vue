<template>
  <div class="cra-week-summary">
    <div class="cra-week-summary__main">
      <p class="cra-week-summary__title">{{ title }}</p>
      <p class="cra-week-summary__hours">{{ hoursLabel }}</p>
    </div>
    <div class="cra-week-summary__filters" role="group" :aria-label="$t('cra.origin_filter_label')">
      <button
        v-for="opt in filterOptions"
        :key="opt.value"
        type="button"
        class="cra-week-summary__filter"
        :class="{ 'cra-week-summary__filter--active': originFilter === opt.value }"
        @click="emit('update:originFilter', opt.value)"
      >
        {{ opt.label }}
      </button>
    </div>
    <div class="cra-week-summary__bar" role="progressbar" :aria-valuenow="progress" aria-valuemin="0" aria-valuemax="100">
      <div class="cra-week-summary__fill" :style="{ width: `${progress}%` }" />
    </div>
    <AppBadge v-if="submittedAt" variant="success">{{ $t('cra.week_submitted') }}</AppBadge>
  </div>
</template>

<script setup lang="ts">
import { minutesToHoursLabel } from '~/composables/useWeekCalendar'

export type OriginFilter = 'all' | 'prefill' | 'manual'

const props = defineProps<{
  title: string
  totalMinutes: number
  capacityMinutes?: number
  submittedAt?: string | null
  originFilter?: OriginFilter
}>()

const emit = defineEmits<{
  'update:originFilter': [value: OriginFilter]
}>()

const { t } = useI18n()

const originFilter = computed(() => props.originFilter ?? 'all')

const filterOptions = computed(() => [
  { value: 'all' as const, label: t('cra.origin_filter_all') },
  { value: 'prefill' as const, label: t('cra.origin_filter_prefill') },
  { value: 'manual' as const, label: t('cra.origin_filter_manual') }
])

const hoursLabel = computed(() => {
  const total = minutesToHoursLabel(props.totalMinutes)
  const cap = props.capacityMinutes ? minutesToHoursLabel(props.capacityMinutes) : null
  return cap ? `${total}h / ${cap}h` : `${total}h`
})

const progress = computed(() => {
  if (!props.capacityMinutes || props.capacityMinutes <= 0) return 0
  return Math.min(100, Math.round((props.totalMinutes / props.capacityMinutes) * 100))
})
</script>

<style scoped>
.cra-week-summary {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--kore-space-md);
  padding: var(--kore-space-md);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg-elevated);
}

.cra-week-summary__main {
  flex: 1 1 12rem;
}

.cra-week-summary__title {
  margin: 0;
  font-weight: 600;
}

.cra-week-summary__hours {
  margin: 0.25rem 0 0;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.cra-week-summary__filters {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-xs);
}

.cra-week-summary__filter {
  border: 1px solid var(--kore-border);
  border-radius: 999px;
  padding: 0.2rem 0.6rem;
  font-size: var(--kore-text-small);
  background: transparent;
  color: var(--kore-text-muted);
  cursor: pointer;
}

.cra-week-summary__filter--active {
  border-color: var(--kore-brand-gold);
  color: var(--kore-text);
  background: var(--kore-bg);
}

.cra-week-summary__bar {
  flex: 1 1 8rem;
  height: 0.5rem;
  border-radius: var(--kore-radius-sm);
  background: var(--kore-border);
  overflow: hidden;
}

.cra-week-summary__fill {
  height: 100%;
  background: var(--kore-brand-gold);
  transition: width 0.2s ease;
}
</style>
