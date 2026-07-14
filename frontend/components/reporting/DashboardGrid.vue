<template>
  <div class="dashboard-grid">
    <AppKpiCard
      v-for="item in items"
      :key="item.key"
      :icon="item.icon"
      :tone="item.tone"
      :value="item.value"
      :label="item.label"
      :hint="item.hint"
      :to="item.to"
      clickable
    />
  </div>
</template>

<script setup lang="ts">
export type DashboardGridItem = {
  key: string
  icon: string
  tone?: 'gold' | 'warn' | 'success' | 'blue'
  value: string | number
  label: string
  hint?: string
  to?: string
}

defineProps<{
  items: DashboardGridItem[]
}>()
</script>

<style scoped>
.dashboard-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: var(--kore-space-md);
}

@media (max-width: 768px) {
  .dashboard-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
