<template>
  <div class="cra-month-preview">
    <div class="cra-month-preview__kpi">
      <p class="cra-month-preview__label">{{ $t('cra.month_total') }}</p>
      <p class="cra-month-preview__value">{{ hoursLabel }}</p>
    </div>
    <div class="cra-month-preview__kpi">
      <p class="cra-month-preview__label">{{ $t('cra.month_weeks') }}</p>
      <p class="cra-month-preview__value">{{ weeksSubmitted }} / {{ weeksTotal }}</p>
    </div>
    <div class="cra-month-preview__kpi">
      <p class="cra-month-preview__label">{{ $t('cra.month_prefill') }}</p>
      <p class="cra-month-preview__value">{{ prefillRatio }}%</p>
    </div>
    <div class="cra-month-preview__bar" role="progressbar" :aria-valuenow="progress" aria-valuemin="0" aria-valuemax="100">
      <div class="cra-month-preview__fill" :style="{ width: `${progress}%` }" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { minutesToHoursLabel } from '~/composables/useWeekCalendar'
import { safeMinutes } from '~/utils/craDuration'

const props = defineProps<{
  totalMinutes: number
  capacityMinutes: number
  weeksSubmitted: number
  weeksTotal: number
  prefillRatio: number
  progress: number
}>()

const hoursLabel = computed(() => {
  const total = minutesToHoursLabel(safeMinutes(props.totalMinutes))
  const capMinutes = safeMinutes(props.capacityMinutes)
  const cap = capMinutes > 0 ? minutesToHoursLabel(capMinutes) : null
  return cap ? `${total}h / ${cap}h` : `${total}h`
})
</script>

<style scoped>
.cra-month-preview {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--kore-space-md);
  align-items: end;
}

.cra-month-preview__kpi {
  min-width: 0;
}

.cra-month-preview__label {
  margin: 0;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.cra-month-preview__value {
  margin: 0.25rem 0 0;
  font-weight: 700;
  font-size: 1.125rem;
}

.cra-month-preview__bar {
  grid-column: 1 / -1;
  height: 0.5rem;
  border-radius: var(--kore-radius-sm);
  background: var(--kore-border);
  overflow: hidden;
}

.cra-month-preview__fill {
  height: 100%;
  background: var(--kore-brand-gold);
  transition: width 0.2s ease;
}

@media (max-width: 768px) {
  .cra-month-preview {
    grid-template-columns: 1fr 1fr;
  }

  .cra-month-preview__bar {
    grid-column: 1 / -1;
  }
}
</style>
