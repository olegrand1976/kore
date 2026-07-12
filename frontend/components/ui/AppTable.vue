<script setup lang="ts">
defineProps<{
  columns: { key: string; label: string }[]
  rows: Record<string, unknown>[]
  rowKey?: string
}>()
</script>

<template>
  <div class="app-table-wrap">
    <table class="app-table">
      <thead>
        <tr>
          <th v-for="col in columns" :key="col.key">{{ col.label }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(row, idx) in rows" :key="String(row[rowKey || 'id'] ?? idx)">
          <td v-for="col in columns" :key="col.key">
            <slot :name="`cell-${col.key}`" :row="row" :value="row[col.key]">
              {{ row[col.key] }}
            </slot>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.app-table-wrap {
  overflow-x: auto;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-lg);
}

.app-table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--kore-text-small);
}

.app-table th,
.app-table td {
  padding: 0.75rem 1rem;
  text-align: left;
  border-bottom: 1px solid var(--kore-border);
}

.app-table th {
  background: var(--kore-bg-subtle);
  color: var(--kore-text-muted);
  font-weight: 600;
  text-transform: uppercase;
  font-size: var(--kore-text-caption);
  letter-spacing: 0.04em;
}

.app-table tbody tr:last-child td {
  border-bottom: none;
}

.app-table tbody tr:hover {
  background: var(--kore-bg-subtle);
}
</style>
