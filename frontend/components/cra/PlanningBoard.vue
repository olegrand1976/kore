<template>
  <div class="planning">
    <div v-if="!rows.length" class="planning__empty">{{ emptyTitle }}</div>
    <div v-else class="planning__scroll">
      <table class="planning__table">
        <thead>
          <tr>
            <th scope="col">{{ userHeader }}</th>
            <th v-for="day in days" :key="day" scope="col">{{ formatDay(day) }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in rows" :key="row.userId">
            <th scope="row" class="planning__user">{{ row.userName || row.userId.slice(0, 8) }}</th>
            <td v-for="day in days" :key="`${row.userId}-${day}`" class="planning__cell">
              <span v-if="slotFor(row, day)" class="planning__slot" :title="slotFor(row, day)?.label">
                {{ slotFor(row, day)?.hours ? `${slotFor(row, day)?.hours}h` : '—' }}
              </span>
              <span v-else class="planning__empty-cell">—</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { PlanningRow } from '~/composables/useReporting'

const props = defineProps<{
  rows: PlanningRow[]
  days: string[]
  userHeader: string
  emptyTitle: string
}>()

const { locale } = useI18n()

const formatDay = (day: string) => {
  const d = new Date(`${day}T12:00:00Z`)
  return d.toLocaleDateString(locale.value === 'en' ? 'en-US' : 'fr-FR', {
    weekday: 'short',
    day: '2-digit'
  })
}

const slotFor = (row: PlanningRow, day: string) =>
  row.slots.find((slot) => slot.date.slice(0, 10) === day.slice(0, 10))
</script>

<style scoped>
.planning__scroll {
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
}

.planning__table {
  width: 100%;
  min-width: 36rem;
  border-collapse: collapse;
  font-size: var(--kore-text-small);
}

.planning__table th,
.planning__table td {
  border: 1px solid var(--kore-border);
  padding: 0.375rem 0.5rem;
  text-align: center;
}

.planning__user {
  text-align: left;
  white-space: nowrap;
  font-weight: 600;
  position: sticky;
  left: 0;
  background: var(--kore-bg);
  z-index: 1;
}

.planning__slot {
  display: inline-block;
  min-width: 2rem;
  padding: 0.125rem 0.25rem;
  border-radius: var(--kore-radius-sm);
  background: var(--kore-bg-muted);
  color: var(--kore-text);
}

.planning__empty-cell {
  color: var(--kore-text-muted);
}

.planning__empty {
  padding: var(--kore-space-lg);
  text-align: center;
  color: var(--kore-text-muted);
}

@media (max-width: 768px) {
  .planning__table { min-width: 28rem; font-size: 0.75rem; }
}
</style>
