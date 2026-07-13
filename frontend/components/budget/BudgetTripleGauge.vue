<script setup lang="ts">
const props = defineProps<{
  plannedDays: number
  consumedDays: number
  plannedUO: number
  consumedUO: number
  plannedAmount: number
  consumedAmount: number
  currency?: string
}>()

const { t } = useI18n()

const rows = computed(() => [
  {
    key: 'days',
    label: t('budget.triple_days'),
    planned: props.plannedDays,
    consumed: props.consumedDays,
    unit: t('budget.unit_days')
  },
  {
    key: 'uo',
    label: t('budget.triple_uo'),
    planned: props.plannedUO,
    consumed: props.consumedUO,
    unit: t('budget.unit_uo')
  },
  {
    key: 'amount',
    label: t('budget.triple_amount'),
    planned: props.plannedAmount,
    consumed: props.consumedAmount,
    unit: props.currency ?? 'EUR'
  }
])

function pct(consumed: number, planned: number) {
  if (planned <= 0) return consumed > 0 ? 100 : 0
  return Math.min(100, Math.round((consumed / planned) * 100))
}
</script>

<template>
  <div class="triple-gauge">
    <div v-for="row in rows" :key="row.key" class="triple-gauge__row">
      <div class="triple-gauge__head">
        <span>{{ row.label }}</span>
        <span class="triple-gauge__values">
          {{ row.consumed }} / {{ row.planned }} {{ row.unit }}
        </span>
      </div>
      <div class="triple-gauge__track" role="progressbar" :aria-valuenow="pct(row.consumed, row.planned)" aria-valuemin="0" aria-valuemax="100">
        <div class="triple-gauge__fill" :style="{ width: `${pct(row.consumed, row.planned)}%` }" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.triple-gauge {
  display: grid;
  gap: var(--kore-space-md);
}

.triple-gauge__head {
  display: flex;
  justify-content: space-between;
  gap: var(--kore-space-sm);
  font-size: var(--kore-text-small);
  margin-bottom: 0.35rem;
}

.triple-gauge__values {
  color: var(--kore-text-muted);
  white-space: nowrap;
}

.triple-gauge__track {
  height: 0.5rem;
  background: var(--kore-bg-subtle);
  border-radius: var(--kore-radius-full);
  overflow: hidden;
}

.triple-gauge__fill {
  height: 100%;
  background: var(--kore-accent);
  border-radius: var(--kore-radius-full);
  transition: width 0.2s ease;
}

@media (max-width: 640px) {
  .triple-gauge__head {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
