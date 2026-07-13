<script setup lang="ts">
import type { BudgetStatus } from '~/composables/useBudgetDisplay'

const props = defineProps<{
  plannedDays: number
  consumedDays: number
  remainingDays: number
  plannedUO: number
  consumedUO: number
  remainingUO: number
  plannedAmount: number
  consumedAmount: number
  remainingAmount: number
  currency?: string
}>()

const { t } = useI18n()
const { budgetStatus, formatBudgetAmount } = useBudgetDisplay()

type GaugeRow = {
  key: 'days' | 'uo' | 'amount'
  label: string
  planned: number
  consumed: number
  remaining: number
  unit: string
  displayConsumed: string
  displayPlanned: string
  displayRemaining: string
  status: BudgetStatus
}

const rows = computed((): GaugeRow[] => {
  const currency = props.currency ?? 'EUR'
  return [
    {
      key: 'days',
      label: t('budget.triple_days'),
      planned: props.plannedDays,
      consumed: props.consumedDays,
      remaining: props.remainingDays,
      unit: t('budget.unit_days'),
      displayConsumed: String(props.consumedDays),
      displayPlanned: String(props.plannedDays),
      displayRemaining: `${props.remainingDays} ${t('budget.unit_days')}`,
      status: budgetStatus(props.consumedDays, props.plannedDays)
    },
    {
      key: 'uo',
      label: t('budget.triple_uo'),
      planned: props.plannedUO,
      consumed: props.consumedUO,
      remaining: props.remainingUO,
      unit: t('budget.unit_uo'),
      displayConsumed: String(props.consumedUO),
      displayPlanned: String(props.plannedUO),
      displayRemaining: `${props.remainingUO} ${t('budget.unit_uo')}`,
      status: budgetStatus(props.consumedUO, props.plannedUO)
    },
    {
      key: 'amount',
      label: t('budget.triple_amount'),
      planned: props.plannedAmount,
      consumed: props.consumedAmount,
      remaining: props.remainingAmount,
      unit: currency,
      displayConsumed: formatBudgetAmount(props.consumedAmount, currency),
      displayPlanned: formatBudgetAmount(props.plannedAmount, currency),
      displayRemaining: formatBudgetAmount(props.remainingAmount, currency),
      status: budgetStatus(props.consumedAmount, props.plannedAmount)
    }
  ]
})

function pct(consumed: number, planned: number) {
  if (planned <= 0) return consumed > 0 ? 100 : 0
  return Math.min(100, Math.round((consumed / planned) * 100))
}

function fillClass(status: BudgetStatus) {
  switch (status) {
    case 'ok':
      return 'triple-gauge__fill--ok'
    case 'warn':
      return 'triple-gauge__fill--warn'
    case 'overrun':
      return 'triple-gauge__fill--overrun'
    default: {
      const _exhaustive: never = status
      return _exhaustive
    }
  }
}
</script>

<template>
  <div class="triple-gauge">
    <div v-for="row in rows" :key="row.key" class="triple-gauge__row">
      <div class="triple-gauge__head">
        <span>{{ row.label }}</span>
        <span class="triple-gauge__values">
          {{ row.displayConsumed }} / {{ row.displayPlanned }}
          <span v-if="row.key !== 'amount'" class="triple-gauge__unit">{{ row.unit }}</span>
        </span>
      </div>
      <div
        class="triple-gauge__track"
        role="progressbar"
        :aria-valuenow="pct(row.consumed, row.planned)"
        aria-valuemin="0"
        aria-valuemax="100"
      >
        <div
          class="triple-gauge__fill"
          :class="fillClass(row.status)"
          :style="{ width: `${pct(row.consumed, row.planned)}%` }"
        />
      </div>
      <p class="triple-gauge__remaining">{{ $t('budget.remaining', { value: row.displayRemaining }) }}</p>
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

.triple-gauge__unit {
  margin-left: 0.15rem;
}

.triple-gauge__track {
  height: 0.5rem;
  background: var(--kore-bg-subtle);
  border-radius: var(--kore-radius-full);
  overflow: hidden;
}

.triple-gauge__fill {
  height: 100%;
  border-radius: var(--kore-radius-full);
  transition: width 0.2s ease;
}

.triple-gauge__fill--ok {
  background: var(--kore-accent);
}

.triple-gauge__fill--warn {
  background: var(--kore-brand-gold);
}

.triple-gauge__fill--overrun {
  background: var(--kore-danger);
}

.triple-gauge__remaining {
  margin: 0.35rem 0 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}

@media (max-width: 640px) {
  .triple-gauge__head {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
